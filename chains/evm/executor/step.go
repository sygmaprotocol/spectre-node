package executor

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/spectre-node/chains/evm/message"
	"github.com/sygmaprotocol/sygma-core/chains/evm/transactor"
	"github.com/sygmaprotocol/sygma-core/relayer/proposal"
)

type StepSubmitter interface {
	Step(args message.SyncStepInput, poseidonCommitment [32]byte, opts transactor.TransactOptions) (*common.Hash, error)
}

type EVMStepExecutor struct {
	stepSubmitter StepSubmitter
}

func NewEVMStepExecutor(stepSubmitter StepSubmitter) *EVMStepExecutor

func (e *EVMStepExecutor) Execute(props []*proposal.Proposal) error {
	stepData := props[0].Data.(message.StepData)
	hash, err := e.stepSubmitter.Step(stepData.Args, stepData.Proof, transactor.TransactOptions{})
	if err != nil {
		return err
	}

	log.Info().Msgf("Sent step with hash: %s", hash)
	return nil
}
