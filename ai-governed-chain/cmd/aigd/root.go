package main

import (
	"os"

	"github.com/spf13/cobra"

	tmcfg "github.com/tendermint/tendermint/config"
	tmcli "github.com/tendermint/tendermint/libs/cli"

	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/client/config"
	"github.com/cosmos/cosmos-sdk/client/debug"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/client/keys"
	"github.com/cosmos/cosmos-sdk/client/rpc"
	"github.com/cosmos/cosmos-sdk/server"

	authcli "github.com/cosmos/cosmos-sdk/x/auth/client/cli"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	genutilcli "github.com/cosmos/cosmos-sdk/x/genutil/client/cli"

	aiapp "github.com/donovansolms/cosmos-experiments/ai-governed-chain/app"
	"github.com/donovansolms/cosmos-experiments/ai-governed-chain/x/gov"
)

//------------------------------------------------------------------------------
// Constructor for aigd root command
//------------------------------------------------------------------------------

// NewRootCmd creates root command for the AI governed app-chain daemon
func NewRootCmd(encodingConfig aiapp.EncodingConfig) *cobra.Command {
	initClientCtx := client.Context{}.
		WithCodec(encodingConfig.Codec).
		WithInterfaceRegistry(encodingConfig.InterfaceRegistry).
		WithTxConfig(encodingConfig.TxConfig).
		WithLegacyAmino(encodingConfig.Amino).
		WithInput(os.Stdin).
		WithAccountRetriever(authtypes.AccountRetriever{}).
		WithBroadcastMode(flags.BroadcastBlock).
		WithHomeDir(aiapp.DefaultNodeHome).
		WithViper("AIG")

	// **** create root command ****

	rootCmd := &cobra.Command{
		Use:   "aigd",
		Short: "AI governed app-chain daemon",
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			// set the default command outputs
			cmd.SetOut(cmd.OutOrStdout())
			cmd.SetErr(cmd.ErrOrStderr())

			initClientCtx, err := client.ReadPersistentCommandFlags(initClientCtx, cmd.Flags())
			if err != nil {
				return err
			}

			initClientCtx, err = config.ReadFromClientConfig(initClientCtx)
			if err != nil {
				return err
			}

			if err := client.SetCmdClientContextHandler(initClientCtx, cmd); err != nil {
				return err
			}

			customAppTemplate, customAppConfig := initAppConfig()

			return server.InterceptConfigsPreRunHandler(cmd, customAppTemplate, customAppConfig, tmcfg.DefaultConfig())
		},
		SilenceUsage: true,
	}

	// **** add subcommands ****

	ac := appCreator{encodingConfig}
	server.AddCommands(
		rootCmd,
		aiapp.DefaultNodeHome,
		ac.createApp,
		ac.exportApp,
		func(startCmd *cobra.Command) {
			crisis.AddModuleInitFlags(startCmd)
			gov.AddModuleInitFlags(startCmd)
		},
	)

	rootCmd.AddCommand(
		genesisCommand(encodingConfig),
		queryCommand(),
		txCommand(),
		genutilcli.InitCmd(aiapp.ModuleBasics, aiapp.DefaultNodeHome),
		tmcli.NewCompletionCmd(rootCmd, true),
		config.Cmd(),
		debug.Cmd(),
		keys.Commands(aiapp.DefaultNodeHome),
		rpc.StatusCommand(),
	)

	return rootCmd
}

func genesisCommand(encodingConfig aiapp.EncodingConfig) *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "genesis",
		Short:                      "Utilities for preparing the genesis state",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		genutilcli.CollectGenTxsCmd(banktypes.GenesisBalancesIterator{}, aiapp.DefaultNodeHome),
		genutilcli.MigrateGenesisCmd(),
		genutilcli.GenTxCmd(
			aiapp.ModuleBasics,
			encodingConfig.TxConfig,
			banktypes.GenesisBalancesIterator{},
			aiapp.DefaultNodeHome,
		),
		genutilcli.ValidateGenesisCmd(aiapp.ModuleBasics),
		addGenesisAccountCmd(aiapp.DefaultNodeHome),
		addGenesisWasmMsgCmd(aiapp.DefaultNodeHome),
	)

	return cmd
}

func queryCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "query",
		Aliases:                    []string{"q"},
		Short:                      "Querying subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		rpc.BlockCommand(),
		rpc.ValidatorCommand(),
		authcli.QueryTxCmd(),
		authcli.QueryTxsByEventsCmd(),
	)

	aiapp.ModuleBasics.AddQueryCommands(cmd)

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}

func txCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                        "tx",
		Short:                      "Transactions subcommands",
		DisableFlagParsing:         true,
		SuggestionsMinimumDistance: 2,
		RunE:                       client.ValidateCmd,
	}

	cmd.AddCommand(
		authcli.GetSignCommand(),
		authcli.GetSignBatchCommand(),
		authcli.GetMultiSignCommand(),
		authcli.GetValidateSignaturesCommand(),
		authcli.GetBroadcastCommand(),
		authcli.GetEncodeCommand(),
		authcli.GetDecodeCommand(),
	)

	aiapp.ModuleBasics.AddTxCommands(cmd)

	cmd.PersistentFlags().String(flags.FlagChainID, "", "The network chain ID")

	return cmd
}
