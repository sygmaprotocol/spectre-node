package prover

import (
	"context"
	"fmt"

	"github.com/attestantio/go-eth2-client/api"
	"github.com/attestantio/go-eth2-client/spec/phase0"
	consensus "github.com/umbracle/go-eth-consensus"
)

type StepArgs struct {
	Pubkeys [512][48]byte
	Domain  phase0.Domain
	Update  *consensus.LightClientFinalityUpdateCapella
}

type RotateArgs struct {
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

type Prover struct {
	lightClient  LightClient
	beaconClient BeaconClient

	epochSize     uint64
	committeeSize uint64
}

func NewProver(beaconClient BeaconClient, lightClient LightClient, epochSize uint64, committeeSize uint64) *Prover {
	return &Prover{
		epochSize:     epochSize,
		committeeSize: committeeSize,
		beaconClient:  beaconClient,
		lightClient:   lightClient,
	}
}

// StepProof generates the proof for the sync step
func (p *Prover) StepProof() (*EvmProof, error) {
	args, err := p.stepArgs()
	if err != nil {
		return nil, err
	}
	fmt.Println(args)

	return nil, nil
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

	domain, err := p.beaconClient.Domain(context.Background(), phase0.DomainType{}, phase0.Epoch(update.FinalizedHeader.Header.Slot/32))
	if err != nil {
		return nil, err
	}

	return &StepArgs{
		Pubkeys: pubkeys,
		Domain:  domain,
		Update:  update,
	}, nil
}

func (p *Prover) RotateProof(slot uint64) (*EvmProof, error) {
	args, err := p.rotateArgs(slot)
	if err != nil {
		return nil, err
	}
	fmt.Println(args)

	return nil, nil
}

func (p *Prover) rotateArgs(slot uint64) (*RotateArgs, error) {
	period := slot / p.epochSize / p.committeeSize
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
	}, nil
}
