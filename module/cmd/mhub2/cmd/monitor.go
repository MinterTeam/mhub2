package cmd

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"strconv"
	"strings"
	"time"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

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

			newText := func(t string) string {
				return fmt.Sprintf("Watching...\n%s\n%s", time.Now().Format(time.Stamp), t)
			}

			startMsg, err := bot.Send(tgbotapi.NewMessage(int64(chatId), newText("")))
			if err != nil {
				panic(err)
			}

			ourAddress := args[0]
			nonceErrorCounter := 0
			hasNonceError := false

			handleErr := func(err error) {
				pc, filename, line, _ := runtime.Caller(1)

				str := fmt.Sprintf("[error] in %s[%s:%d] %v", runtime.FuncForPC(pc).Name(), filename, line, err)
				if _, err := bot.Send(tgbotapi.NewMessage(int64(chatId), str)); err != nil {
					println(err.Error())
				}

				startMsg, _ = bot.Send(tgbotapi.NewMessage(int64(chatId), newText("")))
			}

			i := 0
			for {
				time.Sleep(time.Second * 5)

				t := ""

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

				if time.Now().Sub(status.SyncInfo.LatestBlockTime).Seconds() > 30 {
					handleErr(errors.New("last block on Minter Hub node was more than 30 seconds ago"))
					continue
				}

				queryClient := types.NewQueryClient(clientCtx)
				stakingQueryClient := stakingtypes.NewQueryClient(clientCtx)
				vals, _ := stakingQueryClient.Validators(context.Background(), &stakingtypes.QueryValidatorsRequest{})

				chains := []types.ChainID{"ethereum", "minter", "bsc"}
				for _, chain := range chains {
					t = fmt.Sprintf("%s\n\n<b>%s</b>", t, chain.String())

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

						for _, v := range vals.GetValidators() {
							if v.OperatorAddress == k.ValidatorAddress {
								t = fmt.Sprintf("%s\n%s\t%d", t, v.GetMoniker(), response.GetEventNonce())
							}
						}

						if nonce < response.GetEventNonce() {
							hasNonceError = true
						}
					}

					if !hasNonceError {
						nonceErrorCounter = 0
					}
				}

				if hasNonceError {
					nonceErrorCounter += 1
				}
				hasNonceError = false

				if nonceErrorCounter > 5 {
					handleErr(errors.New("event nonce on some external network was not updated for too long. Check your orchestrators and minter-connector"))
					continue
				}

				i++
				if i%12 == 0 {
					msg := tgbotapi.NewEditMessageText(startMsg.Chat.ID, startMsg.MessageID, newText(t))
					msg.ParseMode = "html"
					_, err := bot.Send(msg)
					if err != nil {
						println(err.Error())
					}
				}
			}

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
