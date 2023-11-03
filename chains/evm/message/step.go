// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package message

import "github.com/sygmaprotocol/sygma-core/relayer/message"

const (
	EVMStepMessage message.MessageType = "EVMStepMessage"
)

func NewEvmStepMessage(source uint8, destination uint8, stepProof interface{}) *message.Message {
	return &message.Message{
		Source:      source,
		Destination: destination,
		Data:        stepProof,
		Type:        EVMStepMessage,
	}
}
