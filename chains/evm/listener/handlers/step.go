// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	mapset "github.com/deckarep/golang-set/v2"
	ethereumABI "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/spectre-node/chains/evm/abi"
	"github.com/sygmaprotocol/spectre-node/chains/evm/listener/events"
	evmMessage "github.com/sygmaprotocol/spectre-node/chains/evm/message"
	"github.com/sygmaprotocol/spectre-node/chains/evm/prover"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
)

const EXECUTION_STATE_ROOT_INDEX = 34

type EventFetcher interface {
	FetchEventLogs(ctx context.Context, contractAddress common.Address, event string, startBlock *big.Int, endBlock *big.Int) ([]types.Log, error)
}

type Prover interface {
	StepProof(args *prover.StepArgs) (*prover.EvmProof[evmMessage.SyncStepInput], error)
	RotateProof(args *prover.RotateArgs) (*prover.EvmProof[evmMessage.RotateInput], error)
	StepArgs() (*prover.StepArgs, error)
	RotateArgs(epoch uint64) (*prover.RotateArgs, error)
}

type BlockFetcher interface {
	SignedBeaconBlock(ctx context.Context, opts *api.SignedBeaconBlockOpts) (*api.Response[*spec.VersionedSignedBeaconBlock], error)
}

type StepEventHandler struct {
	msgChan chan []*message.Message

	eventFetcher EventFetcher
	blockFetcher BlockFetcher
	prover       Prover

	domainID      uint8
	allDomains    []uint8
	routerABI     ethereumABI.ABI
	routerAddress common.Address

	latestBlock uint64
}

func NewStepEventHandler(
	msgChan chan []*message.Message,
	eventFetcher EventFetcher,
	blockFetcher BlockFetcher,
	prover Prover,
	routerAddress common.Address,
	domainID uint8,
	domains []uint8,
) *StepEventHandler {
	routerABI, _ := ethereumABI.JSON(strings.NewReader(abi.RouterABI))
	return &StepEventHandler{
		eventFetcher:  eventFetcher,
		blockFetcher:  blockFetcher,
		prover:        prover,
		routerAddress: routerAddress,
		routerABI:     routerABI,
		msgChan:       msgChan,
		domainID:      domainID,
		allDomains:    domains,
		latestBlock:   0,
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

	endBlock := block.Data.Capella.Message.Body.ExecutionPayload.BlockNumber
	if h.latestBlock == 0 {
		return h.allDomains, endBlock, nil
	}

	deposits, err := h.fetchDeposits(big.NewInt(int64(h.latestBlock)), big.NewInt(int64(endBlock)))
	if err != nil {
		return domains.ToSlice(), endBlock, err
	}
	if len(deposits) == 0 {
		return domains.ToSlice(), endBlock, nil
	}
	for _, deposit := range deposits {
		domains.Add(deposit.DestinationDomainID)
	}

	return domains.ToSlice(), endBlock, nil
}

func (h *StepEventHandler) fetchDeposits(startBlock *big.Int, endBlock *big.Int) ([]*events.Deposit, error) {
	logs, err := h.eventFetcher.FetchEventLogs(context.Background(), h.routerAddress, string(events.DepositSig), startBlock, endBlock)
	if err != nil {
		return nil, err
	}

	deposits := make([]*events.Deposit, 0)
	for _, dl := range logs {
		d, err := h.unpackDeposit(dl.Data)
		if err != nil {
			log.Error().Msgf("Failed unpacking deposit event log: %v", err)
			continue
		}
		d.SenderAddress = common.BytesToAddress(dl.Topics[1].Bytes())

		log.Debug().Msgf("Found deposit log in block: %d, TxHash: %s, contractAddress: %s, sender: %s", dl.BlockNumber, dl.TxHash, dl.Address, d.SenderAddress)
		deposits = append(deposits, d)
	}

	return deposits, nil
}

func (h *StepEventHandler) unpackDeposit(data []byte) (*events.Deposit, error) {
	var d events.Deposit
	err := h.routerABI.UnpackIntoInterface(&d, "Deposit", data)
	if err != nil {
		return &events.Deposit{}, err
	}

	return &d, nil
}
