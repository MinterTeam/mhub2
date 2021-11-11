package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/MinterTeam/mhub2/auto-tests/cosmos"
	"github.com/MinterTeam/mhub2/auto-tests/erc20"
	"github.com/MinterTeam/mhub2/module/app"
	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
	oracletypes "github.com/MinterTeam/mhub2/module/x/oracle/types"
	minterTypes "github.com/MinterTeam/minter-go-node/coreV2/types"
	"github.com/MinterTeam/minter-go-sdk/v2/api/http_client"
	"github.com/MinterTeam/minter-go-sdk/v2/transaction"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"math/big"
	"os"
	"strings"
	"sync"
	"time"
)

func init() {
	app.SetAddressConfig()
}

const (
	denom       = "stake"
	mhubChainId = "mhub-test"
	ethChainId  = 15
	bscChainId  = 16
)

type Context struct {
	TestsWg             *sync.WaitGroup
	EthPrivateKey       *ecdsa.PrivateKey
	EthClient           *ethclient.Client
	Bep20addr           string
	Erc20addr           string
	BscContract         string
	MinterClient        *http_client.Client
	EthContract         string
	BscClient           *ethclient.Client
	EthPrivateKeyString string
	CosmosConn          *grpc.ClientConn
	EthAddress          common.Address
	MinterMultisig      string
	CosmosMnemonic      string
}

