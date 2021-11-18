package cmd

import (
	"encoding/json"
	"fmt"

	"github.com/MinterTeam/mhub2/module/x/mhub2/types"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

func AddPrepareGenesisForTestsCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prepare-genesis-for-tests [external-addr] [external-addr-2] [external-addr-3]",
		Short: "Prepare genesis for tests",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx := client.GetClientContextFromCmd(cmd)
			depCdc := clientCtx.Codec
			cdc := depCdc.(codec.Codec)

			serverCtx := server.GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			config.SetRoot(clientCtx.HomeDir)

			genFile := config.GenesisFile()
			appState, genDoc, err := genutiltypes.GenesisStateFromGenFile(genFile)
			if err != nil {
				return fmt.Errorf("failed to unmarshal genesis state: %w", err)
			}

			genState := types.GetGenesisStateFromAppState(cdc, appState)

			genState.Params.OutgoingTxTimeout = 60000
			genState.TokenInfos.TokenInfos[0].ExternalTokenId = args[0]
			genState.TokenInfos.TokenInfos[1].ExternalTokenId = args[1]
			genState.TokenInfos.TokenInfos[2].ExternalTokenId = args[2]
			genState.TokenInfos.TokenInfos[1].ExternalDecimals = 6

			genStateJson, err := cdc.MarshalJSON(genState)
			if err != nil {
				return fmt.Errorf("failed to marshal genesis state: %w", err)
			}

			appState[types.ModuleName] = genStateJson

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}
