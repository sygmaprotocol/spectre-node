package contracts

import (
	"strings"

	ethereumABI "github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"

	"github.com/sygmaprotocol/spectre-node/chains/evm/abi"
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

type SyncStepInput struct {
	AttestedSlot         uint64
	FinalizedSlot        uint64
	Participation        uint64
	FinalizedHeaderRoot  [32]byte
	ExecutionPayloadRoot [32]byte
}

func (c *Spectre) Step(args SyncStepInput, poseidonCommitment [32]byte, opts transactor.TransactOptions) (*common.Hash, error) {
	return c.ExecuteTransaction(
		"step",
		opts,
		args, poseidonCommitment,
	)
}
