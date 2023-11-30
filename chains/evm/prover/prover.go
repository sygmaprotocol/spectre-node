// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package prover

import (
	"context"
	"fmt"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	ssz "github.com/ferranbt/fastssz"
	"github.com/sygmaprotocol/spectre-node/chains/evm/message"
	consensus "github.com/umbracle/go-eth-consensus"
)

type StepArgs struct {
	Spec    Spec                                        `json:"spec"`
	Pubkeys [512][48]byte                               `json:"pubkeys"`
	Domain  phase0.Domain                               `json:"domain"`
	Update  *consensus.LightClientFinalityUpdateCapella `json:"light_client_finality_update"`
}

type RotateArgs struct {
	Spec   Spec                                `json:"spec"`
	Update *consensus.LightClientUpdateCapella `json:"light_client_update"`
}

type ProverResponse struct {
	Proof        [32]byte `json:"proof"`
	PublicInputs [][]byte `json:"public_inputs"`
}

type CommitmentResponse struct {
	Commitment [32]byte `json:"commitment"`
}

type CommitmentArgs struct {
	Pubkeys [512][48]byte `json:"pubkeys"`
}

type EvmProof[T any] struct {
	Proof [32]byte
	Input T
}

type LightClient interface {
	FinalityUpdate() (*consensus.LightClientFinalityUpdateCapella, error)
	Updates(period uint64) ([]*consensus.LightClientUpdateCapella, error)
	Bootstrap(blockRoot string) (*consensus.LightClientBootstrapCapella, error)
}

type BeaconClient interface {
	BeaconBlockRoot(ctx context.Context, opts *api.BeaconBlockRootOpts) (*api.Response[*phase0.Root], error)
	Domain(ctx context.Context, domainType phase0.DomainType, epoch phase0.Epoch) (phase0.Domain, error)
}

type ProverClient interface {
	Call(serviceMethod string, args any, reply any) error
}

type Prover struct {
	lightClient  LightClient
	beaconClient BeaconClient
	proverClient ProverClient

	spec                  Spec
	committeePeriodLength uint64
}

func NewProver(
	proverClient ProverClient,
	beaconClient BeaconClient,
	lightClient LightClient,
	spec Spec,
	committeePeriodLength uint64,
) *Prover {
	return &Prover{
		proverClient:          proverClient,
		spec:                  spec,
		committeePeriodLength: committeePeriodLength,
		beaconClient:          beaconClient,
		lightClient:           lightClient,
	}
}

// StepProof generates the proof for the sync step
func (p *Prover) StepProof() (*EvmProof[message.SyncStepInput], error) {
	args, err := p.stepArgs()
	if err != nil {
		return nil, err
	}

	var resp ProverResponse
	err = p.proverClient.Call("genEvmProofAndInstancesStepSyncCircuit", args, &resp)
	if err != nil {
		return nil, err
	}

	finalizedHeaderRoot, err := args.Update.FinalizedHeader.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	executionRoot, err := args.Update.FinalizedHeader.Execution.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	participation := uint64(0)
	for _, byte := range args.Update.SyncAggregate.SyncCommiteeBits {
		participation += uint64(byte)
	}
	proof := &EvmProof[message.SyncStepInput]{
		Proof: resp.Proof,
		Input: message.SyncStepInput{
			AttestedSlot:         args.Update.AttestedHeader.Header.Slot,
			FinalizedSlot:        args.Update.FinalizedHeader.Header.Slot,
			Participation:        participation,
			FinalizedHeaderRoot:  finalizedHeaderRoot,
			ExecutionPayloadRoot: executionRoot,
		},
	}
	return proof, nil
}

// RotateProof generates the proof for the sync committee rotation for the period
func (p *Prover) RotateProof(epoch uint64) (*EvmProof[message.RotateInput], error) {
	args, err := p.rotateArgs(epoch)
	if err != nil {
		return nil, err
	}

	var resp ProverResponse
	err = p.proverClient.Call("genEvmProofAndInstancesRotationCircuit", args, &resp)
	if err != nil {
		return nil, err
	}

	syncCommiteeRoot, err := p.committeeKeysRoot(args.Update.NextSyncCommittee.PubKeys)
	if err != nil {
		return nil, err
	}

	var commitmentResp CommitmentResponse
	err = p.proverClient.Call("syncCommitteePoseidonCompressed", CommitmentArgs{Pubkeys: args.Update.NextSyncCommittee.PubKeys}, &commitmentResp)
	if err != nil {
		return nil, err
	}

	proof := &EvmProof[message.RotateInput]{
		Proof: resp.Proof,
		Input: message.RotateInput{
			SyncCommitteeSSZ:      syncCommiteeRoot,
			SyncCommitteePoseidon: commitmentResp.Commitment,
		},
	}
	return proof, nil
}

func (p *Prover) stepArgs() (*StepArgs, error) {
	update, err := p.lightClient.FinalityUpdate()
	if err != nil {
		return nil, err
	}
	blockRoot, err := p.beaconClient.BeaconBlockRoot(context.Background(), &api.BeaconBlockRootOpts{
		Block: fmt.Sprint(update.FinalizedHeader.Header.Slot),
	})
	if err != nil {
		return nil, err
	}
	bootstrap, err := p.lightClient.Bootstrap(blockRoot.Data.String())
	if err != nil {
		return nil, err
	}
	pubkeys := bootstrap.CurrentSyncCommittee.PubKeys

	domain, err := p.beaconClient.Domain(context.Background(), SYNC_COMMITTEE_DOMAIN, phase0.Epoch(update.FinalizedHeader.Header.Slot/32))
	if err != nil {
		return nil, err
	}

	return &StepArgs{
		Pubkeys: pubkeys,
		Domain:  domain,
		Update:  update,
		Spec:    p.spec,
	}, nil
}

func (p *Prover) rotateArgs(epoch uint64) (*RotateArgs, error) {
	period := epoch / p.committeePeriodLength
	updates, err := p.lightClient.Updates(period)
	if err != nil {
		return nil, err
	}
	if len(updates) == 0 {
		return nil, fmt.Errorf("missing light client updates")
	}
	update := updates[0]
	return &RotateArgs{
		Update: update,
		Spec:   p.spec,
	}, nil
}

func (p *Prover) committeeKeysRoot(pubkeys [512][48]byte) ([32]byte, error) {
	keysSSZ := make([]byte, 0)
	for i := 0; i < 512; i++ {
		keysSSZ = append(keysSSZ, pubkeys[i][:]...)
	}

	hh := ssz.NewHasher()
	hh.PutBytes(keysSSZ)
	return hh.HashRoot()
}
