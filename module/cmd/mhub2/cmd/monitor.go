package cmd

import (
	"context"
	"time"

	"github.com/MinterTeam/mhub2/module/x/mhub2/types"

	"github.com/cosmos/cosmos-sdk/client"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client/flags"
)

// AddMonitorCmd returns monitor cobra Command.
func AddMonitorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitor",
		Short: "",
		Long:  ``,
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ourAddress := args[0]
			nonceErrorCounter := 0

			for {
				clientCtx, err := client.GetClientQueryContext(cmd)
				if err != nil {
					return err
				}

				node, err := clientCtx.GetNode()
				if err != nil {
					return err
				}

				status, err := node.Status(context.Background())
				if err != nil {
					return err
				}

				if time.Now().Sub(status.SyncInfo.LatestBlockTime).Minutes() > 1 {
					panic("Last block on Minter Hub node was more than 1 minute ago")
				}

				queryClient := types.NewQueryClient(clientCtx)

				chains := []types.ChainID{"ethereum", "minter", "bsc"}
			loop:
				for _, chain := range chains {
					delegatedKeys, err := queryClient.DelegateKeys(context.Background(), &types.DelegateKeysRequest{
						ChainId: chain.String(),
					})
					if err != nil {
						return err
					}

					response, err := queryClient.LastSubmittedExternalEvent(context.Background(), &types.LastSubmittedExternalEventRequest{
						Address: ourAddress,
						ChainId: chain.String(),
					})

					if err != nil {
						return err
					}

					nonce := response.EventNonce

					for _, k := range delegatedKeys.GetDelegateKeys() {
						response, err := queryClient.LastSubmittedExternalEvent(context.Background(), &types.LastSubmittedExternalEventRequest{
							Address: k.OrchestratorAddress,
							ChainId: chain.String(),
						})
						if err != nil {
							return err
						}

						if nonce < response.GetEventNonce() {
							nonceErrorCounter++
							break loop
						}
					}

					nonceErrorCounter = 0
				}

				if nonceErrorCounter > 5 {
					panic("Event nonce on some external network was not updated for too long. Check your orchestrators and minter-connector")
				}

				time.Sleep(time.Second * 5)
			}

			return nil
		},
	}

	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
