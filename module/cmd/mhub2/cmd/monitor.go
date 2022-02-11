package cmd

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/MinterTeam/mhub2/module/x/mhub2/types"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"
)

// AddMonitorCmd returns monitor cobra Command.
func AddMonitorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor [our_address] [telegram_bot_token] [chat_id]",
		Short: "",
		Long:  ``,
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			bot, err := tgbotapi.NewBotAPI(args[1])
			if err != nil {
				panic(err)
			}

			chatId, err := strconv.Atoi(args[2])
			if err != nil {
				panic(err)
			}

			if _, err := bot.Send(tgbotapi.NewMessage(int64(chatId), "Starting")); err != nil {
				panic(err)
			}

			ourAddress := args[0]
			nonceErrorCounter := 0

			handleErr := func(err error) {
				pc, filename, line, _ := runtime.Caller(1)

				str := fmt.Sprintf("[error] in %s[%s:%d] %v", runtime.FuncForPC(pc).Name(), filename, line, err)
				bot.Send(tgbotapi.NewMessage(int64(chatId), str))
			}

			for {
				time.Sleep(time.Second * 5)

				clientCtx, err := client.GetClientQueryContext(cmd)
				if err != nil {
					handleErr(err)
					continue
				}

				node, err := clientCtx.GetNode()
				if err != nil {
					handleErr(err)
					continue
				}

				status, err := node.Status(context.Background())
				if err != nil {
					handleErr(err)
					continue
				}

				if time.Now().Sub(status.SyncInfo.LatestBlockTime).Minutes() > 1 {
					handleErr(errors.New("last block on Minter Hub node was more than 1 minute ago"))
					continue
				}

				queryClient := types.NewQueryClient(clientCtx)

				chains := []types.ChainID{"ethereum", "minter", "bsc"}
			loop:
				for _, chain := range chains {
					delegatedKeys, err := queryClient.DelegateKeys(context.Background(), &types.DelegateKeysRequest{
						ChainId: chain.String(),
					})
					if err != nil {
						handleErr(err)
						continue
					}

					response, err := queryClient.LastSubmittedExternalEvent(context.Background(), &types.LastSubmittedExternalEventRequest{
						Address: ourAddress,
						ChainId: chain.String(),
					})

					if err != nil {
						handleErr(err)
						continue
					}

					nonce := response.EventNonce

					for _, k := range delegatedKeys.GetDelegateKeys() {
						response, err := queryClient.LastSubmittedExternalEvent(context.Background(), &types.LastSubmittedExternalEventRequest{
							Address: k.OrchestratorAddress,
							ChainId: chain.String(),
						})
						if err != nil {
							if !strings.Contains(err.Error(), "validator is not bonded") {
								handleErr(err)
							}
							continue
						}

						if nonce < response.GetEventNonce() {
							nonceErrorCounter++
							break loop
						}
					}

					nonceErrorCounter = 0
				}

				if nonceErrorCounter > 5 {
					handleErr(errors.New("event nonce on some external network was not updated for too long. Check your orchestrators and minter-connector"))
					continue
				}
			}

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