func main() {
	wd := getWd()

	ethPrivateKey, _ := crypto.GenerateKey()
	ethPrivateKeyString := fmt.Sprintf("%x", crypto.FromECDSA(ethPrivateKey))
	ethAddress := crypto.PubkeyToAddress(ethPrivateKey.PublicKey)

	if err := os.RemoveAll(fmt.Sprintf("%s/data", wd)); err != nil {
		panic(err)
	}
	_ = os.Remove(fmt.Sprintf("%s/connector-status.json", wd))

	runOrPanic("mkdir %s/data", wd)
	runOrPanic("mkdir %s/data/logs", wd)
	runOrPanic(os.ExpandEnv("rm -rf $HOME/.mhub2"))

	runHoldersServer()

	// minter node
	runOrPanic("mkdir %s/data/minter", wd)
	runOrPanic("mkdir %s/data/minter/config", wd)
	runOrPanic("cp priv_validator.json data/minter/config", wd)
	runOrPanic("cp priv_validator_state.json data/minter/config", wd)
	runOrPanic("cp minter-config.toml data/minter/config/config.toml", wd)
	populateMinterGenesis(minterTypes.HexToPubkey("Mp582a2ed384d44ab3bf6d4bf751a515b62907fb54cdf496c02bcd1e3f6809b933"), ethAddress)
	go runOrPanic("minter node --testnet --home-dir=%s/data/minter", wd)
	time.Sleep(time.Second * 5)
	minterClient, _ := http_client.New("http://localhost:8843/v2")
	minterMultisig := createMinterMultisig(ethPrivateKeyString, ethAddress, minterClient)
	time.Sleep(time.Second * 5)
	fundMinterAddress(minterMultisig, ethPrivateKeyString, ethAddress, minterClient)

	// init and run ethereum node
	populateEthGenesis(ethAddress, ethChainId)
	populateEthGenesis(ethAddress, bscChainId)

	ethClient, err, ethContract, erc20addr := runEvmChain(ethChainId, wd, ethAddress, ethPrivateKey, "8545")
	if err != nil {
		panic(err)
	}

	bscClient, err, bscContract, bep20addr := runEvmChain(bscChainId, wd, ethAddress, ethPrivateKey, "8546")
	if err != nil {
		panic(err)
	}

	// init and run mhub2 testnet
	runOrPanic("mhub2 init --chain-id=%s validator1", mhubChainId)
	cosmosMnemonic := strings.Split(runOrPanic("mhub2 keys add --keyring-backend test validator1"), "\n")[11]
	hubAddress := strings.Trim(runOrPanic("mhub2 keys show validator1 -a --keyring-backend test"), "\n")

	runOrPanic("mhub2 add-genesis-account --keyring-backend test validator1 100000000000000000000000000%s,100000000000000000000000000hub", denom)
	runOrPanic("mhub2 add-genesis-token-info %s %s", erc20addr, bep20addr)

	runOrPanic("mhub2 gentx --keyring-backend test --moniker validator1 --chain-id=%s validator1 1000000000000000000000000%s %s %s 0x00", mhubChainId, denom, ethAddress.Hex(), hubAddress)
	runOrPanic("mhub2 collect-gentxs test")

	go runOrPanic("mhub2 start --trace --p2p.laddr tcp://0.0.0.0:36656")
	time.Sleep(time.Second)

	cosmosConn, err := grpc.DialContext(context.Background(), "localhost:9090", grpc.WithInsecure(), grpc.WithConnectParams(grpc.ConnectParams{
		Backoff:           backoff.DefaultConfig,
		MinConnectTimeout: time.Second * 5,
	}))
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 10)
	go runOrPanic("mhub-oracle --config=oracle-config.toml --cosmos-mnemonic=%s", cosmosMnemonic)

	go runOrPanic("mhub-minter-connector --config=connector-config.toml --cosmos-mnemonic=%s --minter-private-key=%s --minter-multisig-addr=%s", cosmosMnemonic, ethPrivateKeyString, minterMultisig)

	go runOrPanic("orchestrator --chain-id=ethereum --cosmos-phrase=%s --ethereum-key=%s --cosmos-grpc=%s --ethereum-rpc=%s --contract-address=%s --fees=%s --address-prefix=hub --metrics-listen=127.0.0.1:3000", cosmosMnemonic, ethPrivateKeyString, "http://localhost:9090", "http://localhost:8545", ethContract, denom)
	go runOrPanic("orchestrator --chain-id=bsc --cosmos-phrase=%s --ethereum-key=%s --cosmos-grpc=%s --ethereum-rpc=%s --contract-address=%s --fees=%s --address-prefix=hub --metrics-listen=127.0.0.1:3001", cosmosMnemonic, ethPrivateKeyString, "http://localhost:9090", "http://localhost:8546", bscContract, denom)

	approveERC20ToHub(ethPrivateKey, ethClient, ethContract, erc20addr, ethChainId)
	approveERC20ToHub(ethPrivateKey, bscClient, bscContract, bep20addr, bscChainId)

	time.Sleep(time.Second * 10)

	testsWg := &sync.WaitGroup{}

	ctx := &Context{
		TestsWg:             testsWg,
		EthPrivateKey:       ethPrivateKey,
		EthPrivateKeyString: ethPrivateKeyString,
		CosmosConn:          cosmosConn,
		EthAddress:          ethAddress,
		EthClient:           ethClient,
		EthContract:         ethContract,
		BscClient:           bscClient,
		BscContract:         bscContract,
		Erc20addr:           erc20addr,
		Bep20addr:           bep20addr,
		MinterClient:        minterClient,
		MinterMultisig:      minterMultisig,
		CosmosMnemonic:      cosmosMnemonic,
	}

	println("Start running tests")
	testEthereumToMinterTransfer(ctx)
	testEthereumToBscTransfer(ctx)
	testBSCToEthereumTransfer(ctx)
	testEthereumToHubTransfer(ctx)
	testBSCToHubTransfer(ctx)
	testMinterToHubTransfer(ctx)
	testHubToMinterTransfer(ctx)
	testHubToEthereumTransfer(ctx)
	testHubToBSCTransfer(ctx)
	testMinterToEthereumTransfer(ctx)

	// tests associated with holders
	ctx.TestsWg.Add(2)
	go func() {
		randomPk, _ := crypto.GenerateKey()
		recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()

		testUpdateHolders(ctx, recipient)
		testDiscountForHolder(ctx, recipient)
	}()

	testsWg.Wait()
	println("All tests are done")
}

