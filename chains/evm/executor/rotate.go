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

type RotateSubmitter interface {
	Rotate(rotateInput message.RotateInput, rotateProof [32]byte, stepInput message.SyncStepInput, stepProof [32]byte, opts transactor.TransactOptions) (*common.Hash, error)
}

type EVMRotateExecutor struct {
	domainID uint8

	rotateSubmitter RotateSubmitter
}

func NewEVMRotateExecutor(domainID uint8, rotateSubmitter RotateSubmitter) *EVMRotateExecutor {
	return &EVMRotateExecutor{
		rotateSubmitter: rotateSubmitter,
		domainID:        domainID,
	}
}

func (e *EVMRotateExecutor) Execute(props []*proposal.Proposal) error {
	rotateData := props[0].Data.(message.RotateData)
	hash, err := e.rotateSubmitter.Rotate(
		rotateData.RotateInput,
		rotateData.RotateProof,
		rotateData.StepInput,
		rotateData.StepProof,
		transactor.TransactOptions{})
	if err != nil {
		return err
	}

	log.Info().Uint8("domainID", e.domainID).Msgf("Sent EVM rotate with hash: %s", hash)
	return nil
}
