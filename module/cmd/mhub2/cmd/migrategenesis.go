package cmd

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	oracletypes "github.com/MinterTeam/mhub2/module/x/oracle/types"

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

			legacyOracleState := &legacyOracle{}
			if err := json.Unmarshal(appState["oracle"], legacyOracleState); err != nil {
				panic(err)
			}

			legacyPeggyState := &legacyPeggy{}
			if err := json.Unmarshal(appState["peggy"], legacyPeggyState); err != nil {
				panic(err)
			}

			var minterDelegatedKeys []*types.MsgDelegateKeys
			var ethDelegatedKeys []*types.MsgDelegateKeys
			var bscDelegatedKeys []*types.MsgDelegateKeys

			for _, item := range legacyMinterState.MinterAddresses {
				valAddress, err := sdk.ValAddressFromBech32(item.Validator)
				if err != nil {
					panic(err)
				}

				minterDelegatedKeys = append(minterDelegatedKeys, &types.MsgDelegateKeys{
					ValidatorAddress:    valAddress.String(),
					OrchestratorAddress: sdk.AccAddress(valAddress).String(),
					ExternalAddress:     "0x" + item.Address[2:],
					EthSignature:        []byte{1},
					ChainId:             "minter",
				})
			}

			for _, item := range legacyPeggyState.OrchestratorAddresses {
				valAddress, err := sdk.ValAddressFromBech32(item.Validator)
				if err != nil {
					panic(err)
				}

				ethDelegatedKeys = append(ethDelegatedKeys, &types.MsgDelegateKeys{
					ValidatorAddress:    valAddress.String(),
					OrchestratorAddress: sdk.AccAddress(valAddress).String(),
					ExternalAddress:     item.EthAddress,
					EthSignature:        []byte{1},
					ChainId:             "ethereum",
				})

				bscDelegatedKeys = append(bscDelegatedKeys, &types.MsgDelegateKeys{
					ValidatorAddress:    valAddress.String(),
					OrchestratorAddress: sdk.AccAddress(valAddress).String(),
					ExternalAddress:     item.EthAddress,
					EthSignature:        []byte{1},
					ChainId:             "bsc",
				})
			}

			startMinterNonce, _ := strconv.Atoi(legacyMinterState.StartMinterNonce)

			defaultGenesis := types.DefaultGenesisState()
			genState := &types.GenesisState{
				Params: defaultGenesis.Params,
				ExternalStates: []*types.ExternalState{
					{
						ChainId:      "minter",
						Sequence:     uint64(startMinterNonce),
						DelegateKeys: minterDelegatedKeys,
					},
					{
						ChainId:      "ethereum",
						DelegateKeys: ethDelegatedKeys,
					},
					{
						ChainId:      "bsc",
						DelegateKeys: bscDelegatedKeys,
					},
				},
				TokenInfos: &types.TokenInfos{
					TokenInfos: nil,
				},
			}

			genState.Params.GravityId = "minter-hub-2"

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
			delete(appState, "minter")
			delete(appState, "peggy")

			oracleGenStateJson, err := cdc.MarshalJSON(oracletypes.DefaultGenesisState())
			if err != nil {
				return fmt.Errorf("failed to marshal genesis state: %w", err)
			}
			appState[oracletypes.ModuleName] = oracleGenStateJson

			appStateJSON, err := json.Marshal(appState)
			if err != nil {
				return fmt.Errorf("failed to marshal application genesis state: %w", err)
			}

			genDoc.AppState = appStateJSON
			genDoc.ChainID = "mhub-mainnet-2"
			genDoc.GenesisTime = time.Date(2021, 12, 2, 13, 0, 0, 0, time.UTC)

			return genutil.ExportGenesisFile(genDoc, genFile)
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	flags.AddQueryFlagsToCmd(cmd)

	return cmd
}

type legacyMinter struct {
	StartMinterNonce string `json:"start_minter_nonce"`
	MinterAddresses  []struct {
		Address   string `json:"address"`
		Validator string `json:"validator"`
	} `json:"minter_addresses"`
}

type legacyPeggy struct {
	OrchestratorAddresses []struct {
		EthAddress   string `json:"eth_address"`
		Orchestrator string `json:"orchestrator"`
		Validator    string `json:"validator"`
	} `json:"orchestrator_addresses"`
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
