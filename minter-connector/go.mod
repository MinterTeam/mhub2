module github.com/MinterTeam/mhub2/minter-connector

go 1.13

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4

replace github.com/MinterTeam/mhub2/module => ../../mhub2/module

require (
	github.com/MinterTeam/mhub2/module v0.0.0-20210417174508-bac3972b7846
	github.com/MinterTeam/minter-go-sdk/v2 v2.2.0-alpha1.0.20210312102425-6b1675c84520
	github.com/cosmos/cosmos-sdk v0.43.0
	github.com/cosmos/go-bip39 v1.0.0
	github.com/ethereum/go-ethereum v1.9.25
	github.com/mitchellh/mapstructure v1.4.1
	github.com/spf13/viper v1.8.0
	github.com/tendermint/tendermint v0.34.12
	google.golang.org/grpc v1.38.0
)
