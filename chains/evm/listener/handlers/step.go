// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers

import (
	"context"
	"fmt"
	"math/big"

	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	mapset "github.com/deckarep/golang-set/v2"
	"github.com/rs/zerolog/log"
	evmMessage "github.com/sygmaprotocol/spectre-node/chains/evm/message"
	"github.com/sygmaprotocol/spectre-node/chains/evm/prover"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
)

const EXECUTION_STATE_ROOT_INDEX = 34

type Prover interface {
	StepProof(args *prover.StepArgs) (*prover.EvmProof[evmMessage.SyncStepInput], error)
	RotateProof(args *prover.RotateArgs) (*prover.EvmProof[struct{}], error)
	StepArgs() (*prover.StepArgs, error)
	RotateArgs(epoch uint64) (*prover.RotateArgs, error)
}

type BlockFetcher interface {
	SignedBeaconBlock(ctx context.Context, opts *api.SignedBeaconBlockOpts) (*api.Response[*spec.VersionedSignedBeaconBlock], error)
}

type DomainCollector interface {
	CollectDomains(startBlock *big.Int, endBlock *big.Int) ([]uint8, error)
}

type StepEventHandler struct {
	msgChan chan []*message.Message

	blockFetcher     BlockFetcher
	domainCollectors []DomainCollector
	prover           Prover

	domainID uint8
	domains  []uint8

	latestBlock uint64
}

func NewStepEventHandler(
	msgChan chan []*message.Message,
	domainCollectors []DomainCollector,
	blockFetcher BlockFetcher,
	prover Prover,
	domainID uint8,
	domains []uint8,
) *StepEventHandler {
	return &StepEventHandler{
		blockFetcher:     blockFetcher,
		prover:           prover,
		domainCollectors: domainCollectors,
		msgChan:          msgChan,
		domainID:         domainID,
		domains:          domains,
		latestBlock:      0,
	}
}

// HandleEvents executes the step for the latest finality checkpoint
func (h *StepEventHandler) HandleEvents(checkpoint *apiv1.Finality) error {
	args, err := h.prover.StepArgs()
	if err != nil {
		return err
	}
	domains, latestBlock, err := h.destinationDomains(args.Update.FinalizedHeader.Header.Slot)
	if err != nil {
		return err
	}
	if len(domains) == 0 {
		h.latestBlock = latestBlock
		log.Debug().Uint8("domainID", h.domainID).Uint64("slot", args.Update.FinalizedHeader.Header.Slot).Msgf("Skipping step...")
		return nil
	}

	log.Info().Uint8("domainID", h.domainID).Uint64("slot", args.Update.FinalizedHeader.Header.Slot).Msgf("Executing sync step")

	proof, err := h.prover.StepProof(args)
	if err != nil {
		return err
	}
	node, err := args.Update.FinalizedHeader.Execution.GetTree()
	if err != nil {
		return err
	}
	stateRootProof, err := node.Prove(EXECUTION_STATE_ROOT_INDEX)
	if err != nil {
		return err
	}

	for _, destDomain := range domains {
		if destDomain == h.domainID {
			continue
		}

		log.Debug().Uint8("domainID", h.domainID).Msgf("Sending step message to domain %d", destDomain)
		h.msgChan <- []*message.Message{
			evmMessage.NewEvmStepMessage(
				h.domainID,
				destDomain,
				evmMessage.StepData{
					Proof:          proof.Proof,
					Args:           proof.Input,
					StateRoot:      args.Update.FinalizedHeader.Execution.StateRoot,
					StateRootProof: stateRootProof.Hashes,
				},
			),
		}
	}
	h.latestBlock = latestBlock
	return nil
}

func (h *StepEventHandler) destinationDomains(slot uint64) ([]uint8, uint64, error) {
	domains := mapset.NewSet[uint8]()
	block, err := h.blockFetcher.SignedBeaconBlock(context.Background(), &api.SignedBeaconBlockOpts{
		Block: fmt.Sprint(slot),
	})
	if err != nil {
		return domains.ToSlice(), 0, err
	}
	endBlock := block.Data.Deneb.Message.Body.ExecutionPayload.BlockNumber
	if h.latestBlock == 0 {
		return h.domains, endBlock, nil
	}

	for _, collector := range h.domainCollectors {
		collectedDomains, err := collector.CollectDomains(new(big.Int).SetUint64(h.latestBlock), new(big.Int).SetUint64(endBlock))
		if err != nil {
			return domains.ToSlice(), 0, err
		}
		domains.Append(collectedDomains...)
	}
	return domains.ToSlice(), endBlock, nil
}
