package main

import (
	"context"
	"crypto/ecdsa"
	"fmt"
	"github.com/MinterTeam/mhub2/auto-tests/cosmos"
	"github.com/MinterTeam/mhub2/auto-tests/erc20"
	"github.com/MinterTeam/mhub2/auto-tests/hub2"
	"github.com/MinterTeam/mhub2/module/app"
	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
	mhubtypes "github.com/MinterTeam/mhub2/module/x/mhub2/types"
	oracletypes "github.com/MinterTeam/mhub2/module/x/oracle/types"
	minterTypes "github.com/MinterTeam/minter-go-node/coreV2/types"
	"github.com/MinterTeam/minter-go-sdk/v2/api/grpc_client"
	"github.com/MinterTeam/minter-go-sdk/v2/transaction"
	"github.com/MinterTeam/node-grpc-gateway/api_pb"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	secp256k12 "github.com/ethereum/go-ethereum/crypto/secp256k1"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/status-im/keycard-go/hexutils"
	"github.com/stretchr/testify/assert"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"math"
	"math/big"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

func init() {
	app.SetAddressConfig()
}

/// Requirements on local binaries:
/// 1. minter node - v2.6.0+
/// 2. geth node - v1.10.11+
/// 3. oracle - v2
/// 4. orchestrator - v2
/// 5. minter-connector - v2
/// 6. mhub - v2

const (
	denom              = "stake"
	mhubChainId        = "mhub-test"
	ethChainId         = 15
	bscChainId         = 16
	ethSignaturePrefix = "\x19Ethereum Signed Message:\n32"
	minterCoinId       = 1
)

type Context struct {
	TestsWg             *sync.WaitGroup
	EthPrivateKey       *ecdsa.PrivateKey
	EthClient           *ethclient.Client
	Bep20addr           string
	Erc20addr           string
	BscContract         string
	MinterClient        *grpc_client.Client
	EthContract         string
	BscClient           *ethclient.Client
	EthPrivateKeyString string
	CosmosConn          *grpc.ClientConn
	EthAddress          common.Address
	MinterMultisig      string
	CosmosMnemonic      string
	WethAddr            string
	WbnbAddr            string
}

var (
	bridgeCommission, _ = sdk.NewDecFromStr("0.01")
	testTimeout         = time.Minute * 10
)

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
	runEthFeeCalculatorServer()

	// minter node
	runOrPanic("mkdir %s/data/minter", wd)
	runOrPanic("mkdir %s/data/minter/config", wd)
	runOrPanic("cp priv_validator.json data/minter/config", wd)
	runOrPanic("cp priv_validator_state.json data/minter/config", wd)
	runOrPanic("cp minter-config.toml data/minter/config/config.toml", wd)
	populateMinterGenesis(minterTypes.HexToPubkey("Mp582a2ed384d44ab3bf6d4bf751a515b62907fb54cdf496c02bcd1e3f6809b933"), ethAddress)
	go runOrPanic("minter node --testnet --home-dir=%s/data/minter", wd)

	minterClient, err := grpc_client.New("localhost:8842")
	for {
		if _, err := minterClient.Status(); err != nil {
			time.Sleep(time.Millisecond * 200)
			continue
		}

		break
	}
	minterMultisig := createMinterMultisig(ethPrivateKeyString, ethAddress, minterClient)
	fundMinterAddress(minterMultisig, ethPrivateKeyString, ethAddress, minterClient)

	// init and run ethereum node
	populateEthGenesis(ethAddress, ethChainId)
	populateEthGenesis(ethAddress, bscChainId)

	wg := sync.WaitGroup{}
	var (
		ethClient   *ethclient.Client
		ethContract string
		erc20addr   string
		wethAddr    string

		bscClient   *ethclient.Client
		bscContract string
		bep20addr   string
		wbnbAddr    string
	)

	wg.Add(2)
	go func() {
		ethClient, err, ethContract, erc20addr, wethAddr = runEvmChain(ethChainId, wd, ethAddress, ethPrivateKey, ethPrivateKeyString, "8545")
		if err != nil {
			panic(err)
		}
		wg.Done()
	}()

	go func() {
		bscClient, err, bscContract, bep20addr, wbnbAddr = runEvmChain(bscChainId, wd, ethAddress, ethPrivateKey, ethPrivateKeyString, "8546")
		if err != nil {
			panic(err)
		}
		wg.Done()
	}()

	wg.Wait()

	// init and run mhub2 testnet
	runOrPanic("mhub2 init --chain-id=%s validator1", mhubChainId)
	cosmosMnemonic := strings.Split(runOrPanic("mhub2 keys add --keyring-backend test validator1"), "\n")[11]
	hubAddress := strings.Trim(runOrPanic("mhub2 keys show validator1 -a --keyring-backend test"), "\n")

	runOrPanic("mhub2 add-genesis-account --keyring-backend test validator1 100000000000000000000000000%s,100000000000000000000000000hub", denom)
	runOrPanic("mhub2 prepare-genesis-for-tests %s %s %s %s %s", erc20addr, bep20addr, strconv.Itoa(minterCoinId), "2", wethAddr)

	valAddress, _ := cosmos.GetAccount(cosmosMnemonic)
	signMsgBz := app.MakeEncodingConfig().Marshaler.MustMarshal(&types.DelegateKeysSignMsg{
		ValidatorAddress: sdk.ValAddress(valAddress).String(),
		Nonce:            0,
	})
	hash := append([]uint8(ethSignaturePrefix), crypto.Keccak256Hash(signMsgBz).Bytes()...)
	sig, err := secp256k12.Sign(crypto.Keccak256Hash(hash).Bytes(), hexutils.HexToBytes(ethPrivateKeyString))
	if err != nil {
		panic(err)
	}

	runOrPanic("mhub2 gentx --keyring-backend test --moniker validator1 --chain-id=%s validator1 1000000000000000000000000%s %s %s 0x%x", mhubChainId, denom, ethAddress.Hex(), hubAddress, sig)
	runOrPanic("mhub2 collect-gentxs test")

	runOrPanic(os.ExpandEnv("cp mhub2-config.toml $HOME/.mhub2/config/config.toml"))
	runOrPanic(os.ExpandEnv("cp mhub2-app.toml $HOME/.mhub2/config/app.toml"))
	go runOrPanic("mhub2 start --trace --p2p.laddr tcp://0.0.0.0:36656")

	cosmosConn, err := grpc.DialContext(context.Background(), "localhost:9090", grpc.WithInsecure(), grpc.WithConnectParams(grpc.ConnectParams{
		Backoff:           backoff.DefaultConfig,
		MinConnectTimeout: time.Second * 5,
	}))
	if err != nil {
		panic(err)
	}

	time.Sleep(time.Second * 10)
	go runOrPanic("mhub-minter-connector --config=connector-config.toml --cosmos-mnemonic=%s --minter-private-key=%s --minter-multisig-addr=%s", cosmosMnemonic, ethPrivateKeyString, minterMultisig)
	time.Sleep(time.Second * 10)
	go runOrPanic("mhub-oracle --testnet --config=oracle-config.toml")

	client := oracletypes.NewQueryClient(cosmosConn)
	startTime := time.Now()

	for {
		response, err := client.Prices(context.TODO(), &oracletypes.QueryPricesRequest{})
		if err != nil {
			panic(err)
		}

		if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
			panic("Timeout waiting for prices to update")
		}

		if len(response.GetPrices().GetList()) == 0 {
			continue
		}

		time.Sleep(time.Second)
		println("SUCCESS: update prices")
		break
	}

	go runOrPanic("orchestrator --chain-id=ethereum --eth-fee-calculator-url=http://localhost:8840 --ethereum-key=%s --cosmos-grpc=%s --ethereum-rpc=%s --contract-address=%s --metrics-listen=127.0.0.1:3000", ethPrivateKeyString, "http://localhost:9090", "http://localhost:8545", ethContract)
	go runOrPanic("orchestrator --chain-id=bsc --eth-fee-calculator-url=http://localhost:8840 --ethereum-key=%s --cosmos-grpc=%s --ethereum-rpc=%s --contract-address=%s --metrics-listen=127.0.0.1:3001", ethPrivateKeyString, "http://localhost:9090", "http://localhost:8546", bscContract)
	approveERC20ToHub(ethPrivateKey, ethClient, ethContract, erc20addr, ethChainId)
	approveERC20ToHub(ethPrivateKey, bscClient, bscContract, bep20addr, bscChainId)

	time.Sleep(time.Second * 10)

	ctx := &Context{
		TestsWg:             &sync.WaitGroup{},
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
		WethAddr:            wethAddr,
		WbnbAddr:            wbnbAddr,
	}

	println("Start running preparation tests")
	testMinterMultisigChanges(ctx)
	testEthereumMultisigChanges(ctx)
	testBSCMultisigChanges(ctx)
	ctx.TestsWg.Wait()
	time.Sleep(time.Second * 20) // todo: wait for all of networks to confirm valset updates

	println("Start real tests")
	testEthereumToMinterTransfer(ctx)
	testEthereumToBscTransfer(ctx)
	testEthereumToHubTransfer(ctx)
	testBSCToEthereumTransfer(ctx)
	testBSCToHubTransfer(ctx)
	testHubToMinterTransfer(ctx)
	testHubToEthereumTransfer(ctx)
	testHubToBSCTransfer(ctx)
	testMinterToHubTransfer(ctx)
	testMinterToEthereumTransfer(ctx)
	testMinterToBscTransfer(ctx)
	testColdStorageTransfer(ctx)
	testVoteForTokenInfosUpdate(ctx)
	testEthEthereumToMinterTransfer(ctx)
	testEthMinterToEthereumTransfer(ctx)

	// tests associated with holders
	ctx.TestsWg.Add(2)
	go func() {
		randomPk, _ := crypto.GenerateKey()
		recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()

		testUpdateHolders(ctx, recipient)
		testDiscountForHolder(ctx, recipient)
	}()
	ctx.TestsWg.Wait()

	println("Running test for fee refund")
	testFeeRefund(ctx)
	ctx.TestsWg.Wait()

	println("Running test for fee refund with different decimals")
	testFeeRefundWithDifferentDecimals(ctx)
	ctx.TestsWg.Wait()

	println("Running test for transaction relayer fee")
	testFeeForTransactionRelayer(ctx)
	ctx.TestsWg.Wait()

	println("Running test for validator's commission")
	testValidatorsCommissions(ctx)
	ctx.TestsWg.Wait()

	println("Running test for validator's commission with different decimals")
	testValidatorsCommissionsWithDifferentDecimals(ctx)
	ctx.TestsWg.Wait()

	println("Killing orchestrator")
	stopProcess("orchestrator")
	testTxTimeout(ctx)
	testTxTimeoutDifferentDecimals(ctx)
	ctx.TestsWg.Wait()

	println("All tests are done")
}

