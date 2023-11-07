// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers

import (
	"context"
	"fmt"
	"math/big"
	"strings"

	ethereumABI "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/spectre-node/chains/evm/abi"
	"github.com/sygmaprotocol/spectre-node/chains/evm/listener/events"
	evmMessage "github.com/sygmaprotocol/spectre-node/chains/evm/message"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
)

type EventFetcher interface {
	FetchEventLogs(ctx context.Context, contractAddress common.Address, event string, startBlock *big.Int, endBlock *big.Int) ([]types.Log, error)
}

type StepProver interface {
	StepProof(blocknumber *big.Int) ([32]byte, error)
}

type DepositEventHandler struct {
	msgChan chan []*message.Message

	eventFetcher EventFetcher
	stepProver   StepProver

	domainID      uint8
	routerABI     ethereumABI.ABI
	routerAddress common.Address
}

func NewDepositEventHandler(
	msgChan chan []*message.Message,
	eventFetcher EventFetcher,
	stepProver StepProver,
	routerAddress common.Address,
	domainID uint8,
) *DepositEventHandler {
	routerABI, _ := ethereumABI.JSON(strings.NewReader(abi.RouterABI))
	return &DepositEventHandler{
		eventFetcher:  eventFetcher,
		stepProver:    stepProver,
		routerAddress: routerAddress,
		routerABI:     routerABI,
		msgChan:       msgChan,
		domainID:      domainID,
	}
}

// HandleEvents fetches deposit events and if deposits exists, submits a step message
// to be executed on the destination network
func (h *DepositEventHandler) HandleEvents(startBlock *big.Int, endBlock *big.Int) error {
	deposits, err := h.fetchDeposits(startBlock, endBlock)
	if err != nil {
		return fmt.Errorf("unable to fetch deposit events because of: %+v", err)
	}
	domainDeposits := make(map[uint8][]*events.Deposit)
	for _, d := range deposits {
		domainDeposits[d.DestinationDomainID] = append(domainDeposits[d.DestinationDomainID], d)
	}
	if len(domainDeposits) == 0 {
		return nil
	}

	proof, err := h.stepProver.StepProof(endBlock)
	if err != nil {
		return err
	}
	for _, deposits := range domainDeposits {
		h.msgChan <- []*message.Message{
			evmMessage.NewEvmStepMessage(
				h.domainID,
				deposits[0].DestinationDomainID,
				evmMessage.StepData{
					Proof: proof,
				},
			),
		}
	}
	return nil
}

func (h *DepositEventHandler) fetchDeposits(startBlock *big.Int, endBlock *big.Int) ([]*events.Deposit, error) {
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

func (h *DepositEventHandler) unpackDeposit(data []byte) (*events.Deposit, error) {
	var d events.Deposit
	err := h.routerABI.UnpackIntoInterface(&d, "Deposit", data)
	if err != nil {
		return &events.Deposit{}, err
	}

	return &d, nil
}
