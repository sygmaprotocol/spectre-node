// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/sygmaprotocol/spectre-node/config"
)

type EVMConfig struct {
	config.BaseNetworkConfig
	BeaconEndpoint        string  `required:"true"`
	Router                string  `required:"true"`
	Executor              string  `required:"true"`
	MaxGasPrice           int64   `default:"500000000000"`
	GasMultiplier         float64 `default:"1"`
	GasIncreasePercentage int64   `default:"15"`
	RetryInterval         uint64  `default:"12"`
}

// LoadEVMConfig loads EVM config from the environment and validates the fields
func LoadEVMConfig(domainID uint8) (*EVMConfig, error) {
	var c EVMConfig
	err := envconfig.Process(fmt.Sprintf("%s_DOMAINS_%d", config.PREFIX, domainID), &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