func testVoteForTokenInfosUpdate(ctx *Context) {
	ctx.TestsWg.Add(1)
	addr, priv := cosmos.GetAccount(ctx.CosmosMnemonic)

	initialDeposit := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(10000000)))

	client := mhubtypes.NewQueryClient(ctx.CosmosConn)

	response, err := client.TokenInfos(context.TODO(), &mhubtypes.TokenInfosRequest{})
	if err != nil {
		panic(err)
	}

	newInfos := &response.List

	newInfos.TokenInfos = append(newInfos.TokenInfos, &mhubtypes.TokenInfo{
		Id:               999,
		Denom:            "test",
		ChainId:          "minter",
		ExternalTokenId:  "123",
		ExternalDecimals: 18,
		Commission:       bridgeCommission,
	})

	proposal := &govtypes.MsgSubmitProposal{}
	proposal.SetInitialDeposit(initialDeposit)
	proposal.SetProposer(addr)
	if err := proposal.SetContent(&mhubtypes.TokenInfosChangeProposal{
		NewInfos: newInfos,
	}); err != nil {
		panic(err)
	}

	cosmos.SendCosmosTx([]sdk.Msg{
		proposal,
		govtypes.NewMsgVote(addr, 4, govtypes.OptionYes),
	}, addr, priv, ctx.CosmosConn, log.NewTMLogger(os.Stdout), true)

	go func() {
		startTime := time.Now()
		for {
			if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
				panic("Timeout waiting for the token infos to update")
			}

			response, err := client.TokenInfos(context.TODO(), &mhubtypes.TokenInfosRequest{})
			if err != nil {
				panic(err)
			}

			if !assert.ObjectsAreEqual(&response.List, newInfos) {
				time.Sleep(time.Second)
				continue
			}

			println("SUCCESS: vote for new token list")
			ctx.TestsWg.Done()
			break
		}
	}()
}

