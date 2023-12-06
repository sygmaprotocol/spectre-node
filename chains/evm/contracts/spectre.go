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

func (c *Spectre) Step(stepInput message.SyncStepInput, stepProof []byte, opts transactor.TransactOptions) (*common.Hash, error) {
	return c.ExecuteTransaction(
		"step",
		opts,
		stepInput, stepProof,
	)
}

func (c *Spectre) Rotate(rotateInput message.RotateInput, rotateProof []byte, stepInput message.SyncStepInput, stepProof []byte, opts transactor.TransactOptions) (*common.Hash, error) {
	type ContractRotateInput struct {
		SyncCommitteeSSZ      [32]byte
		SyncCommitteePoseidon [32]byte
	}
	return c.ExecuteTransaction(
		"rotate",
		opts,
		ContractRotateInput{
			SyncCommitteeSSZ:      rotateInput.SyncCommitteeSSZ,
			SyncCommitteePoseidon: rotateInput.SyncCommitteePoseidon,
		}, rotateProof, stepInput, stepProof, rotateInput.Accumulator,
	)
}