func testDiscountForHolder(ctx *Context, recipient string) {
	sendMinterCoinToEthereum(ctx.EthPrivateKeyString, ctx.EthAddress, ctx.MinterMultisig, ctx.MinterClient, recipient)

	fee := int64(100)
	expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))
	expectedValue = expectedValue.AddRaw(fee)
	expectedValue = expectedValue.Sub(expectedValue.MulRaw(4).QuoRaw(1000))
	expectedValue = expectedValue.SubRaw(fee)

	startTime := time.Now()
	timeout := time.Minute * 2

	hubContract, _ := erc20.NewErc20(common.HexToAddress(ctx.Erc20addr), ctx.EthClient)

	for {
		hubBalanceInt, err := hubContract.BalanceOf(nil, common.HexToAddress(recipient))
		if err != nil {
			panic(err)
		}

		hubBalance := sdk.NewIntFromBigInt(hubBalanceInt)
		if hubBalance.IsZero() {
			if time.Now().Sub(startTime).Seconds() > timeout.Seconds() {
				panic("Timeout waiting for the balance to update")
			}

			time.Sleep(time.Second)
			continue
		}

		if !hubBalance.Equal(expectedValue) {
			panic(fmt.Sprintf("Balance is not equal to expected value. Expected %s, got %s", expectedValue.String(), hubBalance.String()))
		}

		println("SUCCESS: test Minter -> Ethereum transfer with discount")
		ctx.TestsWg.Done()
		break
	}
}

func testUpdateHolders(ctx *Context, recipient string) {
	balance := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(100)))
	HoldersList = map[string]string{
		recipient: balance.String(),
	}

	client := oracletypes.NewQueryClient(ctx.CosmosConn)
	startTime := time.Now()
	timeout := time.Minute

	for {
		response, err := client.Holders(context.TODO(), &oracletypes.QueryHoldersRequest{})
		if err != nil {
			panic(err)
		}

		if time.Now().Sub(startTime).Seconds() > timeout.Seconds() {
			panic("Timeout waiting for holders to update")
		}

		if response.Holders.Size() == 0 || len(response.Holders.List) == 0 {
			continue
		}

		if response.Holders.List[0].Address != recipient {
			continue
		}

		if !response.Holders.List[0].Value.Equal(balance) {
			continue
		}

		time.Sleep(time.Second)
		println("SUCCESS: update holders")
		ctx.TestsWg.Done()
		break
	}
}

func testMinterToEthereumTransfer(ctx *Context) {
	ctx.TestsWg.Add(1)
	randomPk, _ := crypto.GenerateKey()
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	sendMinterCoinToEthereum(ctx.EthPrivateKeyString, ctx.EthAddress, ctx.MinterMultisig, ctx.MinterClient, recipient)

	go func() {
		fee := int64(100)
		expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))
		expectedValue = expectedValue.AddRaw(fee)
		expectedValue = expectedValue.Sub(expectedValue.QuoRaw(100))
		expectedValue = expectedValue.SubRaw(fee)

		startTime := time.Now()
		timeout := time.Minute * 2

		hubContract, _ := erc20.NewErc20(common.HexToAddress(ctx.Erc20addr), ctx.EthClient)

		for {
			hubBalanceInt, err := hubContract.BalanceOf(nil, common.HexToAddress(recipient))
			if err != nil {
				panic(err)
			}

			hubBalance := sdk.NewIntFromBigInt(hubBalanceInt)
			if hubBalance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > timeout.Seconds() {
					panic("Timeout waiting for the balance to update")
				}

				time.Sleep(time.Second)
				continue
			}

			if !hubBalance.Equal(expectedValue) {
				panic(fmt.Sprintf("Balance is not equal to expected value. Expected %s, got %s", expectedValue.String(), hubBalance.String()))
			}

			println("SUCCESS: test Minter -> Ethereum transfer")
			ctx.TestsWg.Done()
			break
		}
	}()
}