func testColdStorageTransfer(ctx *Context) {
	ctx.TestsWg.Add(1)
	addr, priv := cosmos.GetAccount(ctx.CosmosMnemonic)

	recipientMinter := "0x7072558b2b91e62dbed78e9a3453e5c9e01fec5e"
	recipientEth := "0x58BD8047F441B9D511aEE9c581aEb1caB4FE0b6d"
	recipientBsc := "0xbCc2Fa395c6198096855c932f4087cF1377d28EE"
	initialDeposit := sdk.NewCoins(sdk.NewCoin(denom, sdk.NewInt(10000000)))

	proposalEth := &govtypes.MsgSubmitProposal{}
	proposalEth.SetInitialDeposit(initialDeposit)
	proposalEth.SetProposer(addr)
	if err := proposalEth.SetContent(&mhubtypes.ColdStorageTransferProposal{ChainId: "ethereum", Amount: sdk.NewCoins(sdk.NewCoin("hub", sdk.NewInt(1e18)))}); err != nil {
		panic(err)
	}

	proposalBsc := &govtypes.MsgSubmitProposal{}
	proposalBsc.SetInitialDeposit(initialDeposit)
	proposalBsc.SetProposer(addr)
	if err := proposalBsc.SetContent(&mhubtypes.ColdStorageTransferProposal{ChainId: "bsc", Amount: sdk.NewCoins(sdk.NewCoin("hub", sdk.NewInt(1e18)))}); err != nil {
		panic(err)
	}

	proposalMinter := &govtypes.MsgSubmitProposal{}
	proposalMinter.SetInitialDeposit(initialDeposit)
	proposalMinter.SetProposer(addr)
	if err := proposalMinter.SetContent(&mhubtypes.ColdStorageTransferProposal{ChainId: "minter", Amount: sdk.NewCoins(sdk.NewCoin("hub", sdk.NewInt(1e18)))}); err != nil {
		panic(err)
	}

	cosmos.SendCosmosTx([]sdk.Msg{
		proposalEth,
		proposalMinter,
		proposalBsc,
		govtypes.NewMsgVote(addr, 1, govtypes.OptionYes),
		govtypes.NewMsgVote(addr, 2, govtypes.OptionYes),
		govtypes.NewMsgVote(addr, 3, govtypes.OptionYes),
	}, addr, priv, ctx.CosmosConn, log.NewTMLogger(os.Stdout), true)

	go func() {
		expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))

		startTime := time.Now()

		hubContractEth, _ := erc20.NewErc20(common.HexToAddress(ctx.Erc20addr), ctx.EthClient)
		hubContractBsc, _ := erc20.NewErc20(common.HexToAddress(ctx.Bep20addr), ctx.BscClient)

		for {
			// eth
			{
				hubBalanceEthInt, err := hubContractEth.BalanceOf(nil, common.HexToAddress(recipientEth))
				if err != nil {
					panic(err)
				}

				hubBalanceEth := sdk.NewIntFromBigInt(hubBalanceEthInt)
				if hubBalanceEth.IsZero() {
					if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
						panic("Timeout waiting for the balance to update")
					}

					time.Sleep(time.Second)
					continue
				}

				if !hubBalanceEth.Equal(expectedValue) {
					panic(fmt.Sprintf("Balance is not equal to expected value. Expected %s, got %s", expectedValue.String(), hubBalanceEth.String()))
				}
			}

			// bsc
			{
				hubBalanceBscInt, err := hubContractBsc.BalanceOf(nil, common.HexToAddress(recipientBsc))
				if err != nil {
					panic(err)
				}

				hubBalanceBsc := sdk.NewIntFromBigInt(hubBalanceBscInt)
				if hubBalanceBsc.IsZero() {
					if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
						panic("Timeout waiting for the balance to update")
					}

					time.Sleep(time.Second)
					continue
				}

				if !hubBalanceBsc.Equal(expectedValue.QuoRaw(1e12)) {
					panic(fmt.Sprintf("Balance is not equal to expected value. Expected %s, got %s", expectedValue.String(), hubBalanceBsc.String()))
				}
			}

			// minter
			{
				response, err := ctx.MinterClient.Address("Mx" + recipientMinter[2:])
				if err != nil {
					panic(err)
				}

				hubBalance := getMinterCoinBalance(response.Balance, minterCoinId)
				if hubBalance.IsZero() {
					if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
						panic("Timeout waiting for the balance to update")
					}

					time.Sleep(time.Second)
					continue
				}

				if !hubBalance.Equal(expectedValue) {
					panic(fmt.Sprintf("Balance is not equal to expected value. Expected %s, got %s", expectedValue.String(), hubBalance.String()))
				}
			}

			println("SUCCESS: test cold storage transfer")
			ctx.TestsWg.Done()
			break
		}
	}()
}

func testMinterMultisigChanges(ctx *Context) {
	ctx.TestsWg.Add(1)
	go func() {
		startTime := time.Now()

		for {
			if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
				panic("Timeout waiting for the minter multisig to change")
			}

			response, err := ctx.MinterClient.Address(ctx.MinterMultisig)
			if err != nil {
				panic(err)
			}

			if strings.ToLower(response.Multisig.Addresses[0][2:]) == strings.ToLower(ctx.EthAddress.Hex()[2:]) && response.Multisig.Weights[0] == 1000 {
				println("SUCCESS: test minter multisig changes")
				ctx.TestsWg.Done()
				return
			}

			time.Sleep(time.Second)
		}
	}()
}

