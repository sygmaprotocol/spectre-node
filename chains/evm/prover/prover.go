package prover

import (
	"context"
	"fmt"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec/phase0"
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

type EvmProof struct {
	Proof        [32]byte
	PublicInputs [][]byte
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
func (p *Prover) StepProof() (*EvmProof, error) {
	args, err := p.stepArgs()
	if err != nil {
		return nil, err
	}

	var proof *EvmProof
	err = p.proverClient.Call("genEvmProofAndInstancesStepSyncCircuit", args, &proof)
	if err != nil {
		return nil, err
	}

	return proof, nil
}

// RotateProof generates the proof for the sync committee rotation for the period
func (p *Prover) RotateProof(epoch uint64) (*EvmProof, error) {
	args, err := p.rotateArgs(epoch)
	if err != nil {
		return nil, err
	}

	var proof *EvmProof
	err = p.proverClient.Call("genEvmProofAndInstancesRotationCircuit", args, &proof)
	if err != nil {
		return nil, err
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
