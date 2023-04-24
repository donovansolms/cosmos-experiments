package gov

// DONTCOVER

import (
	"github.com/spf13/cobra"

	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"

	keeper "github.com/donovansolms/cosmos-experiments/ai-governed-validator/x/gov/keeper"
)

// Module init related flags
const (
	FlagOpenAIKey = "ai.openai_key"
	FlagAIRules   = "ai.rules"
)

// AppModule must implement the `module.AppModule` interface
var _ module.AppModule = AppModule{}

// AppModule implements an application module for the custom gov module
//
// NOTE: our custom AppModule wraps the vanilla `gov.AppModule` to inherit most
// of its functions. However, we overwrite the `EndBlock` function to replace it
// with our custom vote tallying logic.
type AppModule struct {
	gov.AppModule

	keeper        keeper.Keeper
	accountKeeper govtypes.AccountKeeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, keeper keeper.Keeper, ak govtypes.AccountKeeper, bk govtypes.BankKeeper) AppModule {
	return AppModule{
		AppModule:     gov.NewAppModule(cdc, keeper.Keeper, ak, bk),
		keeper:        keeper,
		accountKeeper: ak,
	}
}

// RegisterServices registers module services.
//
// NOTE: this overwrites the vanilla gov module RegisterServices function
func (am AppModule) RegisterServices(cfg module.Configurator) {
	macc := am.accountKeeper.GetModuleAddress(govtypes.ModuleName).String()

	// msg server - use the vanilla implementation
	// The changes we've made to execution are in EndBlocker, so the msgServer
	// doesn't need to be changed.
	msgServer := keeper.NewMsgServerImpl(am.keeper)
	govv1beta1.RegisterMsgServer(cfg.MsgServer(), keeper.NewLegacyMsgServerImpl(macc, msgServer))
	govv1.RegisterMsgServer(cfg.MsgServer(), msgServer)

	// query server - use our custom implementation
	queryServer := keeper.NewQueryServerImpl(am.keeper)
	govv1beta1.RegisterQueryServer(cfg.QueryServer(), keeper.NewLegacyQueryServerImpl(queryServer))
	govv1.RegisterQueryServer(cfg.QueryServer(), queryServer)
}

// AddModuleInitFlags implements servertypes.ModuleInitFlags interface.
func AddModuleInitFlags(startCmd *cobra.Command) {
	startCmd.Flags().String(FlagOpenAIKey, "", "Set the OpenAI API key")
	startCmd.Flags().String(FlagAIRules, "", "Set the AI rules for this validator")
}