func testEthereumMultisigChanges(ctx *Context) {
	ctx.TestsWg.Add(1)

	client, err := hub2.NewHub2(common.HexToAddress(ctx.EthContract), ctx.EthClient)
	if err != nil {
		panic(err)
	}

	mhubClient := mhubtypes.NewQueryClient(ctx.CosmosConn)
	response, err := mhubClient.LatestSignerSetTx(context.TODO(), &mhubtypes.LatestSignerSetTxRequest{ChainId: "ethereum"})
	targetCheckpointStr := fmt.Sprintf("%x", response.SignerSet.GetCheckpoint([]byte("defaultgravityid")))

	go func() {
		startTime := time.Now()

		for {
			if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
				panic("Timeout waiting for the ethereum multisig to change")
			}

			checkpoint, err := client.StateLastValsetCheckpoint(nil)
			if err != nil {
				panic(err)
			}

			checkpointStr := fmt.Sprintf("%x", checkpoint)
			if checkpointStr == targetCheckpointStr {
				println("SUCCESS: test ethereum multisig changes")
				ctx.TestsWg.Done()
				return
			}

			time.Sleep(time.Second)
		}
	}()
}

func testBSCMultisigChanges(ctx *Context) {
	ctx.TestsWg.Add(1)

	client, err := hub2.NewHub2(common.HexToAddress(ctx.BscContract), ctx.BscClient)
	if err != nil {
		panic(err)
	}

	mhubClient := mhubtypes.NewQueryClient(ctx.CosmosConn)
	response, err := mhubClient.LatestSignerSetTx(context.TODO(), &mhubtypes.LatestSignerSetTxRequest{ChainId: "bsc"})
	targetCheckpointStr := fmt.Sprintf("%x", response.SignerSet.GetCheckpoint([]byte("defaultgravityid")))

	go func() {
		startTime := time.Now()

		for {
			if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
				panic("Timeout waiting for the bsc multisig to change")
			}

			checkpoint, err := client.StateLastValsetCheckpoint(nil)
			if err != nil {
				panic(err)
			}

			checkpointStr := fmt.Sprintf("%x", checkpoint)
			if checkpointStr == targetCheckpointStr {
				println("SUCCESS: test bsc multisig changes")
				ctx.TestsWg.Done()
				return
			}

			time.Sleep(time.Second)
		}
	}()
}

func testFeeForTransactionRelayer(ctx *Context) {
	ctx.TestsWg.Add(1)
	randomPk, _ := crypto.GenerateKey()
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	fee := sdk.NewInt(8888)
	sendMinterCoinToEthereum(ctx.EthPrivateKeyString, ctx.EthAddress, ctx.MinterMultisig, ctx.MinterClient, minterCoinId, recipient, sdk.NewInt(1e18), fee)
	minterStatus, err := ctx.MinterClient.Status()
	if err != nil {
		panic(err)
	}

	minterHeight := minterStatus.LatestBlockHeight

	go func() {
		startTime := time.Now()

		for {
			if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
				panic("Timeout waiting for the fee to arrive to relayer")
			}

			response, err := ctx.MinterClient.Block(minterHeight)
			if err != nil {
				time.Sleep(time.Millisecond * 200)
				continue
			}

			for _, transactionResponse := range response.Transactions {
				if transactionResponse.From != ctx.MinterMultisig {
					continue
				}

				if transaction.Type(transactionResponse.Type) == transaction.TypeMultisend {
					var data api_pb.MultiSendData
					if err := transactionResponse.Data.UnmarshalTo(&data); err != nil {
						panic(err)
					}

					for _, item := range data.List {
						if item.Coin.Id == minterCoinId && strings.ToLower(ctx.EthAddress.String())[2:] == item.To[2:] && item.Value == fee.String() {
							println("SUCCESS: test fee for transaction relayer")
							ctx.TestsWg.Done()
							return
						}
					}
				}
			}

			minterHeight++
		}
	}()
}

func testDiscountForHolder(ctx *Context, recipient string) {
	sendMinterCoinToEthereum(ctx.EthPrivateKeyString, ctx.EthAddress, ctx.MinterMultisig, ctx.MinterClient, minterCoinId, recipient, sdk.NewInt(1e18), sdk.NewInt(100))

	fee := int64(100)
	expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))
	expectedValue = expectedValue.Sub(expectedValue.MulRaw(4).QuoRaw(1000))
	expectedValue = expectedValue.SubRaw(fee)

	startTime := time.Now()

	hubContract, _ := erc20.NewErc20(common.HexToAddress(ctx.Erc20addr), ctx.EthClient)

	for {
		hubBalanceInt, err := hubContract.BalanceOf(nil, common.HexToAddress(recipient))
		if err != nil {
			panic(err)
		}

		hubBalance := sdk.NewIntFromBigInt(hubBalanceInt)
		if hubBalance.IsZero() {
			if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
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
	HoldersList = HoldersResult{
		Data: []struct {
			Address string `json:"address"`
			Balance string `json:"balance"`
		}{
			{
				Address: recipient[2:],
				Balance: balance.String(),
			},
		},
	}

	client := oracletypes.NewQueryClient(ctx.CosmosConn)
	startTime := time.Now()

	for {
		response, err := client.Holders(context.TODO(), &oracletypes.QueryHoldersRequest{})
		if err != nil {
			panic(err)
		}

		if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
			panic("Timeout waiting for holders to update")
		}

		if response.Holders.Size() == 0 || len(response.Holders.List) == 0 {
			continue
		}

		if response.Holders.List[0].Address != strings.ToLower(recipient[2:]) {
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

func testEthMinterToEthereumTransfer(ctx *Context) {
	ctx.TestsWg.Add(1)
	randomPk, _ := crypto.GenerateKey()
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	sendMinterCoinToEthereum(ctx.EthPrivateKeyString, ctx.EthAddress, ctx.MinterMultisig, ctx.MinterClient, 2, recipient, sdk.NewInt(1e18), sdk.NewInt(100))

	go func() {
		fee := int64(100)
		expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))
		expectedValue = expectedValue.Sub(expectedValue.ToDec().Mul(bridgeCommission).TruncateInt())
		expectedValue = expectedValue.SubRaw(fee)

		startTime := time.Now()

		for {
			ethBalanceInt, err := ctx.EthClient.BalanceAt(context.TODO(), common.HexToAddress(recipient), nil)
			if err != nil {
				panic(err)
			}

			ethBalance := sdk.NewIntFromBigInt(ethBalanceInt)
			if ethBalance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
					panic("Timeout waiting for the balance to update")
				}

				time.Sleep(time.Second)
				continue
			}

			if !ethBalance.Equal(expectedValue) {
				panic(fmt.Sprintf("Balance is not equal to expected value. Expected %s, got %s", expectedValue.String(), ethBalance.String()))
			}

			println("SUCCESS: test Minter -> Ethereum (ETH) transfer")
			ctx.TestsWg.Done()
			break
		}
	}()
}

