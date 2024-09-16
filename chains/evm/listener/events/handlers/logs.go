// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package handlers

import (
	"context"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
)

const (
	MAX_BLOCK_RANGE int64 = 1000
)

type EventFetcher interface {
	FetchEventLogs(ctx context.Context, contractAddress common.Address, event string, startBlock *big.Int, endBlock *big.Int) ([]types.Log, error)
}

// fetchLogs calls fetch event logs multiple times with a predefined block range to prevent
// rpc errors when the block range is too large
func fetchLogs(eventFetcher EventFetcher, startBlock, endBlock *big.Int, contract common.Address, eventSignature string) ([]types.Log, error) {
	allLogs := make([]types.Log, 0)
	for startBlock.Cmp(endBlock) < 0 {
		rangeEnd := new(big.Int).Add(startBlock, big.NewInt(MAX_BLOCK_RANGE))
		if rangeEnd.Cmp(endBlock) > 0 {
			rangeEnd = endBlock
		}

		logs, err := eventFetcher.FetchEventLogs(context.Background(), contract, eventSignature, startBlock, rangeEnd)
		if err != nil {
			return nil, err
		}
		allLogs = append(allLogs, logs...)
		startBlock = new(big.Int).Add(rangeEnd, big.NewInt(1))
	}

	return allLogs, nil
}
