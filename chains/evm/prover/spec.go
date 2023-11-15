package prover

import "github.com/attestantio/go-eth2-client/spec/phase0"

type Spec string

const (
	TESTNET_SPEC Spec = "testnet"
	MINIMAL_SPEC Spec = "minimal"
	MAINNET_SPEC Spec = "mainnet"
)

var (
	SYNC_COMMITTEE_DOMAIN phase0.DomainType = [4]byte{7, 0, 0, 0}
)