func testMinterToEthereumTransfer(ctx *Context) {
	ctx.TestsWg.Add(1)
	randomPk, _ := crypto.GenerateKey()
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	sendMinterCoinToEthereum(ctx.EthPrivateKeyString, ctx.EthAddress, ctx.MinterMultisig, ctx.MinterClient, minterCoinId, recipient, sdk.NewInt(1e18), sdk.NewInt(100))

	go func() {
		fee := int64(100)
		expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))
		expectedValue = expectedValue.Sub(expectedValue.ToDec().Mul(bridgeCommission).TruncateInt())
		expectedValue = expectedValue.SubRaw(fee)

		startTime := time.Now()

		hubContract, _ := erc20.NewErc20(common.HexToAddress(ctx.Erc20addr), ctx.EthClient)

		for {
			hubBalanceInt, err := hubContract.BalanceOf(nil, common.HexToAddress(recipient))
			if err != nil {
				panic(err)
			}

			hubBalance := sdk.NewIntFromBigInt(hubBalanceInt)
			if hubBalance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
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

func testMinterToBscTransfer(ctx *Context) {
	ctx.TestsWg.Add(1)
	randomPk, _ := crypto.GenerateKey()
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	sendMinterCoinToBsc(ctx.EthPrivateKeyString, ctx.EthAddress, ctx.MinterMultisig, ctx.MinterClient, recipient, sdk.NewInt(1e18), sdk.NewInt(100))

	go func() {
		fee := int64(100)
		expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))
		expectedValue = expectedValue.Sub(expectedValue.ToDec().Mul(bridgeCommission).TruncateInt())
		expectedValue = expectedValue.SubRaw(fee)
		expectedValue = expectedValue.QuoRaw(1e12)

		startTime := time.Now()

		hubContract, _ := erc20.NewErc20(common.HexToAddress(ctx.Bep20addr), ctx.BscClient)

		for {
			hubBalanceInt, err := hubContract.BalanceOf(nil, common.HexToAddress(recipient))
			if err != nil {
				panic(err)
			}

			hubBalance := sdk.NewIntFromBigInt(hubBalanceInt)
			if hubBalance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
					panic("Timeout waiting for the balance to update")
				}

				time.Sleep(time.Second)
				continue
			}

			if !hubBalance.Equal(expectedValue) {
				panic(fmt.Sprintf("Balance is not equal to expected value. Expected %s, got %s", expectedValue.String(), hubBalance.String()))
			}

			println("SUCCESS: test Minter -> Bsc transfer")
			ctx.TestsWg.Done()
			break
		}
	}()
}

func testTxTimeout(ctx *Context) {
	ctx.TestsWg.Add(1)
	randomPk, _ := crypto.GenerateKey()
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	value := big.NewInt(1e18)
	sendMinterCoinToEthereum(ctx.EthPrivateKeyString, ctx.EthAddress, ctx.MinterMultisig, ctx.MinterClient, minterCoinId, recipient, sdk.NewInt(1e18), sdk.NewInt(100))

	minterStatus, err := ctx.MinterClient.Status()
	if err != nil {
		panic(err)
	}

	mhubClient := mhubtypes.NewQueryClient(ctx.CosmosConn)
	for {
		response, err := mhubClient.UnbatchedSendToExternals(context.TODO(), &mhubtypes.UnbatchedSendToExternalsRequest{
			ChainId: "ethereum",
		})
		if err != nil {
			panic(err)
		}

		if len(response.SendToExternals) > 0 {
			println("waiting for batch to be created")
			time.Sleep(time.Second)
			continue
		}

		break
	}

	addr, priv := cosmos.GetAccount(ctx.CosmosMnemonic)
	response, err := mhubClient.LastSubmittedExternalEvent(context.TODO(), &mhubtypes.LastSubmittedExternalEventRequest{
		Address: addr.String(),
		ChainId: "ethereum",
	})
	if err != nil {
		panic(err)
	}

	event, err := mhubtypes.PackEvent(&mhubtypes.SendToHubEvent{
		EventNonce:     response.EventNonce + 1,
		ExternalCoinId: ctx.Erc20addr,
		Amount:         sdk.NewInt(1),
		Sender:         recipient,
		CosmosReceiver: addr.String(),
		ExternalHeight: math.MaxUint64,
		TxHash:         "Mt...",
	})
	if err != nil {
		panic(err)
	}
	cosmos.SendCosmosTx([]sdk.Msg{&mhubtypes.MsgSubmitExternalEvent{
		Event:   event,
		Signer:  addr.String(),
		ChainId: "ethereum",
	}}, addr, priv, ctx.CosmosConn, log.NewTMLogger(os.Stdout), true)

	minterHeight := minterStatus.LatestBlockHeight

	go func() {
		startTime := time.Now()

		for {
			if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
				panic("Timeout waiting for the refund")
			}

			response, err := ctx.MinterClient.Block(minterHeight)
			if err != nil {
				time.Sleep(time.Millisecond * 200)
				continue
			}

			for _, transactionResponse := range response.Transactions {
				if transactionResponse.From != ctx.MinterMultisig {
					continue
				}

				if transaction.Type(transactionResponse.Type) == transaction.TypeMultisend {
					var data api_pb.MultiSendData
					if err := transactionResponse.Data.UnmarshalTo(&data); err != nil {
						panic(err)
					}

					for _, item := range data.List {
						if item.Coin.Id == minterCoinId && strings.ToLower(ctx.EthAddress.String())[2:] == item.To[2:] && item.Value == value.String() { // todo???
							println("SUCCESS: test refunds")
							ctx.TestsWg.Done()
							return
						}
					}
				}
			}

			minterHeight++
		}
	}()
}

