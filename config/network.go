// The Licensed Work is (c) 2023 Sygma
// SPDX-License-Identifier: LGPL-3.0-only

package config

type BaseNetworkConfig struct {
	ID       uint8  `required:"true"`
	Endpoint string `required:"true"`
	Key      string `required:"true"`
}
