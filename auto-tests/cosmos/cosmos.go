package cosmos

import (
	"context"
	"github.com/MinterTeam/mhub2/module/app"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	signing2 "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/go-bip39"
	"github.com/tendermint/tendermint/libs/log"
	tmClient "github.com/tendermint/tendermint/rpc/client/http"
	tmTypes "github.com/tendermint/tendermint/types"
	"google.golang.org/grpc"
	"strings"
	"time"
)

var encoding = app.MakeEncodingConfig()

func SendCosmosTx(msgs []sdk.Msg, address sdk.AccAddress, priv crypto.PrivKey, cosmosConn *grpc.ClientConn, logger log.Logger, retry bool) {
	if len(msgs) > 10 {
		SendCosmosTx(msgs[:10], address, priv, cosmosConn, logger, retry)
		SendCosmosTx(msgs[10:], address, priv, cosmosConn, logger, retry)
		return
	}

	number, sequence := getAccount(address.String(), cosmosConn, logger)

	fee := sdk.NewCoins(sdk.NewCoin("hub", sdk.NewInt(1)))

	tx := encoding.TxConfig.NewTxBuilder()
	err := tx.SetMsgs(msgs...)
	if err != nil {
		panic(err)
	}

	tx.SetMemo("")
	tx.SetFeeAmount(fee)
	tx.SetGasLimit(100000000)

	sigData := signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: nil,
	}
	sig := signing.SignatureV2{
		PubKey:   priv.PubKey(),
		Data:     &sigData,
		Sequence: sequence,
	}

	if err := tx.SetSignatures(sig); err != nil {
		panic(err)
	}

	client, err := tmClient.New("http://localhost:26657", "")
	if err != nil {
		panic(err)
	}

	status, err := client.Status(context.TODO())
	if err != nil {
		panic(err)
	}

	signBytes, err := encoding.TxConfig.SignModeHandler().GetSignBytes(signing.SignMode_SIGN_MODE_DIRECT, signing2.SignerData{
		ChainID:       status.NodeInfo.Network,
		AccountNumber: number,
		Sequence:      sequence,
	}, tx.GetTx())
	if err != nil {
		panic(err)
	}

	// Sign those bytes
	sigBytes, err := priv.Sign(signBytes)
	if err != nil {
		panic(err)
	}

	// Construct the SignatureV2 struct
	sigData = signing.SingleSignatureData{
		SignMode:  signing.SignMode_SIGN_MODE_DIRECT,
		Signature: sigBytes,
	}
	sig = signing.SignatureV2{
		PubKey:   priv.PubKey(),
		Data:     &sigData,
		Sequence: sequence,
	}

	if err := tx.SetSignatures(sig); err != nil {
		panic(err)
	}

	txBytes, err := encoding.TxConfig.TxEncoder()(tx.GetTx())
	if err != nil {
		panic(err)
	}

	result, err := client.BroadcastTxCommit(context.Background(), txBytes)
	if err != nil {
		if !strings.Contains(err.Error(), "incorrect account sequence") {
			logger.Error("Failed broadcast tx", "err", err.Error())
		}

		time.Sleep(5 * time.Second)
		txResponse, err := client.Tx(context.Background(), tmTypes.Tx(txBytes).Hash(), false)
		if (err != nil || txResponse.TxResult.IsErr()) && retry {
			SendCosmosTx(msgs, address, priv, cosmosConn, logger, retry)
		}

		return
	}

	if result.DeliverTx.GetCode() != 0 || result.CheckTx.GetCode() != 0 {
		logger.Error("Error on sending cosmos tx with", "code", result.CheckTx.GetCode(), "deliver-code", result.DeliverTx.GetCode(), "log", result.CheckTx.GetLog(), "deliver-log", result.DeliverTx.GetLog())
		if retry {
			time.Sleep(1 * time.Second)
			SendCosmosTx(msgs, address, priv, cosmosConn, logger, retry)
		}

		return
	}

	logger.Info("Sending cosmos tx", "code", result.DeliverTx.GetCode(), "log", result.DeliverTx.GetLog(), "info", result.DeliverTx.GetInfo())
}

func getAccount(address string, conn *grpc.ClientConn, logger log.Logger) (number, sequence uint64) {
	authClient := types.NewQueryClient(conn)

	response, err := authClient.Account(context.Background(), &types.QueryAccountRequest{Address: address})
	if err != nil {
		logger.Error("Error getting cosmos account", "err", err.Error())
		time.Sleep(1 * time.Second)
		return getAccount(address, conn, logger)
	}

	var account types.AccountI
	if err := encoding.Marshaler.UnpackAny(response.Account, &account); err != nil {
		logger.Error("Error unpacking cosmos account", "err", err.Error())
		time.Sleep(1 * time.Second)
		return getAccount(address, conn, logger)
	}

	return account.GetAccountNumber(), account.GetSequence()
}

func GetAccount(mnemonic string) (sdk.AccAddress, *secp256k1.PrivKey) {
	var orcPriv secp256k1.PrivKey
	seed := bip39.NewSeed(mnemonic, "")
	master, ch := hd.ComputeMastersFromSeed(seed)
	orcPriv.Key, _ = hd.DerivePrivateKeyForPath(master, ch, sdk.FullFundraiserPath)
	orcAddress := sdk.AccAddress(orcPriv.PubKey().Address())

	return orcAddress, &orcPriv
}
