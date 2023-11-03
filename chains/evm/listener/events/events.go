// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package events

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

type EventSig string

func (es EventSig) GetTopic() common.Hash {
	return crypto.Keccak256Hash([]byte(es))
}

const (
	DepositSig EventSig = "Deposit(uint8,uint8,bytes32,uint64,address,bytes)"
)

// Deposit struct holds event data raised by Deposit event on-chain
type Deposit struct {
	// ID of chain deposit will be bridged to
	DestinationDomainID uint8
	// SecurityModel is used to distringuish between block header oracles
	// on the destination network that verify this deposit
	SecurityModel uint8
	// ResourceID used to find address of handler to be used for deposit
	ResourceID [32]byte
	// Nonce of deposit
	DepositNonce uint64
	// Address of sender (msg.sender: user)
	SenderAddress common.Address
	// Additional data to be passed to specified handler
	Data []byte
}
