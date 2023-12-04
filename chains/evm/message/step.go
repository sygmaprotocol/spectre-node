// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package message

import (
	"github.com/rs/zerolog/log"
	"github.com/sygmaprotocol/sygma-core/relayer/message"
	"github.com/sygmaprotocol/sygma-core/relayer/proposal"
)

const (
	EVMStepMessage  message.MessageType   = "EVMStepMessage"
	EVMStepProposal proposal.ProposalType = "EVMStepProposal"
)

type SyncStepInput struct {
	AttestedSlot         uint64
	FinalizedSlot        uint64
	Participation        uint64
	FinalizedHeaderRoot  [32]byte
	ExecutionPayloadRoot [32]byte
}

type StepData struct {
	Proof [32]byte
	Args  SyncStepInput
}

func NewEvmStepMessage(source uint8, destination uint8, stepData StepData) *message.Message {
	return &message.Message{
		Source:      source,
		Destination: destination,
		Data:        stepData,
		Type:        EVMStepMessage,
	}
}

type EvmStepHandler struct{}

func (h *EvmStepHandler) HandleMessage(m *message.Message) (*proposal.Proposal, error) {
	log.Debug().Uint8("domainID", m.Destination).Msgf("Received step message from domain %d", m.Source)

	return &proposal.Proposal{
		Source:      m.Source,
		Destination: m.Destination,
		Data:        m.Data,
		Type:        EVMStepProposal,
	}, nil
}
