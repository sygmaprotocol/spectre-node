// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package contracts

import (
	"strings"

	ethereumABI "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/sygmaprotocol/spectre-node/chains/evm/abi"
	"github.com/sygmaprotocol/spectre-node/chains/evm/message"
	coreContracts "github.com/sygmaprotocol/sygma-core/chains/evm/contracts"
	"github.com/sygmaprotocol/sygma-core/chains/evm/transactor"
)

type Spectre struct {
	coreContracts.Contract
}

func NewSpectreContract(
	address common.Address,
	transactor transactor.Transactor,
) *Spectre {
	a, _ := ethereumABI.JSON(strings.NewReader(abi.SpectreABI))
	return &Spectre{
		Contract: coreContracts.NewContract(address, a, nil, nil, transactor),
	}
}

func (c *Spectre) Step(args message.SyncStepInput, poseidonCommitment [32]byte, opts transactor.TransactOptions) (*common.Hash, error) {
	return c.ExecuteTransaction(
		"step",
		opts,
		args, poseidonCommitment,
	)
}
