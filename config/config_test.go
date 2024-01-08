// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package config_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/sygmaprotocol/spectre-node/config"
)

type ConfigTestSuite struct {
	suite.Suite
}

func TestRunConfigTestSuite(t *testing.T) {
	suite.Run(t, new(ConfigTestSuite))
}

func (c *ConfigTestSuite) TearDownTest() {
	os.Clearenv()
}

func (s *ConfigTestSuite) Test_LoadConfig_MissingField() {
	_, err := config.LoadConfig()

	s.NotNil(err)
}

func (s *ConfigTestSuite) Test_LoadConfig_DefaultValues() {
	os.Setenv("SPECTRE_PROVER_URL", "http://prover.com")
	os.Setenv("SPECTRE_DOMAINS", "1:evm,2:evm")

	c, err := config.LoadConfig()

	domains := make(map[uint8]string)
	domains[1] = "evm"
	domains[2] = "evm"
	s.Nil(err)
	s.Equal(c, &config.Config{
		Observability: &config.Observability{
			LogLevel:   "debug",
			LogFile:    "out.log",
			HealthPort: 9001,
		},
		Prover: &config.Prover{
			URL: "http://prover.com",
		},
		Domains: domains,
	})
}

func (s *ConfigTestSuite) Test_LoadEVMConfig_SuccessfulLoad() {
	os.Setenv("SPECTRE_OBSERVABILITY_LOG_LEVEL", "info")
	os.Setenv("SPECTRE_OBSERVABILITY_LOG_FILE", "out2.log")
	os.Setenv("SPECTRE_OBSERVABILITY_HEALTH_PORT", "9003")
	os.Setenv("SPECTRE_PROVER_URL", "http://prover.com")
	os.Setenv("SPECTRE_DOMAINS", "1:evm,2:evm")

	c, err := config.LoadConfig()

	domains := make(map[uint8]string)
	domains[1] = "evm"
	domains[2] = "evm"
	s.Nil(err)
	s.Equal(c, &config.Config{
		Observability: &config.Observability{
			LogLevel:   "info",
			LogFile:    "out2.log",
			HealthPort: 9003,
		},
		Prover: &config.Prover{
			URL: "http://prover.com",
		},
		Domains: domains,
	})
}
