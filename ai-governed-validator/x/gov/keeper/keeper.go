package keeper

import (
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	customtypes "github.com/donovansolms/cosmos-experiments/ai-governed-validator/x/gov/customtypes"
)

// Keeper defines the governance module Keeper
type Keeper struct {
	govkeeper.Keeper

	cdc        codec.BinaryCodec
	bankKeeper govtypes.BankKeeper
	storeKey   storetypes.StoreKey

	config customtypes.Config
}

// NewKeeper returns a governance keeper. It handles:
// - submitting governance proposals
// - depositing funds into proposals, and activating upon sufficient funds being deposited
// - users voting on proposals, with weight proportional to stake in the system
// - and tallying the result of the vote.
//
// CONTRACT: the parameter Subspace must have the param key table already initialized
func NewKeeper(
	cdc codec.BinaryCodec, key storetypes.StoreKey, paramSpace govtypes.ParamSubspace,
	accountKeeper govtypes.AccountKeeper, bankKeeper govtypes.BankKeeper, stakingKeeper govtypes.StakingKeeper,
	legacyRouter govv1beta1.Router, router *baseapp.MsgServiceRouter,
	config customtypes.Config,
) Keeper {

	govConfig := govtypes.Config{
		MaxMetadataLen: 4096,
	}

	return Keeper{
		Keeper:     govkeeper.NewKeeper(cdc, key, paramSpace, accountKeeper, bankKeeper, stakingKeeper, legacyRouter, router, govConfig),
		bankKeeper: bankKeeper,
		cdc:        cdc,
		storeKey:   key,
		config:     config,
	}
}
