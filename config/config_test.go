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

func (s *ConfigTestSuite) Test_LoadConfig_DefaultValues() {
	c, err := config.LoadConfig()

	s.Nil(err)
	s.Equal(c, &config.Config{
		Observability: config.Observability{
			LogLevel: "debug",
			LogFile:  "out.log",
		},
	})
}

func (s *ConfigTestSuite) Test_LoadEVMConfig_SuccessfulLoad() {
	os.Setenv("SPECTRE_LOG_LEVEL", "info")
	os.Setenv("SPECTRE_LOG_FILE", "out2.log")

	c, err := config.LoadConfig()

	s.Nil(err)
	s.Equal(c, &config.Config{
		Observability: config.Observability{
			LogLevel: "info",
			LogFile:  "out2.log",
		},
	})
}
