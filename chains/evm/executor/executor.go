// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package executor

import (
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/spectre-node/chains/evm/message"
	"github.com/sygmaprotocol/sygma-core/chains/evm/transactor"
	"github.com/sygmaprotocol/sygma-core/relayer/proposal"
)

type ProofSubmitter interface {
	Step(
		domainID uint8,
		input message.SyncStepInput,
		stepProof []byte,
		stateRoot [32]byte,
		stateRootProof [][]byte,
		opts transactor.TransactOptions,
	) (*common.Hash, error)
	Rotate(
		domainID uint8,
		rotateProof []byte,
		stepInput message.SyncStepInput,
		stepProof []byte,
		opts transactor.TransactOptions,
	) (*common.Hash, error)
}

type EVMExecutor struct {
	domainID uint8

	proofSubmitter ProofSubmitter
}

func NewEVMExecutor(domainID uint8, proofSubmitter ProofSubmitter) *EVMExecutor {
	return &EVMExecutor{
		proofSubmitter: proofSubmitter,
		domainID:       domainID,
	}
}

func (e *EVMExecutor) Execute(props []*proposal.Proposal) error {
	switch prop := props[0]; prop.Type {
	case message.EVMRotateProposal:
		rotateData := prop.Data.(message.RotateData)
		return e.rotate(prop.Source, rotateData)
	case message.EVMStepProposal:
		stepData := prop.Data.(message.StepData)
		return e.step(prop.Source, stepData)
	default:
		return fmt.Errorf("no executor configured for prop type %s", prop.Type)
	}
}

func (e *EVMExecutor) step(domainID uint8, stepData message.StepData) error {
	hash, err := e.proofSubmitter.Step(
		domainID,
		stepData.Args,
		stepData.Proof,
		stepData.StateRoot,
		stepData.StateRootProof,
		transactor.TransactOptions{})
	if err != nil {
		return err
	}

	log.Info().Uint8("domainID", e.domainID).Msgf("Sent EVM step with hash: %s", hash)
	return nil
}

func (e *EVMExecutor) rotate(domainID uint8, rotateData message.RotateData) error {
	hash, err := e.proofSubmitter.Rotate(
		domainID,
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
