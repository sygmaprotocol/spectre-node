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
	"github.com/sygmaprotocol/sygma-core/relayer/message"
)

type EventFetcher interface {
	FetchEventLogs(ctx context.Context, contractAddress common.Address, event string, startBlock *big.Int, endBlock *big.Int) ([]types.Log, error)
}

type DepositEventHandler struct {
	msgChan chan []*message.Message

	eventFetcher EventFetcher

	routerABI     ethereumABI.ABI
	routerAddress common.Address
}

func NewDepositEventHandler(msgChan chan []*message.Message, eventFetcher EventFetcher, routerAddress common.Address) *DepositEventHandler {
	routerABI, _ := ethereumABI.JSON(strings.NewReader(abi.RouterABI))
	return &DepositEventHandler{
		eventFetcher:  eventFetcher,
		routerAddress: routerAddress,
		routerABI:     routerABI,
		msgChan:       msgChan,
	}
}

// HandleEvents fetches deposit events and if deposits exists, submits a block root message
// to the destination network of the deposit
func (h *DepositEventHandler) HandleEvents(startBlock *big.Int, endBlock *big.Int) error {
	deposits, err := h.fetchDeposits(startBlock, endBlock)
	if err != nil {
		return fmt.Errorf("unable to fetch deposit events because of: %+v", err)
	}
	if len(deposits) == 0 {
		return nil
	}

	h.msgChan <- []*message.Message{
		{},
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
	fmt.Println(data)

	var d events.Deposit
	err := h.routerABI.UnpackIntoInterface(&d, "Deposit", data)
	if err != nil {
		return &events.Deposit{}, err
	}

	return &d, nil
}
