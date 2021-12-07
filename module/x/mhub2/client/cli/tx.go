package cli

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/MinterTeam/mhub2/module/x/mhub2/client/utils"
	"github.com/cosmos/cosmos-sdk/version"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/tx"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/spf13/cobra"

	"github.com/MinterTeam/mhub2/module/x/mhub2/types"
)

func GetTxCmd(storeKey string) *cobra.Command {
	txCmd := &cobra.Command{
		Use:                        types.ModuleName,
		Short:                      "Mhub2 transaction subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	txCmd.AddCommand(
		CmdSendToEthereum(),
		CmdCancelSendToEthereum(),
		CmdRequestBatchTx(),
		CmdSetDelegateKeys(),
	)

	return txCmd
}

func CmdSendToEthereum() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "send-to-ethereum [ethereum-reciever] [send-coins] [fee-coins]",
		Aliases: []string{"send", "transfer"},
		Args:    cobra.ExactArgs(3),
		Short:   "Send tokens from cosmos chain to connected ethereum chain",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			if from == nil {
				return fmt.Errorf("must pass from flag")
			}

			if !common.IsHexAddress(args[0]) {
				return fmt.Errorf("must be a valid ethereum address got %s", args[0])
			}

			// Get amount of coins
			sendCoin, err := sdk.ParseCoinNormalized(args[1])
			if err != nil {
				return err
			}

			feeCoin, err := sdk.ParseCoinNormalized(args[2])
			if err != nil {
				return err
			}

			msg := types.NewMsgSendToExternal(from, common.HexToAddress(args[0]).Hex(), sendCoin, feeCoin)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdCancelSendToEthereum() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-send-to-external [chain-id] [id]",
		Args:  cobra.ExactArgs(2),
		Short: "Cancel ethereum send by id",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			if from == nil {
				return fmt.Errorf("must pass from flag")
			}

			id, err := strconv.ParseUint(args[1], 10, 64)
			if err != nil {
				return err
			}

			msg := types.NewMsgCancelSendToExternal(id, types.ChainID(args[0]), from)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdRequestBatchTx() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "request-batch-tx [denom] [signer]",
		Args:  cobra.ExactArgs(2),
		Short: "Request batch transaction for denom by signer",
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			denom := args[0]
			signer, err := sdk.AccAddressFromHex(args[1])
			if err != nil {
				return err
			}

			msg := types.NewMsgRequestBatchTx(denom, signer)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}
			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func CmdSetDelegateKeys() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-delegate-keys [chain-id] [validator-address] [orchestrator-address] [ethereum-address] [ethereum-signature]",
		Args:  cobra.ExactArgs(5),
		Short: "Set mhub2 delegate keys",
		Long: `Set a validator's Ethereum and orchestrator addresses. The validator must
sign over a binary Proto-encoded DelegateKeysSignMsg message. The message contains
the validator's address and operator account current nonce.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}

			chainId, err := parseChainId(args[0])
			if err != nil {
				panic(err)
			}

			valAddr, err := sdk.ValAddressFromBech32(args[1])
			if err != nil {
				return err
			}

			orcAddr, err := sdk.AccAddressFromBech32(args[2])
			if err != nil {
				return err
			}

			ethAddr := args[3]

			ethSig, err := hexutil.Decode(args[4])
			if err != nil {
				return err
			}

			msg := types.NewMsgDelegateKeys(valAddr, types.ChainID(chainId), orcAddr, ethAddr, ethSig)
			if err = msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}

	flags.AddTxFlagsToCmd(cmd)
	return cmd
}

func NewSubmitColdStorageTransferProposalTxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "cold-storage-transfer [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a cold storage transfer proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a cold storage transfer proposal along with an initial deposit.
The proposal details must be supplied via a JSON file. For values that contains
objects, only non-empty fields will be updated.

Example:
$ %s tx gov submit-proposal cold-storage-transfer <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:

{
  "amount": [
    {
      "denom": "hub",
      "amount": "100"
    }
  ],
  "chain_id": "minter",
  "deposit": "1000hub"
}
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			proposal, err := utils.ParseColdStorageTransferProposalJSON(clientCtx.LegacyAmino, args[0])
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			content := types.NewColdStorageTransferProposal(
				types.ChainID(proposal.ChainId),
				proposal.Amount,
			)

			deposit, err := sdk.ParseCoinsNormalized(proposal.Deposit)
			if err != nil {
				return err
			}

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}

func NewSubmitTokenInfosChangeProposalTxCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "token-infos-change [proposal-file]",
		Args:  cobra.ExactArgs(1),
		Short: "Submit a token infos change proposal",
		Long: strings.TrimSpace(
			fmt.Sprintf(`Submit a token infos change proposal along with an initial deposit.
The proposal details must be supplied via a JSON file. For values that contains
objects, only non-empty fields will be updated.

Example:
$ %s tx gov submit-proposal token-infos-change <path/to/proposal.json> --from=<key_or_address>

Where proposal.json contains:

{
  "new_token_infos": ...,
  "deposit": "1000hub"
}
`,
				version.AppName,
			),
		),
		RunE: func(cmd *cobra.Command, args []string) error {
			clientCtx, err := client.GetClientTxContext(cmd)
			if err != nil {
				return err
			}
			proposal, err := utils.ParseTokenInfosChangeProposalJSON(clientCtx.LegacyAmino, args[0])
			if err != nil {
				return err
			}

			from := clientCtx.GetFromAddress()
			content := types.NewTokenInfosChangeProposal(
				proposal.NewTokenInfos,
			)

			deposit, err := sdk.ParseCoinsNormalized(proposal.Deposit)
			if err != nil {
				return err
			}

			msg, err := govtypes.NewMsgSubmitProposal(content, deposit, from)
			if err != nil {
				return err
			}
			if err := msg.ValidateBasic(); err != nil {
				return err
			}

			return tx.GenerateOrBroadcastTxCLI(clientCtx, cmd.Flags(), msg)
		},
	}
}
