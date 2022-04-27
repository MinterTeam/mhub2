package config

import (
	"flag"

	"github.com/spf13/viper"
)

type CosmosConfig struct {
	GrpcAddr string `mapstructure:"grpc_addr"`
	RpcAddr  string `mapstructure:"rpc_addr"`
}

type Config struct {
	Cosmos     CosmosConfig
	HoldersUrl string `mapstructure:"holders_url"`
	PricesUrl  string `mapstructure:"prices_url"`
}

func Get() (*Config, bool) {
	cfg := &Config{}

	configPath := flag.String("config", "config.toml", "path to the configuration file")
	testnet := flag.Bool("testnet", false, "")

	flag.Parse()

	v := viper.New()
	v.SetConfigFile(*configPath)

	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		panic(err)
	}

	return cfg, *testnet
}
