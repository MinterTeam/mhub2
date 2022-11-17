package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/big"
	"os"
	"runtime"
	"sort"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"

	"github.com/ethereum/go-ethereum/common"

	"github.com/ethereum/go-ethereum/ethclient"

	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"

	"github.com/MinterTeam/mhub2/module/solidity"
	"github.com/MinterTeam/mhub2/module/x/mhub2/types"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"
)

type MonitorConfig struct {
	OurAddress    string   `json:"our_address"`
	TelegramToken string   `json:"telegram_token"`
	ChatID        int64    `json:"chat_id"`
	EthereumRPC   []string `json:"ethereum_rpc"`
	BNBChainRPC   []string `json:"bnb_chain_rpc"`
}

const blockDelay = 6

// AddMonitorCmd returns monitor cobra Command.
func AddMonitorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor [config]",
		Short: "",
		Long:  ``,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			config, err := readConfig(args[0])
			if err != nil {
				return err
			}

			bot, err := tgbotapi.NewBotAPI(config.TelegramToken)
			if err != nil {
				panic(err)
			}

			newText := func(t string) string {
				return fmt.Sprintf("%s\n%s", time.Now().Format(time.Stamp), t)
			}

			chat, err := bot.GetChat(tgbotapi.ChatInfoConfig{
				ChatConfig: tgbotapi.ChatConfig{
					ChatID: config.ChatID,
				},
			})
			if err != nil {
				panic(err)
			}

			var startMsg tgbotapi.Message

			if chat.PinnedMessage != nil {
				startMsg = *chat.PinnedMessage
			} else {
				msg := tgbotapi.NewMessage(config.ChatID, newText(""))
				msg.DisableNotification = true
				startMsg, err = bot.Send(msg)
				if err != nil {
					panic(err)
				}

				bot.Request(tgbotapi.PinChatMessageConfig{
					ChatID:              startMsg.Chat.ID,
					MessageID:           startMsg.MessageID,
					DisableNotification: true,
				})
			}

			nonceErrorCounter := make(map[types.ChainID]int)
			handleErr := func(err error) {
				pc, filename, line, _ := runtime.Caller(1)

				str := fmt.Sprintf("[error] in %s[%s:%d] %v", runtime.FuncForPC(pc).Name(), filename, line, err)
				if _, err := bot.Send(tgbotapi.NewMessage(config.ChatID, str)); err != nil {
					println(err.Error())
				}

				msg := tgbotapi.NewMessage(config.ChatID, newText(""))
				msg.DisableNotification = true
				startMsg, _ = bot.Send(msg)
				bot.Request(tgbotapi.PinChatMessageConfig{
					ChatID:              startMsg.Chat.ID,
					MessageID:           startMsg.MessageID,
					DisableNotification: true,
				})
			}

			initialized := false

			for {
				if initialized {
					time.Sleep(time.Minute)
				}

				initialized = true
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

				valHasFailure := map[string]bool{}
				valHasNonceFailure := map[string]map[types.ChainID]uint64{}
				failuresLog := ""

				for _, val := range vals.GetValidators() {
					valHasNonceFailure[val.OperatorAddress] = map[types.ChainID]uint64{}
				}

				actualNonces := map[types.ChainID]uint64{}

				chains := []types.ChainID{"ethereum", "minter", "bsc"}
				for _, chain := range chains {
					delegatedKeys, err := queryClient.DelegateKeys(context.Background(), &types.DelegateKeysRequest{
						ChainId: chain.String(),
					})
					if err != nil {
						handleErr(err)
						continue
					}

					actualNonce, err := getActualNonce(chain, config, delegatedKeys.GetDelegateKeys(), queryClient)
					if err != nil {
						handleErr(err)
						continue
					}

					actualNonces[chain] = actualNonce

					if config.OurAddress != "" {
						response, err := queryClient.LastSubmittedExternalEvent(context.Background(), &types.LastSubmittedExternalEventRequest{
							Address: config.OurAddress,
							ChainId: chain.String(),
						})

						if err != nil {
							handleErr(err)
							continue
						}

						ourNonce := response.EventNonce
						if ourNonce < actualNonce {
							nonceErrorCounter[chain]++
						} else {
							nonceErrorCounter[chain] = 0
						}
					}

					for _, k := range delegatedKeys.GetDelegateKeys() {
						response, err := queryClient.LastSubmittedExternalEvent(context.Background(), &types.LastSubmittedExternalEventRequest{
							Address: k.OrchestratorAddress,
							ChainId: chain.String(),
						})
						if err != nil {
							if !strings.Contains(err.Error(), "validator is not bonded") &&
								!strings.Contains(err.Error(), "not orchestrator or validator") {
								handleErr(err)
							}
							continue
						}

						for _, v := range vals.GetValidators() {
							if v.OperatorAddress == k.ValidatorAddress {
								nonce := response.GetEventNonce()
								if nonce < actualNonce {
									valHasNonceFailure[v.OperatorAddress][chain] = nonce
									valHasFailure[v.OperatorAddress] = true
								}
							}
						}
					}

				}

				sortedVals := vals.GetValidators()
				sort.Slice(sortedVals, func(i, j int) bool {
					return sortedVals[i].BondedTokens().GT(sortedVals[j].BondedTokens())
				})

				for _, v := range sortedVals {
					alert := "🟢"
					if v.IsJailed() {
						alert = "⚫️"
					}
					if valHasFailure[v.OperatorAddress] {
						alert = fmt.Sprintf("🔴️")
						failuresLog = fmt.Sprintf("%s⚠️️ <b>%s</b> ", failuresLog, v.GetMoniker())
					}
					t = fmt.Sprintf("%s\n%s <b>%s</b> %d HUB", t, alert, v.GetMoniker(), v.BondedTokens().QuoRaw(1e18).Int64())

					if valHasFailure[v.OperatorAddress] {
						var nonceErrs []string
						for _, chain := range chains {
							if nonce, ok := valHasNonceFailure[v.OperatorAddress][chain]; ok {
								nonceErrs = append(nonceErrs, fmt.Sprintf("nonce <b>%d</b> of <b>%d</b> on <b>%s</b>", nonce, actualNonces[chain], chain.String()))
							}
						}

						failuresLog += strings.Join(nonceErrs, ", ") + "\n"
					}
				}

				for _, chain := range chains {
					if nonceErrorCounter[chain] > 5 {
						handleErr(errors.New("event nonce on " + chain.String() + " was not updated for too long"))
						continue
					}
				}

				if failuresLog != "" {
					t = t + "\n\n" + failuresLog
				}

				msg := tgbotapi.NewEditMessageText(startMsg.Chat.ID, startMsg.MessageID, newText(t))
				msg.ParseMode = "html"
				if _, err := bot.Send(msg); err != nil {
					println(err.Error())
				}
			}

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