func testHubToBSCTransfer(ctx *Context) {
	ctx.TestsWg.Add(1)

	fee := int64(100)
	addr, priv := cosmos.GetAccount(ctx.CosmosMnemonic)

	randomPk, _ := crypto.GenerateKey()
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	cosmos.SendCosmosTx([]sdk.Msg{
		&types.MsgSendToExternal{
			Sender:            addr.String(),
			ExternalRecipient: recipient,
			Amount:            sdk.NewInt64Coin("hub", 1e18),
			BridgeFee:         sdk.NewInt64Coin("hub", fee),
			ChainId:           "bsc",
		},
	}, addr, priv, ctx.CosmosConn, log.NewTMLogger(os.Stdout), true)

	go func() {
		expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))
		expectedValue = expectedValue.AddRaw(fee)
		expectedValue = expectedValue.Sub(expectedValue.QuoRaw(100))
		expectedValue = expectedValue.SubRaw(fee)

		startTime := time.Now()
		timeout := time.Minute * 3

		hubContract, _ := erc20.NewErc20(common.HexToAddress(ctx.Bep20addr), ctx.BscClient)

		for {
			hubBalanceInt, err := hubContract.BalanceOf(nil, common.HexToAddress(recipient))
			if err != nil {
				panic(err)
			}

			hubBalance := sdk.NewIntFromBigInt(hubBalanceInt)
			if hubBalance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > timeout.Seconds() {
					panic("Timeout waiting for the balance to update")
				}

				time.Sleep(time.Second)
				continue
			}

			if !hubBalance.Equal(expectedValue) {
				panic(fmt.Sprintf("Balance is not equal to expected value. Expected %s, got %s", expectedValue.String(), hubBalance.String()))
			}

			println("SUCCESS: test Hub -> Bsc transfer")
			ctx.TestsWg.Done()
			break
		}
	}()
}

func testHubToEthereumTransfer(ctx *Context) {
	ctx.TestsWg.Add(1)

	fee := int64(100)
	addr, priv := cosmos.GetAccount(ctx.CosmosMnemonic)

	randomPk, _ := crypto.GenerateKey()
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	cosmos.SendCosmosTx([]sdk.Msg{
		&types.MsgSendToExternal{
			Sender:            addr.String(),
			ExternalRecipient: recipient,
			Amount:            sdk.NewInt64Coin("hub", 1e18),
			BridgeFee:         sdk.NewInt64Coin("hub", fee),
			ChainId:           "ethereum",
		},
	}, addr, priv, ctx.CosmosConn, log.NewTMLogger(os.Stdout), true)

	go func() {
		expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))
		expectedValue = expectedValue.AddRaw(fee)
		expectedValue = expectedValue.Sub(expectedValue.QuoRaw(100))
		expectedValue = expectedValue.SubRaw(fee)

		startTime := time.Now()
		timeout := time.Minute * 3

		hubContract, _ := erc20.NewErc20(common.HexToAddress(ctx.Erc20addr), ctx.EthClient)

		for {
			hubBalanceInt, err := hubContract.BalanceOf(nil, common.HexToAddress(recipient))
			if err != nil {
				panic(err)
			}

			hubBalance := sdk.NewIntFromBigInt(hubBalanceInt)
			if hubBalance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > timeout.Seconds() {
					panic("Timeout waiting for the balance to update")
				}

				time.Sleep(time.Second)
				continue
			}

			if !hubBalance.Equal(expectedValue) {
				panic(fmt.Sprintf("Balance is not equal to expected value. Expected %s, got %s", expectedValue.String(), hubBalance.String()))
			}

			println("SUCCESS: test Hub -> Ethereum transfer")
			ctx.TestsWg.Done()
			break
		}
	}()
}

func testHubToMinterTransfer(ctx *Context) {
	ctx.TestsWg.Add(1)
	addr, priv := cosmos.GetAccount(ctx.CosmosMnemonic)

	randomPk, _ := crypto.GenerateKey()
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	cosmos.SendCosmosTx([]sdk.Msg{
		&types.MsgSendToExternal{
			Sender:            addr.String(),
			ExternalRecipient: recipient,
			Amount:            sdk.NewInt64Coin("hub", 1e18),
			BridgeFee:         sdk.NewInt64Coin("hub", 0),
			ChainId:           "minter",
		},
	}, addr, priv, ctx.CosmosConn, log.NewTMLogger(os.Stdout), true)

	go func() {
		expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))
		expectedValue = expectedValue.Sub(expectedValue.QuoRaw(100))
		startTime := time.Now()

		for {
			response, err := ctx.MinterClient.Address("Mx" + recipient[2:])
			if err != nil {
				panic(err)
			}

			hubBalance := getMinterCoinBalance(response.Balance, "HUB")
			if hubBalance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > 120 {
					panic("Timeout waiting for the balance to update")
				}

				time.Sleep(time.Second)
				continue
			}

			if !hubBalance.Equal(expectedValue) {
				panic(fmt.Sprintf("Balance is not equal to expected value. Expected %s, got %s", expectedValue.String(), hubBalance.String()))
			}

			println("SUCCESS: test Hub -> Minter transfer")
			ctx.TestsWg.Done()
			break
		}
	}()
}

