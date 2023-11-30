// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package prover

import "github.com/attestantio/go-eth2-client/spec/phase0"

type Spec string

const (
	TESTNET_SPEC Spec = "Testnet"
	MINIMAL_SPEC Spec = "Minimal"
	MAINNET_SPEC Spec = "Mainnet"
)

var (
	SYNC_COMMITTEE_DOMAIN phase0.DomainType = [4]byte{7, 0, 0, 0}
)
