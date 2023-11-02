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
	"github.com/sygmaprotocol/spectre-node/chains/evm/prover"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
)

type EventFetcher interface {
	FetchEventLogs(ctx context.Context, contractAddress common.Address, event string, startBlock *big.Int, endBlock *big.Int) ([]types.Log, error)
	HeaderByHash(ctx context.Context, hash common.Hash) (*types.Header, error)
}

type StepProver interface {
	StepProof(blocknumber *big.Int) ([]byte, error)
}

type DepositEventHandler struct {
	msgChan chan []*message.Message

	eventFetcher EventFetcher
	stepProver   StepProver

	domainID      uint8
	routerABI     ethereumABI.ABI
	routerAddress common.Address

	// stores latest epoch for which we submitted a step proof per domain to
	// prevent submitting proofs to the same domain twice
	latestStepEpoch map[uint8]uint64
	steps           map[uint64][]byte
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
		eventFetcher:    eventFetcher,
		stepProver:      stepProver,
		routerAddress:   routerAddress,
		routerABI:       routerABI,
		msgChan:         msgChan,
		domainID:        domainID,
		latestStepEpoch: make(map[uint8]uint64),
		steps:           make(map[uint64][]byte),
	}
}

// HandleEvents fetches deposit events and if deposits exists, submits a step message
// to be executed on the destination network
func (h *DepositEventHandler) HandleEvents(startBlock *big.Int, endBlock *big.Int) error {
	deposits, timestamp, err := h.fetchDeposits(startBlock, endBlock)
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

	epoch := prover.EpochFromTimestamp(timestamp)
	proof, err := h.stepProver.StepProof(endBlock)
	if err != nil {
		return err
	}
	h.steps[epoch] = proof
	for _, deposits := range domainDeposits {
		h.latestStepEpoch[deposits[0].DestinationDomainID] = epoch
		h.msgChan <- []*message.Message{
			evmMessage.NewEvmStepMessage(
				h.domainID,
				deposits[0].DestinationDomainID,
				proof,
			),
		}
	}
	return nil
}

func (h *DepositEventHandler) fetchDeposits(startBlock *big.Int, endBlock *big.Int) ([]*events.Deposit, uint64, error) {
	logs, err := h.eventFetcher.FetchEventLogs(context.Background(), h.routerAddress, string(events.DepositSig), startBlock, endBlock)
	if err != nil {
		return nil, 0, err
	}

	deposits := make([]*events.Deposit, 0)
	if len(logs) == 0 {
		return deposits, 0, nil
	}

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

	header, err := h.eventFetcher.HeaderByHash(context.Background(), logs[0].BlockHash)
	if err != nil {
		return deposits, 0, err
	}

	return deposits, header.Time, nil
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
