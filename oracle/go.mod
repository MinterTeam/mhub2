module github.com/MinterTeam/mhub2/oracle

go 1.13

replace github.com/gogo/protobuf => github.com/regen-network/protobuf v1.3.2-alpha.regen.4

replace github.com/MinterTeam/mhub2/module => ../../mhub2/module

require (
	github.com/MinterTeam/mhub2/module v0.0.0-20210419142331-8f69449b9069
	github.com/MinterTeam/minter-go-sdk/v2 v2.3.0
	github.com/cosmos/cosmos-sdk v0.43.0
	github.com/cosmos/go-bip39 v1.0.0
	github.com/spf13/viper v1.8.0
	github.com/tendermint/tendermint v0.34.12
	github.com/valyala/fasthttp v1.19.0
	google.golang.org/grpc v1.38.0
)
