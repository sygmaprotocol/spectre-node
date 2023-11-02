package prover_test

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/sygmaprotocol/spectre-node/chains/evm/prover"
)

type UtilsTestSuite struct {
	suite.Suite
}

func TestRunUtilsTestSuite(t *testing.T) {
	suite.Run(t, new(UtilsTestSuite))
}

func (s *UtilsTestSuite) Test_EpochFromTimestamp_ValidEpoch() {
	epoch := prover.EpochFromTimestamp(1606427091)

	s.Equal(epoch, uint64(232))
}
