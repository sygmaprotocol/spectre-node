// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/sygmaprotocol/spectre-node/chains/evm/config"
	baseConfig "github.com/sygmaprotocol/spectre-node/config"
)

type EVMConfigTestSuite struct {
	suite.Suite
}

func TestRunEVMConfigTestSuite(t *testing.T) {
	suite.Run(t, new(EVMConfigTestSuite))
}

func (s *EVMConfigTestSuite) Test_LoadEVMConfig_MissingField() {
	os.Setenv("SPECTRE_DOMAINS_1_ID", "1")
	os.Setenv("SPECTRE_DOMAINS_1_ENDPOINT", "http://endpoint.com")
	os.Setenv("SPECTRE_DOMAINS_1_KEY", "key")
	os.Setenv("SPECTRE_DOMAINS_1_EXECUTOR", "executor")
	os.Setenv("SPECTRE_DOMAINS_2_ROUTER", "invalid")

	_, err := config.LoadEVMConfig(1)

	s.NotNil(err)
}

func (s *EVMConfigTestSuite) Test_LoadEVMConfig_SuccessfulLoad() {
	os.Setenv("SPECTRE_DOMAINS_1_ID", "1")
	os.Setenv("SPECTRE_DOMAINS_1_ENDPOINT", "http://endpoint.com")
	os.Setenv("SPECTRE_DOMAINS_1_KEY", "key")
	os.Setenv("SPECTRE_DOMAINS_1_EXECUTOR", "executor")
	os.Setenv("SPECTRE_DOMAINS_1_ROUTER", "router")
	os.Setenv("SPECTRE_DOMAINS_2_ROUTER", "invalid")

	c, err := config.LoadEVMConfig(1)

	s.Nil(err)
	s.Equal(c, &config.EVMConfig{
		BaseNetworkConfig: baseConfig.BaseNetworkConfig{
			ID:       1,
			Key:      "key",
			Endpoint: "http://endpoint.com",
		},
		Router:   "router",
		Executor: "executor",
	})
}
