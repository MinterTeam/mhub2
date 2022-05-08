package minter

import (
	"encoding/json"
	"github.com/MinterTeam/mhub2/minter-connector/command"
	"github.com/MinterTeam/mhub2/minter-connector/context"
	"github.com/MinterTeam/minter-go-sdk/v2/api/http_client"
	"github.com/MinterTeam/minter-go-sdk/v2/api/http_client/models"
	"github.com/MinterTeam/minter-go-sdk/v2/transaction"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	"math"
	"strconv"
	"time"
)

func GetLatestMinterBlock(client *http_client.Client, logger log.Logger) uint64 {
	status, err := client.Status()
	if err != nil {
		logger.Error("Cannot get Minter status", "err", err.Error())
		time.Sleep(1 * time.Second)
		return GetLatestMinterBlock(client, logger)
	}

	return status.LatestBlockHeight
}

func GetLatestMinterBlockAndNonce(ctx context.Context, currentNonce uint64) context.Context {
	ctx.Logger.Info("Current nonce @ hub", "nonce", currentNonce)
	latestBlock := GetLatestMinterBlock(ctx.MinterClient, ctx.Logger)

	firstBlock := ctx.LastCheckedMinterBlock()

	const blocksPerBatch = 100
	for i := uint64(0); i <= uint64(math.Ceil(float64(latestBlock-firstBlock)/blocksPerBatch)); i++ {
		from := firstBlock + 1 + i*blocksPerBatch
		to := firstBlock + (i+1)*blocksPerBatch

		if to > latestBlock {
			to = latestBlock
		}

		blocks, err := ctx.MinterClient.Blocks(from, to, false, false)
		if err != nil {
			ctx.Logger.Error("Error while getting minter blocks", "err", err.Error())
			time.Sleep(time.Second)
			i--
			continue
		}

		ctx.Logger.Debug("Scanning blocks", "from", from, "to", to, "nonce", ctx.LastEventNonce)

		for _, block := range blocks.Blocks {
			// temp fix for missed withdrawal
			if block.Height == 10233000 {
				ctx.Logger.Debug("Applying temp fix", "height", block.Height)
				ctx.Logger.Debug("Found batch")

				if currentNonce > 0 && currentNonce < ctx.LastEventNonce() {
					ctx.SetLastCheckedMinterBlock(block.Height - 1)
					ctx.Commit()
					return ctx
				}

				ctx.SetLastEventNonce(ctx.LastEventNonce() + 1)
				ctx.SetLastBatchNonce(ctx.LastBatchNonce() + 1)
			}

			for _, tx := range block.Transactions {
				if tx.Type == uint64(transaction.TypeSend) {
					data, _ := tx.Data.UnmarshalNew()
					sendData := data.(*models.SendData)
					if sendData.To != ctx.MinterMultisigAddr {
						continue
					}

					cmd := &command.Command{}
					if err := json.Unmarshal(tx.Payload, &cmd); err != nil {
						ctx.Logger.Error("Cannot validate incoming tx", "err", err.Error())
						continue
					}

					value, _ := sdk.NewIntFromString(sendData.Value)
					if cmd.ValidateAndComplete(value) == nil {
						ctx.Logger.Debug("Found deposit")
						if currentNonce > 0 && currentNonce < ctx.LastEventNonce() {
							ctx.SetLastCheckedMinterBlock(block.Height - 1)
							ctx.Commit()
							return ctx
						}

						ctx.SetLastEventNonce(ctx.LastEventNonce() + 1)
					}
				}

				if tx.Type == uint64(transaction.TypeMultisend) && tx.From == ctx.MinterMultisigAddr {
					ctx.Logger.Debug("Found batch")

					if currentNonce > 0 && currentNonce < ctx.LastEventNonce() {
						ctx.SetLastCheckedMinterBlock(block.Height - 1)
						ctx.Commit()
						return ctx
					}

					ctx.SetLastEventNonce(ctx.LastEventNonce() + 1)
					ctx.SetLastBatchNonce(ctx.LastBatchNonce() + 1)
				}

				if tx.Type == uint64(transaction.TypeEditMultisig) && tx.From == ctx.MinterMultisigAddr {
					nonce, err := strconv.Atoi(string(tx.Payload))
					if err != nil {
						ctx.Logger.Error("Error on decoding valset nonce", "err", err.Error())
					} else {
						ctx.Logger.Debug("Found valset update")

						if currentNonce > 0 && currentNonce < ctx.LastEventNonce() {
							ctx.SetLastCheckedMinterBlock(block.Height - 1)
							ctx.Commit()
							return ctx
						}

						ctx.SetLastValsetNonce(uint64(nonce))
						ctx.SetLastEventNonce(ctx.LastEventNonce() + 1)
					}
				}
			}

			ctx.SetLastCheckedMinterBlock(block.Height)
			ctx.Commit()
		}
	}

	return ctx
}
