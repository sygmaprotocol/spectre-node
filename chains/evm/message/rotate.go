// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package message

import (
	"math/big"

	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
	"github.com/sygmaprotocol/sygma-core/relayer/proposal"
)

const (
	EVMRotateMessage  message.MessageType   = "EVMRotateMessage"
	EVMRotateProposal proposal.ProposalType = "EVMRotateProposal"
)

type RotateInput struct {
	SyncCommitteeSSZ      [32]byte
	SyncCommitteePoseidon *big.Int
	Accumulator           [12]*big.Int
}

type RotateData struct {
	RotateProof []byte
	RotateInput RotateInput
	StepProof   []byte
	StepInput   SyncStepInput
}

func NewEvmRotateMessage(source uint8, destination uint8, rotateData RotateData) *message.Message {
	return &message.Message{
		Source:      source,
		Destination: destination,
		Data:        rotateData,
		Type:        EVMRotateMessage,
	}
}

type EvmRotateHandler struct{}

func (h *EvmRotateHandler) HandleMessage(m *message.Message) (*proposal.Proposal, error) {
	log.Debug().Uint8("domainID", m.Destination).Msgf("Received rotate message from domain %d", m.Source)

	return &proposal.Proposal{
		Source:      m.Source,
		Destination: m.Destination,
		Data:        m.Data,
		Type:        EVMRotateProposal,
	}, nil
}
