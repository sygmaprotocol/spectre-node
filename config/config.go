// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package config

import "github.com/kelseyhightower/envconfig"

const PREFIX = "SPECTRE"

type Config struct {
	Observability
}

type Observability struct {
	LogLevel string `default:"debug" split_words:"true"`
	LogFile  string `default:"out.log" split_words:"true"`
}

// LoadConfig loads config from the environment and validates the fields
func LoadConfig() (*Config, error) {
	var c Config
	err := envconfig.Process(PREFIX, &c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}
