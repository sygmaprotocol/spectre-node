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

func (c *EVMConfigTestSuite) TearDownTest() {
	os.Clearenv()
}

func (s *EVMConfigTestSuite) Test_LoadEVMConfig_MissingField() {
	os.Setenv("SPECTRE_DOMAINS_1_ID", "1")
	os.Setenv("SPECTRE_DOMAINS_1_ENDPOINT", "http://endpoint.com")
	os.Setenv("SPECTRE_DOMAINS_1_KEY", "key")
	os.Setenv("SPECTRE_DOMAINS_1_SPECTRE", "spectre")
	os.Setenv("SPECTRE_DOMAINS_2_ROUTER", "invalid")

	_, err := config.LoadEVMConfig(1)

	s.NotNil(err)
}

func (s *EVMConfigTestSuite) Test_LoadEVMConfig_SuccessfulLoad_DefaultValues() {
	os.Setenv("SPECTRE_DOMAINS_1_ID", "1")
	os.Setenv("SPECTRE_DOMAINS_1_ENDPOINT", "http://endpoint.com")
	os.Setenv("SPECTRE_DOMAINS_1_KEY", "key")
	os.Setenv("SPECTRE_DOMAINS_1_SPECTRE", "spectre")
	os.Setenv("SPECTRE_DOMAINS_1_ROUTER", "router")
	os.Setenv("SPECTRE_DOMAINS_1_BEACON_ENDPOINT", "endpoint")
	os.Setenv("SPECTRE_DOMAINS_2_ROUTER", "invalid")

	c, err := config.LoadEVMConfig(1)

	s.Nil(err)
	s.Equal(c, &config.EVMConfig{
		BaseNetworkConfig: baseConfig.BaseNetworkConfig{
			ID:       1,
			Key:      "key",
			Endpoint: "http://endpoint.com",
		},
		Router:                "router",
		Spectre:               "spectre",
		BlockInterval:         32,
		GasMultiplier:         1,
		GasIncreasePercentage: 15,
		MaxGasPrice:           500000000000,
		RetryInterval:         12,
		CommitteePeriodLength: 256,
		BeaconEndpoint:        "endpoint",
	})
}

func (s *EVMConfigTestSuite) Test_LoadEVMConfig_SuccessfulLoad() {
	os.Setenv("SPECTRE_DOMAINS_1_ID", "1")
	os.Setenv("SPECTRE_DOMAINS_1_ENDPOINT", "http://endpoint.com")
	os.Setenv("SPECTRE_DOMAINS_1_KEY", "key")
	os.Setenv("SPECTRE_DOMAINS_1_SPECTRE", "spectre")
	os.Setenv("SPECTRE_DOMAINS_1_ROUTER", "router")
	os.Setenv("SPECTRE_DOMAINS_1_BEACON_ENDPOINT", "endpoint")
	os.Setenv("SPECTRE_DOMAINS_1_MAX_GAS_PRICE", "1000")
	os.Setenv("SPECTRE_DOMAINS_1_BLOCK_INTERVAL", "10")
	os.Setenv("SPECTRE_DOMAINS_1_GAS_MULTIPLIER", "1")
	os.Setenv("SPECTRE_DOMAINS_1_GAS_INCREASE_PERCENTAGE", "20")
	os.Setenv("SPECTRE_DOMAINS_1_RETRY_INTERVAL", "30")
	os.Setenv("SPECTRE_DOMAINS_1_COMMITTEE_PERIOD_LENGTH", "128")
	os.Setenv("SPECTRE_DOMAINS_2_ROUTER", "invalid")

	c, err := config.LoadEVMConfig(1)

	s.Nil(err)
	s.Equal(c, &config.EVMConfig{
		BaseNetworkConfig: baseConfig.BaseNetworkConfig{
			ID:       1,
			Key:      "key",
			Endpoint: "http://endpoint.com",
		},
		Router:                "router",
		Spectre:               "spectre",
		BlockInterval:         10,
		GasMultiplier:         1,
		GasIncreasePercentage: 20,
		MaxGasPrice:           1000,
		RetryInterval:         30,
		CommitteePeriodLength: 128,
		BeaconEndpoint:        "endpoint",
	})
}
