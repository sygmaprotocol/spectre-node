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
	BeaconEndpoint        string `split_words:"true"`
	Router                string
	Spectre               string
	Yaho                  string
	Spec                  string  `default:"mainnet"`
	MaxGasPrice           int64   `default:"500000000000" split_words:"true"`
	GasMultiplier         float64 `default:"1" split_words:"true"`
	GasIncreasePercentage int64   `default:"15" split_words:"true"`
	RetryInterval         uint64  `default:"12" split_words:"true"`
	CommitteePeriodLength uint64  `default:"256" split_words:"true"`
	StartingPeriod        uint64  `required:"true" split_words:"true"`
	ForcePeriod           bool    `default:"false" split_words:"true"`
	FinalityThreshold     uint64  `default:"342" split_words:"true"`
	SlotsPerEpoch         uint64  `default:"32" split_words:"true"`
	TargetDomains         []int16 `split_words:"true"`
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
