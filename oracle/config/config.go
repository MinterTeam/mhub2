package config

import (
	"flag"

	"github.com/spf13/viper"
)

type MinterConfig struct {
	ApiAddr string `mapstructure:"api_addr"`
}

type CosmosConfig struct {
	Mnemonic string `mapstructure:"mnemonic"`
	GrpcAddr string `mapstructure:"grpc_addr"`
	RpcAddr  string `mapstructure:"rpc_addr"`
}

type EthereumConfig struct {
	GasPriceProviders []string `mapstructure:"gas_price_providers"`
}

type Config struct {
	Cosmos   CosmosConfig
	Minter   MinterConfig
	Ethereum EthereumConfig
}

func Get() *Config {
	cfg := &Config{}

	configPath := flag.String("config", "config.toml", "path to the configuration file")
	cosmosMnemonic := flag.String("cosmos-mnemonic", "", "")
	flag.Parse()

	v := viper.New()
	v.SetConfigFile(*configPath)

	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		panic(err)
	}

	if *cosmosMnemonic != "" {
		cfg.Cosmos.Mnemonic = *cosmosMnemonic
	}

	return cfg
}
