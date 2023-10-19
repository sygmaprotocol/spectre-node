package config

import (
	"fmt"

	"github.com/kelseyhightower/envconfig"
	"github.com/sygmaprotocol/spectre-node/config"
)

type EVMConfig struct {
	config.BaseNetworkConfig
	Router             string `required:"true"`
	Executor           string `required:"true"`
	BlockConfirmations uint8  `default:"5"`
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
