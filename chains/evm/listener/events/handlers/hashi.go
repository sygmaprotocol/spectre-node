// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers

import (
	"math/big"
	"strings"

	ethereumABI "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/sygmaprotocol/spectre-node/chains/evm/abi"
	"github.com/sygmaprotocol/spectre-node/chains/evm/listener/events"
)

type HashiDomainCollector struct {
	domainID     uint8
	yahoAddress  common.Address
	yahoABI      ethereumABI.ABI
	eventFetcher EventFetcher
	domains      []uint8
}

func NewHashiDomainCollector(
	domainID uint8,
	yahoAddress common.Address,
	domains []uint8,
) *HashiDomainCollector {
	abi, _ := ethereumABI.JSON(strings.NewReader(abi.YahoABI))
	return &HashiDomainCollector{
		domainID:    domainID,
		yahoAddress: yahoAddress,
		yahoABI:     abi,
		domains:     domains,
	}
}

func (h *HashiDomainCollector) CollectDomains(startBlock *big.Int, endBlock *big.Int) ([]uint8, error) {
	logs, err := fetchLogs(h.eventFetcher, startBlock, endBlock, h.yahoAddress, string(events.MessageDispatchedSig))
	if err != nil {
		return []uint8{}, nil
	}

	if len(logs) == 0 {
		return []uint8{}, nil
	}
	return h.domains, nil
}