func getActualNonce(chain types.ChainID, config MonitorConfig, keys []*types.MsgDelegateKeys, queryClient types.QueryClient) (uint64, error) {
	switch chain {
	case "ethereum", "bsc":
		var address common.Address
		var RPCs []string

		switch chain {
		case "ethereum":
			address = common.HexToAddress("0x897c27fa372aa730d4c75b1243e7ea38879194e2")
			RPCs = config.EthereumRPC
		case "bsc":
			address = common.HexToAddress("0xf5b0ed82a0b3e11567081694cc66c3df133f7c8f")
			RPCs = config.BNBChainRPC
		}

		maxNonce, err := getEvmNonce(address, RPCs)
		if err != nil {
			return 0, err
		}

		if maxNonce == 0 {
			return 0, errors.New("no available nonce source for " + chain.String())
		}

		return maxNonce, nil
	case "minter":
		maxNonce := uint64(0)
		for _, k := range keys {
			response, err := queryClient.LastSubmittedExternalEvent(context.Background(), &types.LastSubmittedExternalEventRequest{
				Address: k.OrchestratorAddress,
				ChainId: chain.String(),
			})
			if err != nil {
				if !strings.Contains(err.Error(), "validator is not bonded") {
					return 0, err
				}
			}

			if maxNonce < response.GetEventNonce() {
				maxNonce = response.GetEventNonce()
			}
		}

		return maxNonce, nil
	}

	return 0, nil
}

func getEvmNonce(address common.Address, RPCs []string) (uint64, error) {
	maxNonce := uint64(0)
	for _, rpc := range RPCs {
		evmClient, err := ethclient.Dial(rpc)
		if err != nil {
			continue
		}

		instance, err := solidity.NewHub2(address, evmClient)
		if err != nil {
			continue
		}

		latestBlock, err := evmClient.BlockNumber(context.TODO())
		if err != nil {
			continue
		}

		lastNonce, err := instance.StateLastEventNonce(&bind.CallOpts{
			BlockNumber: big.NewInt(int64(latestBlock - blockDelay)),
		})
		if err != nil {
			continue
		}

		if maxNonce < lastNonce.Uint64() {
			maxNonce = lastNonce.Uint64()
		}
	}

	return maxNonce, nil
}

func readConfig(path string) (MonitorConfig, error) {
	config := MonitorConfig{}
	configBody, err := os.ReadFile(path)
	if err != nil {
		return MonitorConfig{}, err
	}

	if err := json.Unmarshal(configBody, &config); err != nil {
		return MonitorConfig{}, err
	}

	return config, nil
}
