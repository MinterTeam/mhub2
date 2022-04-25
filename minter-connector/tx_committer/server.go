package tx_committer

import (
	"context"
	"github.com/MinterTeam/mhub2/minter-connector/cosmos"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/golang/protobuf/proto"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"
	"net"
	"sync"
	"time"
)

type Server struct {
	UnimplementedTxCommitterServer

	cosmosConn *grpc.ClientConn
	addr       sdk.AccAddress
	priv       *secp256k1.PrivKey

	txs []sdk.Msg

	lock   sync.Mutex
	logger log.Logger
}

func (s *Server) CommitTx(_ context.Context, req *CommitTxRequest) (*CommitTxReply, error) {
	s.lock.Lock()
	defer s.lock.Unlock()

	msgs, err := UnmarshalMsgs(req.Msgs)
	if err != nil {
		return nil, err
	}

	s.txs = append(s.txs, msgs...)

	return nil, nil
}

func (s *Server) run() {
	for {
		time.Sleep(time.Second * 5)
		s.lock.Lock()
		cosmos.SendCosmosTx(s.txs, s.addr, s.priv, s.cosmosConn, s.logger, true)
		s.txs = []sdk.Msg{}
		s.lock.Unlock()
	}
}

func RunServer(cosmosConn *grpc.ClientConn, cosmosMnemonic string, logger log.Logger) *Server {
	addr, priv := cosmos.GetAccount(cosmosMnemonic)

	lis, err := net.Listen("unix", "tx_committer_socket")
	if err != nil {
		panic(err)
	}
	s := grpc.NewServer()
	txCommitterServer := &Server{
		cosmosConn: cosmosConn,
		addr:       addr,
		priv:       priv,
		logger:     logger,
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

func MarshalMsgs(msgs []sdk.Msg) [][]byte {
	var result [][]byte
	for _, msg := range msgs {
		bytes, err := proto.Marshal(msg)
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
		var m sdk.Msg
		err := proto.Unmarshal(msg, m)
		if err != nil {
			return nil, err
		}
		result = append(result, m)
	}

	return result, nil
}
