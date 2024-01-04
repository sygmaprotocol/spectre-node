// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers

import (
	"context"
	"math/big"
	"strings"

	"github.com/attestantio/go-eth2-client/api"
	apiv1 "github.com/attestantio/go-eth2-client/api/v1"
	"github.com/attestantio/go-eth2-client/spec"
	ethereumABI "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/spectre-node/chains/evm/abi"
	evmMessage "github.com/sygmaprotocol/spectre-node/chains/evm/message"
	"github.com/sygmaprotocol/spectre-node/chains/evm/prover"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
)

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

type DepositEventHandler struct {
	msgChan chan []*message.Message

	eventFetcher EventFetcher
	blockFetcher BlockFetcher
	prover       Prover

	domainID      uint8
	domains       []uint8
	blockInterval uint64
	routerABI     ethereumABI.ABI
	routerAddress common.Address
}

func NewDepositEventHandler(
	msgChan chan []*message.Message,
	eventFetcher EventFetcher,
	blockFetcher BlockFetcher,
	prover Prover,
	routerAddress common.Address,
	domainID uint8,
	domains []uint8,
	blockInterval uint64,
) *DepositEventHandler {
	routerABI, _ := ethereumABI.JSON(strings.NewReader(abi.RouterABI))
	return &DepositEventHandler{
		eventFetcher:  eventFetcher,
		blockFetcher:  blockFetcher,
		prover:        prover,
		domains:       domains,
		routerAddress: routerAddress,
		routerABI:     routerABI,
		msgChan:       msgChan,
		domainID:      domainID,
		blockInterval: blockInterval,
	}
}

// HandleEvents fetches deposit events and if deposits exists, submits a step message
// to be executed on the destination network
func (h *DepositEventHandler) HandleEvents(checkpoint *apiv1.Finality) error {
	args, err := h.prover.StepArgs()
	if err != nil {
		return err
	}

	proof, err := h.prover.StepProof(args)
	if err != nil {
		return err
	}
	for _, destDomain := range h.domains {
		if destDomain == h.domainID {
			continue
		}

		log.Debug().Uint8("domainID", h.domainID).Msgf("Sending step message to domain %d", destDomain)
		h.msgChan <- []*message.Message{
			evmMessage.NewEvmStepMessage(
				h.domainID,
				destDomain,
				evmMessage.StepData{
					Proof: proof.Proof,
					Args:  proof.Input,
				},
			),
		}
	}
	return nil
}