func testTxTimeoutDifferentDecimals(ctx *Context) {
	ctx.TestsWg.Add(1)
	randomPk, _ := crypto.GenerateKey()
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	value := sdk.NewInt(1e18)
	expectedValue := sdk.NewInt(999999000000000000) // due to conversion errors
	sendMinterCoinToBsc(ctx.EthPrivateKeyString, ctx.EthAddress, ctx.MinterMultisig, ctx.MinterClient, recipient, value, sdk.NewInt(100))

	minterStatus, err := ctx.MinterClient.Status()
	if err != nil {
		panic(err)
	}

	mhubClient := mhubtypes.NewQueryClient(ctx.CosmosConn)
	for {
		response, err := mhubClient.UnbatchedSendToExternals(context.TODO(), &mhubtypes.UnbatchedSendToExternalsRequest{
			ChainId: "bsc",
		})
		if err != nil {
			panic(err)
		}

		if len(response.SendToExternals) > 0 {
			println("waiting for batch to be created")
			time.Sleep(time.Second)
			continue
		}

		break
	}

	addr, priv := cosmos.GetAccount(ctx.CosmosMnemonic)
	response, err := mhubClient.LastSubmittedExternalEvent(context.TODO(), &mhubtypes.LastSubmittedExternalEventRequest{
		Address: addr.String(),
		ChainId: "bsc",
	})
	if err != nil {
		panic(err)
	}

	event, err := mhubtypes.PackEvent(&mhubtypes.SendToHubEvent{
		EventNonce:     response.EventNonce + 1,
		ExternalCoinId: ctx.Erc20addr,
		Amount:         sdk.NewInt(1),
		Sender:         recipient,
		CosmosReceiver: addr.String(),
		ExternalHeight: math.MaxUint64,
		TxHash:         "Mt...",
	})
	if err != nil {
		panic(err)
	}
	cosmos.SendCosmosTx([]sdk.Msg{&mhubtypes.MsgSubmitExternalEvent{
		Event:   event,
		Signer:  addr.String(),
		ChainId: "bsc",
	}}, addr, priv, ctx.CosmosConn, log.NewTMLogger(os.Stdout), true)

	minterHeight := minterStatus.LatestBlockHeight

	go func() {
		startTime := time.Now()

		for {
			if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
				panic("Timeout waiting for the refund")
			}

			response, err := ctx.MinterClient.Block(minterHeight)
			if err != nil {
				time.Sleep(time.Millisecond * 200)
				continue
			}

			for _, transactionResponse := range response.Transactions {
				if transactionResponse.From != ctx.MinterMultisig {
					continue
				}

				if transaction.Type(transactionResponse.Type) == transaction.TypeMultisend {
					var data api_pb.MultiSendData
					if err := transactionResponse.Data.UnmarshalTo(&data); err != nil {
						panic(err)
					}

					for _, item := range data.List {
						if item.Coin.Id == minterCoinId && strings.ToLower(ctx.EthAddress.String())[2:] == item.To[2:] && item.Value == expectedValue.String() {
							println("SUCCESS: test refunds with different decimals")
							ctx.TestsWg.Done()
							return
						}
					}
				}
			}

			minterHeight++
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
		expectedValue := sdk.NewInt(1e18)
		expectedValue = expectedValue.Sub(expectedValue.ToDec().Mul(bridgeCommission).TruncateInt())
		expectedValue = expectedValue.SubRaw(fee)
		expectedValue = expectedValue.QuoRaw(1e12)

		startTime := time.Now()

		hubContract, _ := erc20.NewErc20(common.HexToAddress(ctx.Bep20addr), ctx.BscClient)

		for {
			hubBalanceInt, err := hubContract.BalanceOf(nil, common.HexToAddress(recipient))
			if err != nil {
				panic(err)
			}

			hubBalance := sdk.NewIntFromBigInt(hubBalanceInt)
			if hubBalance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
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
		expectedValue = expectedValue.Sub(expectedValue.ToDec().Mul(bridgeCommission).TruncateInt())
		expectedValue = expectedValue.SubRaw(fee)

		startTime := time.Now()

		hubContract, _ := erc20.NewErc20(common.HexToAddress(ctx.Erc20addr), ctx.EthClient)

		for {
			hubBalanceInt, err := hubContract.BalanceOf(nil, common.HexToAddress(recipient))
			if err != nil {
				panic(err)
			}

			hubBalance := sdk.NewIntFromBigInt(hubBalanceInt)
			if hubBalance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
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
		expectedValue = expectedValue.Sub(expectedValue.ToDec().Mul(bridgeCommission).TruncateInt())
		startTime := time.Now()

		for {
			response, err := ctx.MinterClient.Address("Mx" + recipient[2:])
			if err != nil {
				panic(err)
			}

			hubBalance := getMinterCoinBalance(response.Balance, minterCoinId)
			if hubBalance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
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
				if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
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
	expectedValue := sdk.NewInt(1e18)
	sendERC20ToHub(ctx.EthPrivateKey, ctx.BscClient, ctx.BscContract, ctx.Bep20addr, hubRecipient, bscChainId, big.NewInt(1e6))

	go func() {
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
				if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
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
	expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))

	sendERC20ToHub(ctx.EthPrivateKey, ctx.EthClient, ctx.EthContract, ctx.Erc20addr, hubRecipient, ethChainId, expectedValue.BigInt())

	go func() {
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
				if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
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
	sendERC20ToAnotherChain(ctx.EthPrivateKey, ctx.BscClient, ctx.BscContract, ctx.Bep20addr, recipient, bscChainId, "ethereum", big.NewInt(1e6), big.NewInt(100))

	go func() {
		expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))
		expectedValue = expectedValue.Sub(expectedValue.ToDec().Mul(bridgeCommission).TruncateInt())
		expectedValue = expectedValue.SubRaw(100 * 1e12)

		startTime := time.Now()

		hubContract, _ := erc20.NewErc20(common.HexToAddress(ctx.Erc20addr), ctx.EthClient)

		for {
			hubBalanceInt, err := hubContract.BalanceOf(nil, common.HexToAddress(recipient))
			if err != nil {
				panic(err)
			}

			hubBalance := sdk.NewIntFromBigInt(hubBalanceInt)
			if hubBalance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
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

func testEthEthereumToMinterTransfer(ctx *Context) {
	ctx.TestsWg.Add(1)
	randomPk, _ := crypto.GenerateKey()
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	sendEthToAnotherChain(ctx.EthPrivateKey, ctx.EthClient, ctx.EthContract, recipient, ethChainId, "minter", big.NewInt(1e18))

	go func() {
		expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))
		expectedValue = expectedValue.Sub(expectedValue.ToDec().Mul(bridgeCommission).TruncateInt())
		expectedValue = expectedValue.SubRaw(100)
		startTime := time.Now()

		for {
			response, err := ctx.MinterClient.Address("Mx" + recipient[2:])
			if err != nil {
				panic(err)
			}

			hubBalance := getMinterCoinBalance(response.Balance, 2)
			if hubBalance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
					panic("Timeout waiting for the balance to update")
				}

				time.Sleep(time.Second)
				continue
			}

			if !hubBalance.Equal(expectedValue) {
				panic(fmt.Sprintf("Balance is not equal to expected value. Expected %s, got %s", expectedValue.String(), hubBalance.String()))
			}

			println("SUCCESS: test Eth -> Minter (ETH) transfer")
			ctx.TestsWg.Done()
			break
		}
	}()
}

func testEthereumToMinterTransfer(ctx *Context) {
	ctx.TestsWg.Add(1)
	randomPk, _ := crypto.GenerateKey()
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	sendERC20ToAnotherChain(ctx.EthPrivateKey, ctx.EthClient, ctx.EthContract, ctx.Erc20addr, recipient, ethChainId, "minter", big.NewInt(1e18), big.NewInt(100))

	go func() {
		expectedValue := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(1)))
		expectedValue = expectedValue.Sub(expectedValue.ToDec().Mul(bridgeCommission).TruncateInt())
		expectedValue = expectedValue.SubRaw(100)
		startTime := time.Now()

		for {
			response, err := ctx.MinterClient.Address("Mx" + recipient[2:])
			if err != nil {
				panic(err)
			}

			hubBalance := getMinterCoinBalance(response.Balance, minterCoinId)
			if hubBalance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
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

func testValidatorsCommissionsWithDifferentDecimals(ctx *Context) {
	ctx.TestsWg.Add(1)
	randomPk, _ := crypto.GenerateKey()
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	toSend := sdk.NewInt(1e18)
	sendERC20ToAnotherChain(ctx.EthPrivateKey, ctx.BscClient, ctx.BscContract, ctx.Bep20addr, recipient, bscChainId, "minter", toSend.QuoRaw(1e12).BigInt(), big.NewInt(100))

	go func() {
		expectedValue := toSend.ToDec().Mul(bridgeCommission).TruncateInt()
		startTime := time.Now()

		minterStatus, err := ctx.MinterClient.Status()
		if err != nil {
			panic(err)
		}

		minterHeight := minterStatus.LatestBlockHeight

		for {
			if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
				panic("Timeout waiting for the commission to arrive to validator")
			}

			response, err := ctx.MinterClient.Block(minterHeight)
			if err != nil {
				time.Sleep(time.Millisecond * 200)
				continue
			}

			for _, transactionResponse := range response.Transactions {
				if transactionResponse.From != ctx.MinterMultisig {
					continue
				}

				if transaction.Type(transactionResponse.Type) == transaction.TypeMultisend {
					var data api_pb.MultiSendData
					if err := transactionResponse.Data.UnmarshalTo(&data); err != nil {
						panic(err)
					}

					for _, item := range data.List {
						if item.Coin.Id == minterCoinId && strings.ToLower(ctx.EthAddress.String())[2:] == item.To[2:] && item.Value == expectedValue.String() {
							println("SUCCESS: test commission for validator with different decimals")
							ctx.TestsWg.Done()
							return
						}
					}
				}
			}

			minterHeight++
		}
	}()
}

func testValidatorsCommissions(ctx *Context) {
	ctx.TestsWg.Add(1)
	randomPk, _ := crypto.GenerateKey()
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	toSend := transaction.BipToPip(big.NewInt(999))
	sendERC20ToAnotherChain(ctx.EthPrivateKey, ctx.EthClient, ctx.EthContract, ctx.Erc20addr, recipient, ethChainId, "minter", toSend, big.NewInt(100))

	go func() {
		expectedValue := sdk.NewIntFromBigInt(toSend)
		expectedValue = expectedValue.ToDec().Mul(bridgeCommission).TruncateInt()
		startTime := time.Now()

		minterStatus, err := ctx.MinterClient.Status()
		if err != nil {
			panic(err)
		}

		minterHeight := minterStatus.LatestBlockHeight

		for {
			if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
				panic("Timeout waiting for the commission to arrive to validator")
			}

			response, err := ctx.MinterClient.Block(minterHeight)
			if err != nil {
				time.Sleep(time.Millisecond * 200)
				continue
			}

			for _, transactionResponse := range response.Transactions {
				if transactionResponse.From != ctx.MinterMultisig {
					continue
				}

				if transaction.Type(transactionResponse.Type) == transaction.TypeMultisend {
					var data api_pb.MultiSendData
					if err := transactionResponse.Data.UnmarshalTo(&data); err != nil {
						panic(err)
					}

					for _, item := range data.List {
						if item.Coin.Id == minterCoinId && strings.ToLower(ctx.EthAddress.String())[2:] == item.To[2:] && item.Value == expectedValue.String() {
							println("SUCCESS: test commission for validator")
							ctx.TestsWg.Done()
							return
						}
					}
				}
			}

			minterHeight++
		}
	}()
}

