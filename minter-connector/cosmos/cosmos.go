package cosmos

import (
	"context"
	"fmt"
	"github.com/MinterTeam/mhub2/minter-connector/command"
	"github.com/MinterTeam/mhub2/minter-connector/config"
	"github.com/MinterTeam/mhub2/module/app"
	mhub "github.com/MinterTeam/mhub2/module/x/mhub2/types"
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
	"sort"
	"strings"
	"time"
)

var encoding = app.MakeEncodingConfig()
var cfg = config.Get()

type Batch struct {
	BatchNonce uint64
	EventNonce uint64
	CoinId     uint64
	TxHash     string
	Height     uint64
}

type Valset struct {
	ValsetNonce uint64
	EventNonce  uint64
	TxHash      string
	Height      uint64
	Members     []*mhub.ExternalSigner
}

type Deposit struct {
	Recipient  string
	Amount     string
	Fee        string
	EventNonce uint64
	Sender     string
	CoinID     uint64
	Type       string
	TxHash     string
	Height     uint64
}

func Setup() {
	cosmosConfig := sdk.GetConfig()
	cosmosConfig.SetBech32PrefixForAccount("hub", "hubpub")
	cosmosConfig.Seal()
}

func CreateClaims(orcAddress sdk.AccAddress, deposits []Deposit, batches []Batch, valsets []Valset) []sdk.Msg {
	var msgs []sdk.Msg
	for _, deposit := range deposits {
		amount, _ := sdk.NewIntFromString(deposit.Amount)
		fee, _ := sdk.NewIntFromString(deposit.Fee)

		switch deposit.Type {
		case command.TypeSendToEth, command.TypeSendToBsc:
			receiverChain := "ethereum"
			if deposit.Type == command.TypeSendToBsc {
				receiverChain = "bsc"
			}

			event, err := mhub.PackEvent(&mhub.TransferToChainEvent{
				EventNonce:       deposit.EventNonce,
				ExternalCoinId:   fmt.Sprintf("%d", deposit.CoinID),
				Amount:           amount,
				Fee:              fee,
				Sender:           "0x" + deposit.Sender[2:],
				ReceiverChainId:  receiverChain,
				ExternalReceiver: deposit.Recipient,
				ExternalHeight:   deposit.Height,
				TxHash:           deposit.TxHash,
			})
			if err != nil {
				panic(err)
			}
			msgs = append(msgs, &mhub.MsgSubmitExternalEvent{
				Event:   event,
				Signer:  orcAddress.String(),
				ChainId: "minter",
			})
		case command.TypeSendToHub:
			event, err := mhub.PackEvent(&mhub.SendToHubEvent{
				EventNonce:     deposit.EventNonce,
				ExternalCoinId: fmt.Sprintf("%d", deposit.CoinID),
				Amount:         amount,
				Sender:         "0x" + deposit.Sender[2:],
				CosmosReceiver: deposit.Recipient,
				ExternalHeight: deposit.Height,
				TxHash:         deposit.TxHash,
			})
			if err != nil {
				panic(err)
			}
			msgs = append(msgs, &mhub.MsgSubmitExternalEvent{
				Event:   event,
				Signer:  orcAddress.String(),
				ChainId: "minter",
			})
		default:
			panic("unknown event")
		}
	}

	for _, batch := range batches {
		event, err := mhub.PackEvent(&mhub.BatchExecutedEvent{
			ExternalCoinId: fmt.Sprintf("%d", batch.CoinId),
			EventNonce:     batch.EventNonce,
			ExternalHeight: batch.Height,
			BatchNonce:     batch.BatchNonce,
			TxHash:         batch.TxHash,
		})
		if err != nil {
			panic(err)
		}
		msgs = append(msgs, &mhub.MsgSubmitExternalEvent{
			Event:   event,
			Signer:  orcAddress.String(),
			ChainId: "minter",
		})
	}

	for _, valset := range valsets {
		event, err := mhub.PackEvent(&mhub.SignerSetTxExecutedEvent{
			EventNonce:       valset.EventNonce,
			SignerSetTxNonce: valset.ValsetNonce,
			ExternalHeight:   valset.Height,
			Members:          valset.Members,
			TxHash:           valset.TxHash,
		})
		if err != nil {
			panic(err)
		}
		msgs = append(msgs, &mhub.MsgSubmitExternalEvent{
			Event:   event,
			Signer:  orcAddress.String(),
			ChainId: "minter",
		})
	}

	sort.Slice(msgs, func(i, j int) bool {
		return getEventNonceFromMsg(msgs[i]) < getEventNonceFromMsg(msgs[j])
	})

	return msgs
}

func getEventNonceFromMsg(msg sdk.Msg) uint64 {
	switch m := msg.(type) {
	case *mhub.MsgSubmitExternalEvent:
		event, err := mhub.UnpackEvent(m.Event)
		if err != nil {
			panic(err)
		}

		return event.GetEventNonce()
	}

	return 999999999
}

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

	client, err := tmClient.New(cfg.Cosmos.RpcAddr, "")
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
		if result.CheckTx.GetCode() != 32 {
			logger.Error("Error on sending cosmos tx with", "code", result.CheckTx.GetCode(), "deliver-code", result.DeliverTx.GetCode(), "log", result.CheckTx.GetLog(), "deliver-log", result.DeliverTx.GetLog())
		}

		if retry {
			time.Sleep(1 * time.Second)
			SendCosmosTx(msgs, address, priv, cosmosConn, logger, retry)
		}

		return
	}

	logger.Info("Sending cosmos tx", "code", result.DeliverTx.GetCode(), "log", result.DeliverTx.GetLog(), "info", result.DeliverTx.GetInfo())
}

func GetLastMinterNonce(address string, conn *grpc.ClientConn) uint64 {
	client := mhub.NewQueryClient(conn)

	result, err := client.LastSubmittedExternalEvent(context.Background(), &mhub.LastSubmittedExternalEventRequest{Address: address, ChainId: "minter"})
	if err != nil {
		panic(err)
	}

	return result.EventNonce
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
