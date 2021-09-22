package config

import (
	"errors"
	"flag"
	"reflect"

	"github.com/MinterTeam/minter-go-sdk/v2/transaction"
	"github.com/mitchellh/mapstructure"

	"github.com/spf13/viper"
)

type MinterConfig struct {
	MultisigAddr     string              `mapstructure:"multisig_addr"`
	ChainID          transaction.ChainID `mapstructure:"chain"`
	ApiAddr          string              `mapstructure:"api_addr"`
	PrivateKey       string              `mapstructure:"private_key"`
	StartBlock       uint64              `mapstructure:"start_block"`
	StartEventNonce  uint64              `mapstructure:"start_event_nonce"`
	StartBatchNonce  uint64              `mapstructure:"start_batch_nonce"`
	StartValsetNonce uint64              `mapstructure:"start_valset_nonce"`
}

type CosmosConfig struct {
	Mnemonic string
	GrpcAddr string `mapstructure:"grpc_addr"`
	RpcAddr  string `mapstructure:"rpc_addr"`
}

type Config struct {
	Minter MinterConfig
	Cosmos CosmosConfig
}

var cfg *Config

func Get() *Config {
	if cfg != nil {
		return cfg
	}

	cosmosMnemonic := flag.String("cosmos-mnemonic", "", "")
	minterPrivateKey := flag.String("minter-private-key", "", "")
	minterMultisigAddr := flag.String("minter-multisig-addr", "", "")
	configPath := flag.String("config", "config.toml", "path to the configuration file")
	flag.Parse()

	v := viper.New()
	v.SetConfigFile(*configPath)

	if err := v.ReadInConfig(); err != nil {
		panic(err)
	}

	err := v.Unmarshal(&cfg, viper.DecodeHook(
		stringToChainHook(),
	))

	if *cosmosMnemonic != "" {
		cfg.Cosmos.Mnemonic = *cosmosMnemonic
	}

	if *minterPrivateKey != "" {
		cfg.Minter.PrivateKey = *minterPrivateKey
	}

	if *minterMultisigAddr != "" {
		cfg.Minter.MultisigAddr = *minterMultisigAddr
	}

	if err != nil {
		panic(err)
	}

	return cfg
}

func stringToChainHook() mapstructure.DecodeHookFunc {
	return func(
		f reflect.Type,
		t reflect.Type,
		data interface{},
	) (interface{}, error) {
		if f.Kind() != reflect.String {
			return data, nil
		}

		if t != reflect.TypeOf(transaction.ChainID(0)) {
			return data, nil
		}

		switch data.(string) {
		case "mainnet":
			return transaction.MainNetChainID, nil
		case "testnet":
			return transaction.TestNetChainID, nil
		default:
			return nil, errors.New("unknown minter chain")
		}
	}
}
