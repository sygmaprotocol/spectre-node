package config

type BaseNetworkConfig struct {
	ID       uint8  `required:"true"`
	Endpoint string `required:"true"`
	Key      string `required:"true"`
}
