package main

import (
	"context"
	"crypto/ecdsa"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/MinterTeam/mhub2/auto-tests/erc20"
	"github.com/MinterTeam/mhub2/auto-tests/help"
	"github.com/MinterTeam/mhub2/auto-tests/hub2"
	minterTypes "github.com/MinterTeam/minter-go-node/coreV2/types"
	"github.com/MinterTeam/minter-go-sdk/v2/api/http_client"
	"github.com/MinterTeam/minter-go-sdk/v2/api/http_client/models"
	"github.com/MinterTeam/minter-go-sdk/v2/transaction"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core"
	ethTypes "github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ethereum/go-ethereum/params"
	"github.com/tendermint/tendermint/libs/json"
	tmTypes "github.com/tendermint/tendermint/types"
	"io/ioutil"
	"math"
	"math/big"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

func getMinterCoinBalance(balances []*models.AddressBalance, coin string) sdk.Int {
	for _, balance := range balances {
		if balance.Coin.Symbol == coin {
			b, _ := sdk.NewIntFromString(balance.Value)
			return b
		}
	}

	return sdk.NewInt(0)
}

func sendMinterCoinToHub(privateKeyString string, sender common.Address, multisig string, client *http_client.Client, to string) {
	addr := "Mx" + sender.Hex()[2:]
	tx, _ := transaction.NewBuilder(transaction.TestNetChainID).NewTransaction(
		transaction.NewSendData().MustSetTo(multisig).SetCoin(1).SetValue(transaction.BipToPip(big.NewInt(1))),
	)

	nonce, err := client.Nonce(addr)
	if err != nil {
		panic(err)
	}
	signedTransaction, _ := tx.SetNonce(nonce).SetPayload([]byte("{\"recipient\":\"" + to + "\",\"type\":\"send_to_hub\",\"fee\":\"0\"}")).Sign(privateKeyString)

	encode, _ := signedTransaction.Encode()

	response, err := client.SendTransaction(encode)
	if err != nil {
		panic(err)
	}

	if response.Code != 0 {
		panic(response.Log)
	}
}

func sendMinterCoinToEthereum(privateKeyString string, sender common.Address, multisig string, client *http_client.Client, to string) {
	addr := "Mx" + sender.Hex()[2:]
	tx, _ := transaction.NewBuilder(transaction.TestNetChainID).NewTransaction(
		transaction.NewSendData().MustSetTo(multisig).SetCoin(1).SetValue(transaction.BipToPip(big.NewInt(1))),
	)

	nonce, err := client.Nonce(addr)
	if err != nil {
		panic(err)
	}
	signedTransaction, _ := tx.SetNonce(nonce).SetPayload([]byte("{\"recipient\":\"" + to + "\",\"type\":\"send_to_eth\",\"fee\":\"100\"}")).Sign(privateKeyString)

	encode, _ := signedTransaction.Encode()

	response, err := client.SendTransaction(encode)
	if err != nil {
		panic(err)
	}

	if response.Code != 0 {
		panic(response.Log)
	}
}

func fundMinterAddress(to string, privateKeyString string, sender common.Address, client *http_client.Client) {
	addr := "Mx" + sender.Hex()[2:]
	tx, _ := transaction.NewBuilder(transaction.TestNetChainID).NewTransaction(
		transaction.NewMultisendData().
			AddItem(transaction.NewSendData().MustSetTo(to).SetCoin(0).SetValue(transaction.BipToPip(big.NewInt(100000)))).
			AddItem(transaction.NewSendData().MustSetTo(to).SetCoin(1).SetValue(transaction.BipToPip(big.NewInt(100000)))),
	)

	nonce, err := client.Nonce(addr)
	if err != nil {
		panic(err)
	}
	signedTransaction, _ := tx.SetNonce(nonce).Sign(privateKeyString)

	encode, _ := signedTransaction.Encode()

	response, err := client.SendTransaction(encode)
	if err != nil {
		panic(err)
	}

	if response.Code != 0 {
		panic(response.Log)
	}
}

func createMinterMultisig(prKey string, ethAddress common.Address, client *http_client.Client) string {
	addr := "Mx" + ethAddress.Hex()[2:]
	tx, _ := transaction.NewBuilder(transaction.TestNetChainID).NewTransaction(
		transaction.NewCreateMultisigData().SetThreshold(1).MustAddSigData(addr, 1),
	)

	nonce, err := client.Nonce(addr)
	if err != nil {
		panic(err)
	}
	signedTransaction, _ := tx.SetNonce(nonce).Sign(prKey)

	encode, _ := signedTransaction.Encode()

	response, err := client.SendTransaction(encode)
	if err != nil {
		panic(err)
	}

	if response.Code != 0 {
		panic(response.Log)
	}

	for {
		if _, err := client.Transaction(response.Hash); err != nil {
			time.Sleep(time.Millisecond * 200)
			continue
		}

		break
	}

	return transaction.MultisigAddress(addr, nonce)
}

func runEvmChain(chainId int, wd string, ethAddress common.Address, privateKey *ecdsa.PrivateKey, port string) (*ethclient.Client, error, string, string) {
	runOrPanic("geth --networkid %d --datadir %s/data/eth-%d init data/eth-genesis-%d.json", chainId, wd, chainId, chainId)
	go runOrPanic("geth --port %d --maxpeers 0 --http --http.port %s --networkid %d --miner.threads=1 --mine --miner.etherbase %s --datadir %s/data/eth-%d", 30303+chainId, port, chainId, ethAddress.Hex(), wd, chainId)
	client, err := ethclient.Dial("http://localhost:" + port)

	for {
		if _, err := client.ChainID(context.TODO()); err != nil {
			time.Sleep(time.Millisecond * 200)
			continue
		}

		break
	}

	// deploy contract
	contract := deployContract(privateKey, client, int64(chainId))

	// deploy erc20
	erc20addr := deployERC20(privateKey, client, int64(chainId))

	// fund contract with erc20
	fundContractWithErc20(privateKey, client, contract, erc20addr, int64(chainId))

	return client, err, contract, erc20addr
}

func deployContract(privateKey *ecdsa.PrivateKey, client *ethclient.Client, chainId int64) string {
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.PendingNonceAt(context.TODO(), addr)
	if err != nil {
		panic(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.TODO())
	if err != nil {
		panic(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainId))
	if err != nil {
		panic(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	var (
		gravityId      = [32]byte{}
		powerThreshold = big.NewInt(66)
		validators     = []common.Address{
			addr,
		}
		powers = []*big.Int{
			big.NewInt(math.MaxUint32),
		}
	)

	copy(gravityId[:], "defaultgravityid")
	address, tx, _, err := hub2.DeployHub2(auth, client, gravityId, powerThreshold, validators, powers)
	if err != nil {
		panic(err)
	}

	waitEthTx(tx.Hash(), client)

	return address.Hex()
}

func deployERC20(privateKey *ecdsa.PrivateKey, client *ethclient.Client, chainId int64) string {
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.PendingNonceAt(context.TODO(), addr)
	if err != nil {
		panic(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.TODO())
	if err != nil {
		panic(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainId))
	if err != nil {
		panic(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	address, tx, _, err := erc20.DeployErc20(auth, client, addr, "HUB", "HUB", 18)
	if err != nil {
		panic(err)
	}

	waitEthTx(tx.Hash(), client)

	return address.Hex()
}

func sendERC20ToAnotherChain(privateKey *ecdsa.PrivateKey, client *ethclient.Client, hub2Addr string, erc20addr string, to string, chainId int64, destChain string, value *big.Int) {
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.NonceAt(context.TODO(), addr, nil)
	if err != nil {
		panic(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.TODO())
	if err != nil {
		panic(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainId))
	if err != nil {
		panic(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	hub2Instance, err := hub2.NewHub2(common.HexToAddress(hub2Addr), client)
	if err != nil {
		panic(err)
	}

	recipient, err := hex.DecodeString(to[2:])
	if err != nil {
		panic(err)
	}

	rec := [32]byte{}
	copy(rec[12:], recipient)

	destinationChain := [32]byte{}
	copy(destinationChain[:], destChain)

	response, err := hub2Instance.TransferToChain(auth, common.HexToAddress(erc20addr), destinationChain, rec, value, big.NewInt(100))
	if err != nil {
		panic(err)
	}

	waitEthTx(response.Hash(), client)
}

func approveERC20ToHub(privateKey *ecdsa.PrivateKey, client *ethclient.Client, hub2Addr string, erc20addr string, chainId int64) {
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.NonceAt(context.TODO(), addr, nil)
	if err != nil {
		panic(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.TODO())
	if err != nil {
		panic(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainId))
	if err != nil {
		panic(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	erc20Instance, err := erc20.NewErc20(common.HexToAddress(erc20addr), client)
	if err != nil {
		panic(err)
	}

	response, err := erc20Instance.Approve(auth, common.HexToAddress(hub2Addr), abi.MaxUint256)
	if err != nil {
		panic(err)
	}
	waitEthTx(response.Hash(), client)
}

func sendERC20ToHub(privateKey *ecdsa.PrivateKey, client *ethclient.Client, hub2Addr string, erc20addr string, to string, chainId int64, value *big.Int) {
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.NonceAt(context.TODO(), addr, nil)
	if err != nil {
		panic(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.TODO())
	if err != nil {
		panic(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainId))
	if err != nil {
		panic(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	hub2Instance, err := hub2.NewHub2(common.HexToAddress(hub2Addr), client)
	if err != nil {
		panic(err)
	}

	recipient, err := sdk.GetFromBech32(to, "hub")
	if err != nil {
		panic(err)
	}

	rec := [32]byte{}
	copy(rec[12:], recipient)

	destinationChain := [32]byte{}
	copy(destinationChain[:], "hub")

	response, err := hub2Instance.TransferToChain(auth, common.HexToAddress(erc20addr), destinationChain, rec, value, big.NewInt(0))
	if err != nil {
		panic(err)
	}

	waitEthTx(response.Hash(), client)
}

func fundContractWithErc20(privateKey *ecdsa.PrivateKey, client *ethclient.Client, hub2Addr string, erc20addr string, chainId int64) {
	addr := crypto.PubkeyToAddress(privateKey.PublicKey)
	nonce, err := client.PendingNonceAt(context.TODO(), addr)
	if err != nil {
		panic(err)
	}

	gasPrice, err := client.SuggestGasPrice(context.TODO())
	if err != nil {
		panic(err)
	}

	auth, err := bind.NewKeyedTransactorWithChainID(privateKey, big.NewInt(chainId))
	if err != nil {
		panic(err)
	}
	auth.Nonce = big.NewInt(int64(nonce))
	auth.Value = big.NewInt(0)
	auth.GasLimit = uint64(3000000)
	auth.GasPrice = gasPrice

	erc20Instance, err := erc20.NewErc20(common.HexToAddress(erc20addr), client)
	if err != nil {
		panic(err)
	}

	response, err := erc20Instance.Transfer(auth, common.HexToAddress(hub2Addr), transaction.BipToPip(big.NewInt(100)))
	if err != nil {
		panic(err)
	}

	waitEthTx(response.Hash(), client)
}

func waitEthTx(hash common.Hash, client *ethclient.Client) {
	for {
		_, pending, err := client.TransactionByHash(context.TODO(), hash)
		if err != nil {
			panic(err)
		}

		if !pending {
			receipt, err := client.TransactionReceipt(context.TODO(), hash)
			if err != nil {
				panic(err)
			}

			if receipt.Status == ethTypes.ReceiptStatusFailed {
				panic("ethereum tx failed")
			}

			break
		}

		time.Sleep(time.Millisecond * 200)
	}
}

func runOrPanic(cmdString string, args ...interface{}) string {
	out, err := run(cmdString, args...)
	if err != nil {
		panic(err)
	}

	return out
}

func run(cmdString string, args ...interface{}) (string, error) {
	cmdArgs := strings.Split(cmdString, " ")
	for i, arg := range cmdArgs {
		if count := strings.Count(arg, "%"); count > 0 {
			cmdArgs[i] = fmt.Sprintf(arg, args[:count]...)
			args = args[count:]
		}
	}

	argsHash := md5.Sum([]byte(strings.Join(cmdArgs, " ")))
	logFile := fmt.Sprintf("./data/logs/%s-%x.log", cmdArgs[0], argsHash[:5])
	println("$", strings.Join(cmdArgs, " "))
	println("logs:", logFile)
	buffer := help.NewVBuffer(logFile)
	cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
	cmd.Stdout = buffer
	cmd.Stderr = buffer
	//cmd.Env = append(os.Environ(), "RUST_LOG=debug")

	if err := cmd.Run(); err != nil {
		return "", err
	}

	return buffer.String(), nil
}

func getWd() string {
	wd, err := filepath.Abs("./")
	if err != nil {
		panic(err)
	}

	return wd
}

func populateEthGenesis(address common.Address, chainId int) {
	genesis := &core.Genesis{
		Timestamp:  uint64(time.Now().Unix()),
		GasLimit:   4700000,
		Difficulty: big.NewInt(1),
		Alloc: core.GenesisAlloc{
			address: core.GenesisAccount{
				Balance: big.NewInt(1e18),
				Nonce:   0,
			},
		},
		Config: &params.ChainConfig{
			ChainID:             big.NewInt(int64(chainId)),
			HomesteadBlock:      big.NewInt(0),
			EIP150Block:         big.NewInt(0),
			EIP155Block:         big.NewInt(0),
			EIP158Block:         big.NewInt(0),
			ByzantiumBlock:      big.NewInt(0),
			ConstantinopleBlock: big.NewInt(0),
			PetersburgBlock:     big.NewInt(0),
			IstanbulBlock:       big.NewInt(0),
			Ethash:              &params.EthashConfig{},
		},
		ExtraData: make([]byte, 32),
	}

	jsonGenesis, err := genesis.MarshalJSON()
	if err != nil {
		panic(err)
	}

	if err := ioutil.WriteFile(fmt.Sprintf("data/eth-genesis-%d.json", chainId), jsonGenesis, os.FileMode(777)); err != nil {
		panic(err)
	}
}

func populateMinterGenesis(pubkey minterTypes.Pubkey, address common.Address) {
	state := minterTypes.AppState{
		Note: "",
		Validators: []minterTypes.Validator{
			{
				TotalBipStake: "1000000000000000000000000",
				PubKey:        pubkey,
				AccumReward:   "0",
				AbsentTimes:   minterTypes.NewBitArray(24),
			},
		},
		Candidates: []minterTypes.Candidate{
			{
				ID:             1,
				RewardAddress:  minterTypes.Address(address),
				OwnerAddress:   minterTypes.Address(address),
				ControlAddress: minterTypes.Address(address),
				TotalBipStake:  "1000000000000000000000000",
				PubKey:         pubkey,
				Commission:     100,
				Stakes: []minterTypes.Stake{
					{
						Owner:    minterTypes.Address(address),
						Coin:     0,
						Value:    "1000000000000000000000000",
						BipValue: "1000000000000000000000000",
					},
				},
				Status: 2,
			},
		},
		Accounts: []minterTypes.Account{
			{
				Address: minterTypes.Address(address),
				Balance: []minterTypes.Balance{
					{
						Coin:  0,
						Value: "1000000000000000000000000",
					},
					{
						Coin:  1,
						Value: "1000000000000000000000000",
					},
				},
				Nonce: 0,
			},
		},
		Coins: []minterTypes.Coin{
			{
				ID:        1,
				Name:      "HUB",
				Symbol:    minterTypes.StrToCoinSymbol("HUB"),
				Volume:    "1000000000000000000000000",
				Crr:       100,
				Reserve:   "1000000000000000000000000",
				MaxSupply: "1000000000000000000000000",
				Version:   0,
			},
		},
		Commission: minterTypes.Commission{
			Coin:                    0,
			PayloadByte:             "1",
			Send:                    "1",
			BuyBancor:               "1",
			SellBancor:              "1",
			SellAllBancor:           "1",
			BuyPoolBase:             "1",
			BuyPoolDelta:            "1",
			SellPoolBase:            "1",
			SellPoolDelta:           "1",
			SellAllPoolBase:         "1",
			SellAllPoolDelta:        "1",
			CreateTicker3:           "1",
			CreateTicker4:           "1",
			CreateTicker5:           "1",
			CreateTicker6:           "1",
			CreateTicker7_10:        "1",
			CreateCoin:              "1",
			CreateToken:             "1",
			RecreateCoin:            "1",
			RecreateToken:           "1",
			DeclareCandidacy:        "1",
			Delegate:                "1",
			Unbond:                  "1",
			RedeemCheck:             "1",
			SetCandidateOn:          "1",
			SetCandidateOff:         "1",
			CreateMultisig:          "1",
			MultisendBase:           "1",
			MultisendDelta:          "1",
			EditCandidate:           "1",
			SetHaltBlock:            "1",
			EditTickerOwner:         "1",
			EditMultisig:            "1",
			EditCandidatePublicKey:  "1",
			CreateSwapPool:          "1",
			AddLiquidity:            "1",
			RemoveLiquidity:         "1",
			EditCandidateCommission: "1",
			MintToken:               "1",
			BurnToken:               "1",
			VoteCommission:          "1",
			VoteUpdate:              "1",
			FailedTx:                "1",
			AddLimitOrder:           "1",
			RemoveLimitOrder:        "1",
		},
		MaxGas:       100000,
		TotalSlashed: "0",
		Version:      "v250",
	}

	if err := state.Verify(); err != nil {
		panic(err)
	}

	jsonState, err := json.Marshal(state)
	if err != nil {
		panic(err)
	}

	genesis := tmTypes.GenesisDoc{
		GenesisTime:   time.Now(),
		ChainID:       "local-testnet",
		InitialHeight: 1,
		AppState:      jsonState,
	}

	if err := genesis.ValidateAndComplete(); err != nil {
		panic(err)
	}

	if err := genesis.SaveAs("data/minter/config/genesis.json"); err != nil {
		panic(err)
	}
}

func deployContractsAndMultisig(prKeyString string) {
	b, _ := hex.DecodeString(prKeyString)
	pk := crypto.ToECDSAUnsafe(b)
	ethPrivateKeyString := fmt.Sprintf("%x", crypto.FromECDSA(pk))
	ethAddress := crypto.PubkeyToAddress(pk.PublicKey)

	{
		client, err := ethclient.Dial("https://ropsten.infura.io/v3/c2f9dc16ef8d4897a1bd6f9270e38914")
		if err != nil {
			panic(err)
		}

		contract := deployContract(pk, client, 3)

		println("eth", contract)
	}

	{
		client, err := ethclient.Dial("https://data-seed-prebsc-1-s1.binance.org:8545/")
		if err != nil {
			panic(err)
		}

		contract := deployContract(pk, client, 97)

		println("bsc", contract)
	}

	minterClient, _ := http_client.New("https://node-api.taconet.minter.network/v2")
	minterMultisig := createMinterMultisig(ethPrivateKeyString, ethAddress, minterClient)

	println("minter", minterMultisig)
}
