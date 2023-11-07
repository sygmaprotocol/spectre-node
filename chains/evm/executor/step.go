// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package executor

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/spectre-node/chains/evm/message"
	"github.com/sygmaprotocol/sygma-core/chains/evm/transactor"
	"github.com/sygmaprotocol/sygma-core/relayer/proposal"
)

type StepSubmitter interface {
	Step(input message.SyncStepInput, stepProof [32]byte, opts transactor.TransactOptions) (*common.Hash, error)
}

type EVMStepExecutor struct {
	domainID uint8

	stepSubmitter StepSubmitter
}

func NewEVMStepExecutor(domainID uint8, stepSubmitter StepSubmitter) *EVMStepExecutor {
	return &EVMStepExecutor{
		stepSubmitter: stepSubmitter,
		domainID:      domainID,
	}
}

func (e *EVMStepExecutor) Execute(props []*proposal.Proposal) error {
	stepData := props[0].Data.(message.StepData)
	hash, err := e.stepSubmitter.Step(stepData.Args, stepData.Proof, transactor.TransactOptions{})
	if err != nil {
		return err
	}

	log.Info().Uint8("domainID", e.domainID).Msgf("Sent EVM step with hash: %s", hash)
	return nil
}