func testMinterToHubTransfer(ctx *Context) {
	ctx.TestsWg.Add(1)
	hubRecipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()
	sendMinterCoinToHub(ctx.EthPrivateKeyString, ctx.EthAddress, ctx.MinterMultisig, ctx.MinterClient, hubRecipient) // minter
	go func() {
		client := banktypes.NewQueryClient(ctx.CosmosConn)
		expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))
		startTime := time.Now()

		for {
			response, err := client.Balance(context.TODO(), &banktypes.QueryBalanceRequest{
				Address: hubRecipient,
				Denom:   "hub",
			})
			if err != nil {
				panic(err)
			}

			if response.Balance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > 120 {
					panic("Timeout waiting for the balance to update")
				}

				time.Sleep(time.Second)
				continue
			}

			if !response.Balance.Amount.Equal(expectedValue) {
				panic(fmt.Sprintf("Balance is not equal to expected value. Expected %s, got %s", expectedValue.String(), response.Balance.Amount.String()))
			}

			println("SUCCESS: test Minter -> Hub transfer")
			ctx.TestsWg.Done()
			break
		}

	}()
}

func testBSCToHubTransfer(ctx *Context) {
	ctx.TestsWg.Add(1)
	hubRecipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()
	sendERC20ToHub(ctx.EthPrivateKey, ctx.BscClient, ctx.BscContract, ctx.Bep20addr, hubRecipient, bscChainId)

	go func() {
		expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))
		startTime := time.Now()
		client := banktypes.NewQueryClient(ctx.CosmosConn)

		for {
			response, err := client.Balance(context.TODO(), &banktypes.QueryBalanceRequest{
				Address: hubRecipient,
				Denom:   "hub",
			})
			if err != nil {
				panic(err)
			}

			if response.Balance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > 120 {
					panic("Timeout waiting for the balance to update")
				}

				time.Sleep(time.Second)
				continue
			}

			if !response.Balance.Amount.Equal(expectedValue) {
				panic(fmt.Sprintf("Balance is not equal to expected value. Expected %s, got %s", expectedValue.String(), response.Balance.Amount.String()))
			}

			println("SUCCESS: test BSC -> Hub transfer")
			ctx.TestsWg.Done()
			break
		}
	}()
}

func testEthereumToHubTransfer(ctx *Context) {
	ctx.TestsWg.Add(1)
	hubRecipient := sdk.AccAddress(secp256k1.GenPrivKey().PubKey().Address()).String()
	sendERC20ToHub(ctx.EthPrivateKey, ctx.EthClient, ctx.EthContract, ctx.Erc20addr, hubRecipient, ethChainId)

	go func() {
		expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))
		startTime := time.Now()
		client := banktypes.NewQueryClient(ctx.CosmosConn)

		for {
			response, err := client.Balance(context.TODO(), &banktypes.QueryBalanceRequest{
				Address: hubRecipient,
				Denom:   "hub",
			})
			if err != nil {
				panic(err)
			}

			if response.Balance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > 120 {
					panic("Timeout waiting for the balance to update")
				}

				time.Sleep(time.Second)
				continue
			}

			if !response.Balance.Amount.Equal(expectedValue) {
				panic(fmt.Sprintf("Balance is not equal to expected value. Expected %s, got %s", expectedValue.String(), response.Balance.Amount.String()))
			}

			println("SUCCESS: test Ethereum -> Hub transfer")
			ctx.TestsWg.Done()
			break
		}
	}()
}

