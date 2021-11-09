package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/MinterTeam/mhub2/module/x/mhub2/types"

	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
)

func AddMigrateGenesisCmd(defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "migrate-genesis",
		Short: "",
		Args:  cobra.ExactArgs(0),
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

			legacyMinterState := &legacyMinter{}
			if err := json.Unmarshal(appState["minter"], legacyMinterState); err != nil {
				panic(err)
			}
			startMinterNonce, _ := strconv.Atoi(legacyMinterState.StartMinterNonce)

			legacyOracleState := &legacyOracle{}
			if err := json.Unmarshal(appState["oracle"], legacyOracleState); err != nil {
				panic(err)
			}

			defaultGenesis := types.DefaultGenesisState()
			genState := &types.GenesisState{
				Params: defaultGenesis.Params,
				ExternalStates: []*types.ExternalState{
					{
						ChainId:  "minter",
						Sequence: uint64(startMinterNonce),
					},
				},
				TokenInfos: &types.TokenInfos{
					TokenInfos: nil,
				},
			}

			id := uint64(0)
			for _, coin := range legacyOracleState.Params.Coins {
				id++
				genState.TokenInfos.TokenInfos = append(genState.TokenInfos.TokenInfos, &types.TokenInfo{
					Id:               id,
					Denom:            coin.Denom,
					ChainId:          "minter",
					ExternalTokenId:  coin.MinterId,
					ExternalDecimals: 18,
					Commission:       sdk.MustNewDecFromStr(coin.CustomCommission),
				})

				id++
				ethDecimals, _ := strconv.Atoi(coin.EthDecimals)
				genState.TokenInfos.TokenInfos = append(genState.TokenInfos.TokenInfos, &types.TokenInfo{
					Id:               id,
					Denom:            coin.Denom,
					ChainId:          "ethereum",
					ExternalTokenId:  coin.EthAddr,
					ExternalDecimals: uint64(ethDecimals),
					Commission:       sdk.MustNewDecFromStr(coin.CustomCommission),
				})
			}

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

type legacyMinter struct {
	StartMinterNonce string `json:"start_minter_nonce"`
}

type legacyOracle struct {
	Params struct {
		Coins []struct {
			CustomCommission string `json:"custom_commission"`
			Denom            string `json:"denom"`
			EthAddr          string `json:"eth_addr"`
			EthDecimals      string `json:"eth_decimals"`
			MinterId         string `json:"minter_id"`
		} `json:"coins"`
	} `json:"params"`
}