func testFeeRefund(ctx *Context) {
	ctx.TestsWg.Add(1)
	randomPk, _ := crypto.GenerateKey()
	randomPkString := fmt.Sprintf("%x", crypto.FromECDSA(randomPk))
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	value := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(16000)))
	fee := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(15000)))
	fundMinterAddress("Mx"+recipient[2:], ctx.EthPrivateKeyString, ctx.EthAddress, ctx.MinterClient)

	sendMinterCoinToEthereum(randomPkString, common.HexToAddress(recipient), ctx.MinterMultisig, ctx.MinterClient, minterCoinId, recipient, value, fee)

	go func() {
		expectedValue := fee // todo: calculate this value somehow
		startTime := time.Now()

		minterStatus, err := ctx.MinterClient.Status()
		if err != nil {
			panic(err)
		}

		minterHeight := minterStatus.LatestBlockHeight

		for {
			if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
				panic("Timeout waiting for the fee refund to arrive")
			}

			response, err := ctx.MinterClient.Block(minterHeight)
			if err != nil {
				time.Sleep(time.Millisecond * 200)
				continue
			}

			for _, transactionResponse := range response.Transactions {
				if transactionResponse.From != ctx.MinterMultisig {
					continue
				}

				if transaction.Type(transactionResponse.Type) == transaction.TypeMultisend {
					var data api_pb.MultiSendData
					if err := transactionResponse.Data.UnmarshalTo(&data); err != nil {
						panic(err)
					}

					for _, item := range data.List {
						if item.Coin.Id != minterCoinId || strings.ToLower(recipient)[2:] != item.To[2:] {
							continue
						}

						itemValue, _ := sdk.NewIntFromString(item.Value)
						if itemValue.LTE(expectedValue) && itemValue.IsPositive() {
							println("SUCCESS: test fee refund")
							ctx.TestsWg.Done()
							return
						}
					}
				}
			}

			minterHeight++
		}
	}()
}

