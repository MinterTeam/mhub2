package main

import (
	c "context"
	"encoding/hex"
	"fmt"
	"github.com/MinterTeam/mhub2/minter-connector/command"
	"github.com/MinterTeam/mhub2/minter-connector/config"
	"github.com/MinterTeam/mhub2/minter-connector/context"
	"github.com/MinterTeam/mhub2/minter-connector/cosmos"
	"github.com/MinterTeam/mhub2/minter-connector/minter"
	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
	"github.com/MinterTeam/minter-go-sdk/v2/api/http_client"
	"github.com/MinterTeam/minter-go-sdk/v2/api/http_client/models"
	"github.com/MinterTeam/minter-go-sdk/v2/transaction"
	"github.com/MinterTeam/minter-go-sdk/v2/wallet"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const threshold = 667

var cfg = config.Get()

func main() {
	cosmos.Setup()
	orcAddress, orcPriv := cosmos.GetAccount(cfg.Cosmos.Mnemonic)

	pubKey, err := wallet.PublicKeyByPrivateKey(cfg.Minter.PrivateKey)
	if err != nil {
		panic(err)
	}

	address, err := wallet.AddressByPublicKey(pubKey)
	if err != nil {
		panic(err)
	}

	minterWallet := &wallet.Wallet{
		PrivateKey: cfg.Minter.PrivateKey,
		PublicKey:  pubKey,
		Address:    "0x" + address[2:],
	}

	minterClient, err := http_client.New(cfg.Minter.ApiAddr)
	if err != nil {
		panic(err)
	}

	cosmosConn, err := grpc.DialContext(c.Background(), cfg.Cosmos.GrpcAddr, grpc.WithInsecure(), grpc.WithConnectParams(grpc.ConnectParams{
		Backoff:           backoff.DefaultConfig,
		MinConnectTimeout: time.Second * 5,
	}))

	ctx := context.Context{
		MinterMultisigAddr: cfg.Minter.MultisigAddr,
		CosmosConn:         cosmosConn,
		MinterClient:       minterClient,
		OrcAddress:         orcAddress,
		OrcPriv:            orcPriv,
		MinterWallet:       minterWallet,
		Logger:             log.NewTMLogger(os.Stdout),
	}

	ctx.LoadStatus("connector-status.json", cfg.Minter)

	ctx.Logger.Info("Syncing with Minter")

	ctx = minter.GetLatestMinterBlockAndNonce(ctx, cosmos.GetLastMinterNonce(orcAddress.String(), cosmosConn))

	ctx.Logger.Info("Starting with block", "height", ctx.LastCheckedMinterBlock(), "eventNonce", ctx.LastEventNonce(), "batchNonce", ctx.LastBatchNonce(), "valsetNonce", ctx.LastValsetNonce())

	// main loop
	for {
		relayBatches(ctx)
		relayValsets(ctx)
		ctx = relayMinterEvents(ctx)

		ctx.Logger.Info("Last checked minter block", "height", ctx.LastCheckedMinterBlock(), "eventNonce", ctx.LastEventNonce(), "batchNonce", ctx.LastBatchNonce(), "valsetNonce", ctx.LastValsetNonce())
		time.Sleep(2 * time.Second)
	}
}

func relayBatches(ctx context.Context) {
	cosmosClient := types.NewQueryClient(ctx.CosmosConn)

	{
		response, err := cosmosClient.UnsignedBatchTxs(c.Background(), &types.UnsignedBatchTxsRequest{
			Address: ctx.OrcAddress.String(),
			ChainId: "minter",
		})
		if err != nil {
			ctx.Logger.Error("Error while getting last pending batch", "err", err.Error())
			return
		}

		var confirms []sdk.Msg
		for _, batch := range response.GetBatches() {
			txData := transaction.NewMultisendData()
			for _, out := range batch.Transactions {
				txData.AddItem(transaction.NewSendData().SetCoin(parseCoinId(out.Token.ExternalTokenId)).MustSetTo("Mx" + out.ExternalRecipient[2:]).SetValue(out.Token.Amount.BigInt()))
			}

			tx, _ := transaction.NewBuilder(cfg.Minter.ChainID).NewTransaction(txData)
			signedTx, _ := tx.SetNonce(batch.Sequence).SetGasPrice(1).SetGasCoin(0).SetSignatureType(transaction.SignatureTypeMulti).Sign(
				cfg.Minter.MultisigAddr,
				ctx.MinterWallet.PrivateKey,
			)

			sigData, err := signedTx.SingleSignatureData()
			if err != nil {
				panic(err)
			}

			confirmation, err := types.PackConfirmation(&types.BatchTxConfirmation{
				ExternalTokenId: batch.ExternalTokenId,
				BatchNonce:      batch.BatchNonce,
				ExternalSigner:  ctx.MinterWallet.Address,
				Signature:       HexToBytes(sigData),
			})
			if err != nil {
				panic(err)
			}

			confirms = append(confirms, &types.MsgSubmitExternalTxConfirmation{
				Confirmation: confirmation,
				Signer:       ctx.OrcAddress.String(),
				ChainId:      "minter",
			})
		}

		if len(confirms) > 0 {
			ctx.Logger.Info("Sending batch confirms")
			cosmos.SendCosmosTx(confirms, ctx.OrcAddress, ctx.OrcPriv, ctx.CosmosConn, ctx.Logger, false)
		}
	}

	latestBatches, err := cosmosClient.BatchTxs(c.Background(), &types.BatchTxsRequest{
		ChainId: "minter",
	})
	if err != nil {
		ctx.Logger.Error("Error getting last batches", "err", err.Error())
		return
	}

	sort.Slice(latestBatches.Batches, func(i, j int) bool {
		return latestBatches.Batches[i].Sequence > latestBatches.Batches[j].Sequence
	})

	var oldestSignedBatch *types.BatchTx
	var oldestSignatures []*types.BatchTxConfirmation

	for _, batch := range latestBatches.Batches {
		sigs, err := cosmosClient.BatchTxConfirmations(c.Background(), &types.BatchTxConfirmationsRequest{
			BatchNonce:      batch.BatchNonce,
			ExternalTokenId: batch.ExternalTokenId,
			ChainId:         "minter",
		})
		if err != nil {
			ctx.Logger.Error("Error while getting batch confirms", "err", err.Error())
			return
		}

		if sigs.Size() > 0 { // todo: check if we have enough votes
			oldestSignedBatch = batch
			oldestSignatures = sigs.Signatures
		}
	}

	if oldestSignedBatch == nil {
		return
	}

	if oldestSignedBatch.BatchNonce < ctx.LastBatchNonce() {
		return
	}

	ctx.Logger.Info("Sending batch to Minter")

	txData := transaction.NewMultisendData()
	for _, out := range oldestSignedBatch.Transactions {
		txData.AddItem(transaction.NewSendData().SetCoin(parseCoinId(out.Token.ExternalTokenId)).MustSetTo("Mx" + out.ExternalRecipient[2:]).SetValue(out.Token.Amount.BigInt()))
	}

	tx, _ := transaction.NewBuilder(cfg.Minter.ChainID).NewTransaction(txData)
	tx.SetNonce(oldestSignedBatch.Sequence).SetGasPrice(1).SetGasCoin(0).SetSignatureType(transaction.SignatureTypeMulti)
	signedTx, err := tx.Sign(cfg.Minter.MultisigAddr)
	if err != nil {
		panic(err)
	}

	msig, err := ctx.MinterClient.Address(ctx.MinterMultisigAddr)
	if err != nil {
		println(err.Error())
		time.Sleep(time.Second)
		return
	}

	// Check if signer is in the signer set of multisig
	for _, sig := range oldestSignatures {
		for _, member := range msig.Multisig.Addresses {
			if strings.ToLower(member[2:]) == strings.ToLower(sig.ExternalSigner[2:]) {
				signedTx, err = signedTx.AddSignature(fmt.Sprintf("%x", sig.Signature))
				if err != nil {
					panic(err)
				}
			}
		}
	}

	encodedTx, err := signedTx.Encode()
	if err != nil {
		panic(err)
	}

	ctx.Logger.Debug("Batch tx", "tx", encodedTx)
	response, err := ctx.MinterClient.SendTransaction(encodedTx)
	if err != nil {
		code, body, err := http_client.ErrorBody(err)
		if err != nil {
			ctx.Logger.Error("Error on sending Minter Tx", "err", err.Error())
		} else {
			ctx.Logger.Error("Error on sending Minter Tx", "code", code, "err", body.Error.Message)
		}
	} else if response.Code != 0 {
		ctx.Logger.Error("Error on sending Minter Tx", "err", response.Log)
	}
}

func HexToBytes(data string) []byte {
	b, err := hex.DecodeString(data)
	if err != nil {
		panic(err)
	}

	return b
}

func parseCoinId(id string) uint64 {
	i, err := strconv.Atoi(id)
	if err != nil {
		panic(err)
	}

	return uint64(i)
}

func relayValsets(ctx context.Context) {
	cosmosClient := types.NewQueryClient(ctx.CosmosConn)

	{
		response, err := cosmosClient.UnsignedSignerSetTxs(c.Background(), &types.UnsignedSignerSetTxsRequest{
			Address: ctx.OrcAddress.String(),
			ChainId: "minter",
		})
		if err != nil {
			ctx.Logger.Error("Error while getting last pending valset", "err", err.Error())
			return
		}

		var confirms []sdk.Msg
		for _, valset := range response.GetSignerSets() {
			txData := transaction.NewEditMultisigData()
			txData.Threshold = threshold

			totalPower := uint64(0)
			for _, val := range valset.Signers {
				totalPower += val.Power
			}

			for _, val := range valset.Signers {
				var addr transaction.Address
				bytes, _ := wallet.AddressToHex("Mx" + val.ExternalAddress[2:])
				copy(addr[:], bytes)

				weight := uint32(sdk.NewUint(val.Power).MulUint64(1000).QuoUint64(totalPower).Uint64())

				txData.Addresses = append(txData.Addresses, addr)
				txData.Weights = append(txData.Weights, weight)
			}

			tx, _ := transaction.NewBuilder(cfg.Minter.ChainID).NewTransaction(txData)
			tx.SetPayload([]byte(strconv.Itoa(int(valset.Nonce))))
			signedTx, err := tx.SetNonce(valset.Sequence).SetGasPrice(1).SetGasCoin(0).SetSignatureType(transaction.SignatureTypeMulti).Sign(
				cfg.Minter.MultisigAddr,
				ctx.MinterWallet.PrivateKey,
			)
			if err != nil {
				panic(err)
			}

			sigData, err := signedTx.SingleSignatureData()
			if err != nil {
				panic(err)
			}

			confirmation, err := types.PackConfirmation(&types.SignerSetTxConfirmation{
				SignerSetNonce: valset.Nonce,
				ExternalSigner: ctx.MinterWallet.Address,
				Signature:      HexToBytes(sigData),
			})

			confirms = append(confirms, &types.MsgSubmitExternalTxConfirmation{
				Confirmation: confirmation,
				Signer:       ctx.OrcAddress.String(),
				ChainId:      "minter",
			})

			ctx.Logger.Info("Sending valsets confirm", "nonce", valset.Nonce)
		}

		if len(confirms) > 0 {
			cosmos.SendCosmosTx(confirms, ctx.OrcAddress, ctx.OrcPriv, ctx.CosmosConn, ctx.Logger, false)
		}
	}

	latestValsets, err := cosmosClient.SignerSetTxs(c.Background(), &types.SignerSetTxsRequest{
		Pagination: nil,
		ChainId:    "minter",
	})
	if err != nil {
		ctx.Logger.Error("Error on getting last valset requests", "err", err.Error())
		return
	}

	var oldestSignedValset *types.SignerSetTx
	var oldestSignatures []*types.SignerSetTxConfirmation

	valsets := latestValsets.GetSignerSets()
	for i := len(valsets) - 1; i >= 0; i-- {
		valset := valsets[i]
		sigs, err := cosmosClient.SignerSetTxConfirmations(c.Background(), &types.SignerSetTxConfirmationsRequest{
			SignerSetNonce: valset.Nonce,
			ChainId:        "minter",
		})
		if err != nil {
			ctx.Logger.Error("Error while getting valset confirms", "err", err.Error())
			return
		}

		if sigs.Size() > 0 { // todo: check if we have enough votes
			oldestSignedValset = valset
			oldestSignatures = sigs.Signatures

			if oldestSignedValset.Nonce > ctx.LastValsetNonce() {
				break
			}
		}
	}

	if oldestSignedValset == nil {
		return
	}

	if oldestSignedValset.Nonce <= ctx.LastValsetNonce() {
		return
	}

	ctx.Logger.Info("Sending valset to Minter")

	txData := transaction.NewEditMultisigData()
	txData.Threshold = threshold

	totalPower := uint64(0)
	for _, val := range oldestSignedValset.GetSigners() {
		totalPower += val.Power
	}

	for _, val := range oldestSignedValset.GetSigners() {
		var addr transaction.Address
		bytes, _ := wallet.AddressToHex("Mx" + val.ExternalAddress[2:])
		copy(addr[:], bytes)

		weight := uint32(sdk.NewUint(val.Power).MulUint64(1000).QuoUint64(totalPower).Uint64())

		txData.Addresses = append(txData.Addresses, addr)
		txData.Weights = append(txData.Weights, weight)
	}

	tx, _ := transaction.NewBuilder(cfg.Minter.ChainID).NewTransaction(txData)
	tx.SetNonce(oldestSignedValset.Sequence).SetGasPrice(1).SetGasCoin(0).SetSignatureType(transaction.SignatureTypeMulti)
	tx.SetPayload([]byte(strconv.Itoa(int(oldestSignedValset.Nonce))))
	signedTx, err := tx.Sign(cfg.Minter.MultisigAddr)
	if err != nil {
		panic(err)
	}

	msig, err := ctx.MinterClient.Address(ctx.MinterMultisigAddr)
	if err != nil {
		println(err.Error())
		time.Sleep(time.Second)
		return
	}

	// Check if signer is in the last confirmed valset
	for _, sig := range oldestSignatures {
		hasMember := false
		if msig.Multisig != nil {
			for _, member := range msig.Multisig.Addresses {
				if strings.ToLower(member[2:]) == strings.ToLower(sig.ExternalSigner[2:]) {
					hasMember = true
				}
			}
		}

		if hasMember || msig.Multisig == nil {
			signedTx, err = signedTx.AddSignature(fmt.Sprintf("%x", sig.Signature))
			if err != nil {
				panic(err)
			}
		}
	}

	encodedTx, err := signedTx.Encode()
	if err != nil {
		panic(err)
	}

	ctx.Logger.Debug("Valset update tx", "tx", encodedTx)
	response, err := ctx.MinterClient.SendTransaction(encodedTx)
	if err != nil {
		code, body, err := http_client.ErrorBody(err)
		if err != nil {
			ctx.Logger.Error("Error on sending Minter Tx", "err", err.Error())
		} else {
			ctx.Logger.Error("Error on sending Minter Tx", "code", code, "err", body.Error.Message)
		}
	} else if response.Code != 0 {
		ctx.Logger.Error("Error on sending Minter Tx", "err", response.Log)
	}
}

func relayMinterEvents(ctx context.Context) context.Context {
	latestBlock := minter.GetLatestMinterBlock(ctx.MinterClient, ctx.Logger)
	if latestBlock-ctx.LastCheckedMinterBlock() > 100 {
		latestBlock = ctx.LastCheckedMinterBlock() + 100
	}

	var deposits []cosmos.Deposit
	var batches []cosmos.Batch
	var valsets []cosmos.Valset

	const blocksPerBatch = 100
	for i := uint64(0); i <= uint64(math.Ceil(float64(latestBlock-ctx.LastCheckedMinterBlock())/blocksPerBatch)); i++ {
		from := ctx.LastCheckedMinterBlock() + 1 + i*blocksPerBatch
		to := ctx.LastCheckedMinterBlock() + (i+1)*blocksPerBatch

		if to > latestBlock {
			to = latestBlock
		}

		blocks, err := ctx.MinterClient.Blocks(from, to, false, false)
		if err != nil {
			ctx.Logger.Info("Error getting minter blocks", "err", err.Error())
			time.Sleep(time.Second)
			i--
			continue
		}

		for _, block := range blocks.Blocks {
			ctx.SetLastCheckedMinterBlock(block.Height)
			ctx.Logger.Debug("Checking block", "height", block.Height)

			// temp fix for missed withdrawal
			if block.Height == 10233000 {
				ctx.Logger.Debug("Applying temp fix", "height", block.Height)

				batches = append(batches, cosmos.Batch{
					BatchNonce: ctx.LastBatchNonce(),
					EventNonce: ctx.LastEventNonce(),
					CoinId:     2107,
					TxHash:     "Mt46922c03e9d0a40991c3ab54f227bc8c833bf7230d9e1c43a66e00f3abeee163",
					Height:     block.Height,
				})

				ctx.SetLastEventNonce(ctx.LastEventNonce() + 1)
				ctx.SetLastBatchNonce(ctx.LastBatchNonce() + 1)
			}

			for _, tx := range block.Transactions {
				if tx.Type == uint64(transaction.TypeSend) {
					data, _ := tx.Data.UnmarshalNew()
					sendData := data.(*models.SendData)
					if sendData.To != cfg.Minter.MultisigAddr {
						continue
					}

					cmd := &command.Command{}
					if err := json.Unmarshal(tx.Payload, &cmd); err != nil {
						ctx.Logger.Error("Cannot validate incoming tx", "err", err.Error())
						continue
					}

					value, _ := sdk.NewIntFromString(sendData.Value)

					if err := cmd.ValidateAndComplete(value); err != nil {
						ctx.Logger.Error("Cannot validate incoming tx", "err", err.Error())
						continue
					}

					ctx.Logger.Info("Found new deposit", "from", tx.From, "to", string(tx.Payload), "amount", sendData.Value, "coin", sendData.Coin.ID)
					deposits = append(deposits, cosmos.Deposit{
						Sender:     tx.From,
						Type:       cmd.Type,
						Recipient:  cmd.Recipient,
						Amount:     sendData.Value,
						Fee:        cmd.Fee,
						EventNonce: ctx.LastEventNonce(),
						CoinID:     sendData.Coin.ID,
						TxHash:     tx.Hash,
						Height:     block.Height,
					})

					ctx.SetLastEventNonce(ctx.LastEventNonce() + 1)
				}

				if tx.Type == uint64(transaction.TypeMultisend) && tx.From == cfg.Minter.MultisigAddr {
					ctx.Logger.Info("Found withdrawal")
					data, _ := tx.Data.UnmarshalNew()
					multisendData := data.(*models.MultiSendData)

					batches = append(batches, cosmos.Batch{
						BatchNonce: ctx.LastBatchNonce(),
						EventNonce: ctx.LastEventNonce(),
						CoinId:     multisendData.List[0].Coin.ID,
						TxHash:     tx.Hash,
						Height:     block.Height,
					})

					ctx.SetLastEventNonce(ctx.LastEventNonce() + 1)
					ctx.SetLastBatchNonce(ctx.LastBatchNonce() + 1)
				}

				if tx.Type == uint64(transaction.TypeEditMultisig) && tx.From == cfg.Minter.MultisigAddr {
					data, _ := tx.Data.UnmarshalNew()
					editMultisigData := data.(*models.EditMultisigData)

					nonce, err := strconv.Atoi(string(tx.Payload))

					ctx.Logger.Info("Found valset update", "nonce", nonce)
					if err != nil {
						ctx.Logger.Error("Error while decoding valset update nonce", "err", err.Error())
					} else {
						var members []*types.ExternalSigner
						for n := range editMultisigData.Addresses {
							members = append(members, &types.ExternalSigner{
								Power:           editMultisigData.Weights[n],
								ExternalAddress: "0x" + editMultisigData.Addresses[n][2:],
							})
						}

						valsets = append(valsets, cosmos.Valset{
							ValsetNonce: uint64(nonce),
							EventNonce:  ctx.LastEventNonce(),
							Height:      block.Height,
							TxHash:      tx.Hash,
							Members:     members,
						})

						ctx.SetLastEventNonce(ctx.LastEventNonce() + 1)
						ctx.SetLastValsetNonce(uint64(nonce))
					}
				}
			}

			if len(deposits) == 0 && len(batches) == 0 && len(valsets) == 0 {
				ctx.Commit()
			}
		}
	}

	if len(deposits) > 0 || len(batches) > 0 || len(valsets) > 0 {
		cosmos.SendCosmosTx(cosmos.CreateClaims(ctx.OrcAddress, deposits, batches, valsets), ctx.OrcAddress, ctx.OrcPriv, ctx.CosmosConn, ctx.Logger, true)
		ctx.Commit()
	}

	return ctx
}
