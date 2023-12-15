// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package prover

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	ssz "github.com/ferranbt/fastssz"
	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/spectre-node/chains/evm/message"
	consensus "github.com/umbracle/go-eth-consensus"
)

type StepArgs struct {
	Spec    Spec
	Pubkeys [512][48]byte
	Domain  phase0.Domain
	Update  *consensus.LightClientFinalityUpdateCapella
}

type RotateArgs struct {
	Spec   Spec
	Update *consensus.LightClientUpdateCapella
}

type ProverResponse struct {
	Accumulator [12]string `json:"accumulator"`
	Proof       []uint16   `json:"proof"`
	Commitment  string     `json:"committee_poseidon"`
}

type EvmProof[T any] struct {
	Proof []byte
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
	CallFor(ctx context.Context, reply interface{}, method string, args ...interface{}) error
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
func (p *Prover) StepProof(args *StepArgs) (*EvmProof[message.SyncStepInput], error) {
	log := log.With().Str("Caller", "Prover.StepProof()").Logger()
	log.Info().Msgf("Running step proof...")

	updateSzz, err := args.Update.MarshalSSZ()
	if err != nil {
		return nil, err
	}

	type stepArgs struct {
		Spec    Spec     `json:"spec"`
		Pubkeys []uint16 `json:"pubkeys"`
		Domain  []uint16 `json:"domain"`
		Update  []uint16 `json:"light_client_finality_update"`
	}
	var resp ProverResponse
	log.Info().Msgf("Calling RPC for Step proof...")
	err = p.proverClient.CallFor(context.Background(), &resp, "genEvmProof_SyncStepCompressed", stepArgs{
		Spec:    args.Spec,
		Pubkeys: ByteArrayToU16Array(p.pubkeysSSZ(args.Pubkeys)),
		Update:  ByteArrayToU16Array(updateSzz),
		Domain:  ByteArrayToU16Array(args.Domain[:]),
	})
	if err != nil {
		return nil, err
	}

	accumulator := [12]*big.Int{}
	for i, value := range resp.Accumulator {
		accumulator[i], _ = new(big.Int).SetString(value[2:], 16)
	}

	log.Info().Msgf("Generated step proof")

	finalizedHeaderRoot, err := args.Update.FinalizedHeader.Header.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	executionRoot, err := args.Update.FinalizedHeader.Execution.HashTreeRoot()
	if err != nil {
		return nil, err
	}
	proof := &EvmProof[message.SyncStepInput]{
		Proof: U16ArrayToByteArray(resp.Proof),
		Input: message.SyncStepInput{
			AttestedSlot:         args.Update.AttestedHeader.Header.Slot,
			FinalizedSlot:        args.Update.FinalizedHeader.Header.Slot,
			Participation:        uint64(CountSetBits(args.Update.SyncAggregate.SyncCommiteeBits)),
			FinalizedHeaderRoot:  finalizedHeaderRoot,
			ExecutionPayloadRoot: executionRoot,
			Accumulator:          accumulator,
		},
	}
	return proof, nil
}

// RotateProof generates the proof for the sync committee rotation for the period
func (p *Prover) RotateProof(args *RotateArgs) (*EvmProof[message.RotateInput], error) {
	log := log.With().Str("Caller", "Prover.RotateProof()").Logger()
	log.Info().Msgf("Running Rotate Proof...")

	syncCommiteeRoot, err := p.pubkeysRoot(args.Update.NextSyncCommittee.PubKeys)
	if err != nil {
		return nil, err
	}
	updateSzz, err := args.Update.MarshalSSZ()
	if err != nil {
		return nil, err
	}

	type rotateArgs struct {
		Update []uint16 `json:"light_client_update"`
		Spec   Spec     `json:"spec"`
	}
	var resp ProverResponse
	log.Info().Msgf("Calling RPC for Rotate proof...")
	err = p.proverClient.CallFor(context.Background(), &resp, "genEvmProof_CommitteeUpdateCompressed", rotateArgs{Update: ByteArrayToU16Array(updateSzz), Spec: args.Spec})
	if err != nil {
		return nil, err
	}

	log.Info().Msgf("Generated rotate proof")

	comm, _ := new(big.Int).SetString(resp.Commitment[2:], 16)
	if err != nil {
		return nil, err
	}

	accumulator := [12]*big.Int{}
	for i, value := range resp.Accumulator {
		accumulator[i], _ = new(big.Int).SetString(value[2:], 16)
	}

	// log.Info().Msgf("Sync CommitteeRoot Bytes: %b", syncCommiteeRoot)
	// log.Info().Msg(hex.EncodeToString(syncCommiteeRoot[:]))
	// log.Info().Msg("POSEIDON")
	// log.Info().Msg(comm.String())
	proof := &EvmProof[message.RotateInput]{
		Proof: U16ArrayToByteArray(resp.Proof),
		Input: message.RotateInput{
			SyncCommitteeSSZ:      syncCommiteeRoot,
			SyncCommitteePoseidon: comm,
			Accumulator:           accumulator,
		},
	}
	return proof, nil
}

func (p *Prover) StepArgs() (*StepArgs, error) {
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

func (p *Prover) RotateArgs(epoch uint64) (*RotateArgs, error) {
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

func (p *Prover) pubkeysSSZ(pubkeys [512][48]byte) []byte {
	var pubkeysSSZ []byte
	for _, pubkeys := range pubkeys {
		pubkeysSSZ = append(pubkeysSSZ, pubkeys[:]...)
	}
	return pubkeysSSZ
}

func (p *Prover) pubkeysRoot(pubkeys [512][48]byte) ([32]byte, error) {
	h := ssz.NewHasher()
	subIndx := h.Index()
	for _, key := range pubkeys {
		h.PutBytes(key[:])
	}
	h.Merkleize(subIndx)
	return h.HashRoot()
}

func padHexString(hexString string, desiredLength int) string {
	currentLength := len(hexString)
	if currentLength >= desiredLength {
		return hexString
	}

	// Calculate the number of leading zeroes needed
	leadingZeroes := strings.Repeat("0", desiredLength-currentLength)

	// Concatenate the leading zeroes with the original hex string
	paddedHexString := leadingZeroes + hexString

	return paddedHexString
}