func testFeeRefundWithDifferentDecimals(ctx *Context) {
	ctx.TestsWg.Add(1)
	randomPk, _ := crypto.GenerateKey()
	randomPkString := fmt.Sprintf("%x", crypto.FromECDSA(randomPk))
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	value := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(16000)))
	fee := sdk.NewIntFromBigInt(transaction.BipToPip(big.NewInt(15000)))
	fundMinterAddress("Mx"+recipient[2:], ctx.EthPrivateKeyString, ctx.EthAddress, ctx.MinterClient)

	sendMinterCoinToBsc(randomPkString, common.HexToAddress(recipient), ctx.MinterMultisig, ctx.MinterClient, recipient, value, fee)

	go func() {
		expectedValue := fee // todo: calculate this value somehow
		startTime := time.Now()

		minterStatus, err := ctx.MinterClient.Status()
		if err != nil {
			panic(err)
		}

		minterHeight := minterStatus.LatestBlockHeight

		for {
			if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
				panic("Timeout waiting for the fee refund to arrive")
			}

			response, err := ctx.MinterClient.Block(minterHeight)
			if err != nil {
				time.Sleep(time.Millisecond * 200)
				continue
			}

			for _, transactionResponse := range response.Transactions {
				if transactionResponse.From != ctx.MinterMultisig {
					continue
				}

				if transaction.Type(transactionResponse.Type) == transaction.TypeMultisend {
					var data api_pb.MultiSendData
					if err := transactionResponse.Data.UnmarshalTo(&data); err != nil {
						panic(err)
					}

					for _, item := range data.List {
						if item.Coin.Id != minterCoinId || strings.ToLower(recipient)[2:] != item.To[2:] {
							continue
						}

						itemValue, _ := sdk.NewIntFromString(item.Value)
						if itemValue.LTE(expectedValue) && itemValue.IsPositive() {
							println("SUCCESS: test fee refund with different decimals")
							ctx.TestsWg.Done()
							return
						}
					}
				}
			}

			minterHeight++
		}
	}()
}

func testEthereumToBscTransfer(ctx *Context) {
	ctx.TestsWg.Add(1)
	randomPk, _ := crypto.GenerateKey()
	recipient := crypto.PubkeyToAddress(randomPk.PublicKey).Hex()
	sendERC20ToAnotherChain(ctx.EthPrivateKey, ctx.EthClient, ctx.EthContract, ctx.Erc20addr, recipient, ethChainId, "bsc", big.NewInt(1e18), big.NewInt(100))

	go func() {
		expectedValue := sdk.NewInt(1e18)
		expectedValue = expectedValue.Sub(expectedValue.ToDec().Mul(bridgeCommission).TruncateInt())
		expectedValue = expectedValue.SubRaw(100)
		expectedValue = expectedValue.QuoRaw(1e12)

		startTime := time.Now()

		hubContract, _ := erc20.NewErc20(common.HexToAddress(ctx.Bep20addr), ctx.BscClient)

		for {
			hubBalanceInt, err := hubContract.BalanceOf(nil, common.HexToAddress(recipient))
			if err != nil {
				panic(err)
			}

			hubBalance := sdk.NewIntFromBigInt(hubBalanceInt)
			if hubBalance.IsZero() {
				if time.Now().Sub(startTime).Seconds() > testTimeout.Seconds() {
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
