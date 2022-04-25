package tx_committer

import (
	"context"
	"github.com/MinterTeam/mhub2/module/app"
	sdkTypes "github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/crypto/hd"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	crypto "github.com/cosmos/cosmos-sdk/crypto/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/go-bip39"

	"github.com/cosmos/cosmos-sdk/types/tx/signing"
	signing2 "github.com/cosmos/cosmos-sdk/x/auth/signing"
	"github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/golang/protobuf/proto"
	"github.com/tendermint/tendermint/libs/log"
	tmClient "github.com/tendermint/tendermint/rpc/client/http"
	tmTypes "github.com/tendermint/tendermint/types"
	"google.golang.org/grpc"
	"net"
	"sync"
	"time"
)

var encoding = app.MakeEncodingConfig()

type job struct {
	msg      sdk.Msg
	callback func()
}

type Server struct {
	UnimplementedTxCommitterServer

	cosmosRpcAddr string
	cosmosConn    *grpc.ClientConn
	addr          sdk.AccAddress
	priv          *secp256k1.PrivKey

	jobs []job

	lock   sync.Mutex
	logger log.Logger
}

func (s *Server) CommitTx(_ context.Context, req *CommitTxRequest) (*CommitTxReply, error) {
	msgs, err := UnmarshalMsgs(req.Msgs)
	if err != nil {
		panic(err)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(msgs))

	s.lock.Lock()

	for _, msg := range msgs {

		s.jobs = append(s.jobs, job{
			msg: msg,
			callback: func() {
				println(2)
				wg.Done()
			},
		})
	}

	s.lock.Unlock()
	wg.Wait()

	return nil, nil
}

func (s *Server) run() {
	for {
		time.Sleep(time.Second * 5)
		s.lock.Lock()
		if len(s.jobs) == 0 {
			continue
		}

		var msgs []sdk.Msg
		for _, job := range s.jobs {
			msgs = append(msgs, job.msg)
		}

		SendCosmosTx(s.cosmosRpcAddr, msgs, s.addr, s.priv, s.cosmosConn, s.logger, true)
		for _, job := range s.jobs {
			job.callback()
		}
		s.jobs = []job{}
		s.lock.Unlock()
	}
}

func RunServer(cosmosRpcAddr string, cosmosConn *grpc.ClientConn, cosmosMnemonic string, logger log.Logger) *Server {
	addr, priv := ParseMnemonic(cosmosMnemonic)

	lis, err := net.Listen("tcp", "127.0.0.1:7070")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	txCommitterServer := &Server{
		cosmosRpcAddr: cosmosRpcAddr,
		cosmosConn:    cosmosConn,
		addr:          addr,
		priv:          priv,
		logger:        logger,
	}
	go txCommitterServer.run()
	RegisterTxCommitterServer(s, txCommitterServer)

	go func() {
		if err := s.Serve(lis); err != nil {
			panic(err)
		}
	}()

	return txCommitterServer
}

func ParseMnemonic(mnemonic string) (sdk.AccAddress, *secp256k1.PrivKey) {
	var orcPriv secp256k1.PrivKey
	seed := bip39.NewSeed(mnemonic, "")
	master, ch := hd.ComputeMastersFromSeed(seed)
	orcPriv.Key, _ = hd.DerivePrivateKeyForPath(master, ch, sdk.FullFundraiserPath)
	orcAddress := sdk.AccAddress(orcPriv.PubKey().Address())

	return orcAddress, &orcPriv
}

func MarshalMsgs(msgs []sdk.Msg) [][]byte {
	var result [][]byte
	for _, msg := range msgs {
		bytes, err := encoding.Marshaler.Marshal(sdkTypes.UnsafePackAny(msg))
		if err != nil {
			panic(err)
		}
		result = append(result, bytes)
	}

	return result
}

func UnmarshalMsgs(data [][]byte) ([]sdk.Msg, error) {
	var result []sdk.Msg
	for _, msg := range data {
		anyMsg := new(sdkTypes.Any)
		err := proto.Unmarshal(msg, anyMsg)
		if err != nil {
			return nil, err
		}

		var m sdk.Msg
		err = encoding.InterfaceRegistry.UnpackAny(anyMsg, &m)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}

	return result, nil
}

func SendCosmosTx(cosmosRpcAddr string, msgs []sdk.Msg, address sdk.AccAddress, priv crypto.PrivKey, cosmosConn *grpc.ClientConn, logger log.Logger, retry bool) {
	if len(msgs) > 10 {
		SendCosmosTx(cosmosRpcAddr, msgs[:10], address, priv, cosmosConn, logger, retry)
		SendCosmosTx(cosmosRpcAddr, msgs[10:], address, priv, cosmosConn, logger, retry)
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

	client, err := tmClient.New(cosmosRpcAddr, "")
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
		logger.Error("Failed to broadcast tx", "err", err.Error())

		time.Sleep(5 * time.Second)
		txResponse, err := client.Tx(context.Background(), tmTypes.Tx(txBytes).Hash(), false)
		if (err != nil || txResponse.TxResult.IsErr()) && retry {
			SendCosmosTx(cosmosRpcAddr, msgs, address, priv, cosmosConn, logger, retry)
		}

		return
	}

	if result.DeliverTx.GetCode() != 0 || result.CheckTx.GetCode() != 0 {
		logger.Error("Error on sending cosmos tx with", "code", result.CheckTx.GetCode(), "deliver-code", result.DeliverTx.GetCode(), "log", result.CheckTx.GetLog(), "deliver-log", result.DeliverTx.GetLog())
		if retry {
			time.Sleep(1 * time.Second)
			SendCosmosTx(cosmosRpcAddr, msgs, address, priv, cosmosConn, logger, retry)
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