func testBSCToEthereumTransfer(ctx *Context) {
	ctx.TestsWg.Add(1)
	randomPk, _ := crypto.GenerateKey()
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	sendERC20ToAnotherChain(ctx.EthPrivateKey, ctx.BscClient, ctx.BscContract, ctx.Bep20addr, recipient, bscChainId, "ethereum")

	go func() {
		expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))
		expectedValue = expectedValue.AddRaw(100)
		expectedValue = expectedValue.Sub(expectedValue.QuoRaw(100))
		expectedValue = expectedValue.SubRaw(100)

		startTime := time.Now()
		timeout := time.Minute * 3

		hubContract, _ := erc20.NewErc20(common.HexToAddress(ctx.Erc20addr), ctx.EthClient)

		for {
			hubBalanceInt, err := hubContract.BalanceOf(nil, common.HexToAddress(recipient))
			if err != nil {
				panic(err)
			}

			hubBalance := sdk.NewIntFromBigInt(hubBalanceInt)
			if hubBalance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > timeout.Seconds() {
					panic("Timeout waiting for the balance to update")
				}

				time.Sleep(time.Second)
				continue
			}

			if !hubBalance.Equal(expectedValue) {
				panic(fmt.Sprintf("Balance is not equal to expected value. Expected %s, got %s", expectedValue.String(), hubBalance.String()))
			}

			println("SUCCESS: test BSC -> Ethereum transfer")
			ctx.TestsWg.Done()
			break
		}
	}()
}

func testEthereumToMinterTransfer(ctx *Context) {
	ctx.TestsWg.Add(1)
	randomPk, _ := crypto.GenerateKey()
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	sendERC20ToAnotherChain(ctx.EthPrivateKey, ctx.EthClient, ctx.EthContract, ctx.Erc20addr, recipient, ethChainId, "minter")

	go func() {
		expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))
		expectedValue = expectedValue.AddRaw(100)
		expectedValue = expectedValue.Sub(expectedValue.QuoRaw(100))
		expectedValue = expectedValue.SubRaw(100)
		startTime := time.Now()

		for {
			response, err := ctx.MinterClient.Address("Mx" + recipient[2:])
			if err != nil {
				panic(err)
			}

			hubBalance := getMinterCoinBalance(response.Balance, "HUB")
			if hubBalance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > 180 {
					panic("Timeout waiting for the balance to update")
				}

				time.Sleep(time.Second)
				continue
			}

			if !hubBalance.Equal(expectedValue) {
				panic(fmt.Sprintf("Balance is not equal to expected value. Expected %s, got %s", expectedValue.String(), hubBalance.String()))
			}

			println("SUCCESS: test Eth -> Minter transfer")
			ctx.TestsWg.Done()
			break
		}
	}()
}

func testEthereumToBscTransfer(ctx *Context) {
	ctx.TestsWg.Add(1)
	randomPk, _ := crypto.GenerateKey()
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	sendERC20ToAnotherChain(ctx.EthPrivateKey, ctx.EthClient, ctx.EthContract, ctx.Erc20addr, recipient, ethChainId, "bsc")

	go func() {
		expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))
		expectedValue = expectedValue.AddRaw(100)
		expectedValue = expectedValue.Sub(expectedValue.QuoRaw(100))
		expectedValue = expectedValue.SubRaw(100)

		startTime := time.Now()
		timeout := time.Minute * 3

		hubContract, _ := erc20.NewErc20(common.HexToAddress(ctx.Bep20addr), ctx.BscClient)

		for {
			hubBalanceInt, err := hubContract.BalanceOf(nil, common.HexToAddress(recipient))
			if err != nil {
				panic(err)
			}

			hubBalance := sdk.NewIntFromBigInt(hubBalanceInt)
			if hubBalance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > timeout.Seconds() {
					panic("Timeout waiting for the balance to update")
				}

				time.Sleep(time.Second)
				continue
			}

			if !hubBalance.Equal(expectedValue) {
				panic(fmt.Sprintf("Balance is not equal to expected value. Expected %s, got %s", expectedValue.String(), hubBalance.String()))
			}

			println("SUCCESS: test Ethereum -> BSC transfer")
			ctx.TestsWg.Done()
			break
		}
	}()
}
