package app

import (
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cast"

	// tendermint
	abci "github.com/tendermint/tendermint/abci/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"

	// cosmos SDK
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	tmservice "github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	api "github.com/cosmos/cosmos-sdk/server/api"
	config "github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"

	// core modules
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	// customized core modules
	customgov "github.com/donovansolms/cosmos-experiments/ai-governed-chain/x/gov"
	customgovtypes "github.com/donovansolms/cosmos-experiments/ai-governed-chain/x/gov/customtypes"
	customgovkeeper "github.com/donovansolms/cosmos-experiments/ai-governed-chain/x/gov/keeper"

	// ibc modules
	ibctransfer "github.com/cosmos/ibc-go/v6/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v6/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v6/modules/core"
	ibcporttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v6/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v6/modules/core/keeper"
)

const (
	AccountAddressPrefix = "aig"
	Name                 = "aig"

	// BondDenom is the staking token's denomination
	BondDenom = "uaig"
)

var (
	// DefaultNodeHome the default home directory for the app daemon
	DefaultNodeHome string

	// ModuleBasics defines the module `BasicManager`, which is in charge of
	// setting up basic, non- dependent module elements, such as codec
	// registration and genesis verification
	ModuleBasics = module.NewBasicManager(
		auth.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		bank.AppModuleBasic{},
		capability.AppModuleBasic{},
		crisis.AppModuleBasic{},
		distr.AppModuleBasic{}, // distr AppModuleBasic is not customized, so we just use the vanilla one
		evidence.AppModuleBasic{},
		feegrantmodule.AppModuleBasic{},
		genutil.AppModuleBasic{},
		gov.NewAppModuleBasic(
			[]govclient.ProposalHandler{paramsclient.ProposalHandler, distrclient.ProposalHandler, upgradeclient.LegacyProposalHandler, upgradeclient.LegacyCancelProposalHandler},
		),
		params.AppModuleBasic{},
		slashing.AppModuleBasic{},
		staking.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		ibc.AppModuleBasic{},
		ibctransfer.AppModuleBasic{},
	)

	// module account permissions
	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		govtypes.ModuleName:            {authtypes.Burner},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		ibctransfertypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
	}
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	DefaultNodeHome = filepath.Join(userHomeDir, ".aig")
}

//------------------------------------------------------------------------------
// AI governed app
//------------------------------------------------------------------------------

// `AIApp` must implement `simapp.App` and `servertypes.Application` interfaces
var (
	_ simapp.App              = (*AIApp)(nil)
	_ servertypes.Application = (*AIApp)(nil)
)

// AIApp extends an ABCI application, but with most of its parameters exported.
// They are exported for convenience in creating helper functions, as object
// capabilities aren't needed for testing.
type AIApp struct {
	// baseapp
	*baseapp.BaseApp

	// codecs
	legacyAmino       *codec.LegacyAmino
	Codec             codec.Codec // make codec public for testing purposes
	interfaceRegistry codectypes.InterfaceRegistry

	// invariant check period
	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper     authkeeper.AccountKeeper
	AuthzKeeper       authzkeeper.Keeper
	BankKeeper        bankkeeper.Keeper
	CapabilityKeeper  *capabilitykeeper.Keeper
	CrisisKeeper      crisiskeeper.Keeper
	DistrKeeper       distrkeeper.Keeper
	EvidenceKeeper    evidencekeeper.Keeper
	FeeGrantKeeper    feegrantkeeper.Keeper
	GovKeeper         customgovkeeper.Keeper // replaces the vanilla gov keeper with our custom one
	ParamsKeeper      paramskeeper.Keeper
	SlashingKeeper    slashingkeeper.Keeper
	StakingKeeper     stakingkeeper.Keeper
	UpgradeKeeper     upgradekeeper.Keeper
	IBCKeeper         *ibckeeper.Keeper // must be a pointer, so we can `SetRouter` on it correctly
	IBCTransferKeeper ibctransferkeeper.Keeper

	// make scoped keepers public for testing purposes
	ScopedIBCKeeper         capabilitykeeper.ScopedKeeper
	ScopedIBCTransferKeeper capabilitykeeper.ScopedKeeper

	// the module manager
	mm *module.Manager
}

// NewAIApp creates and initializes a new `AIApp` instance
func NewAIApp(
	logger log.Logger, db dbm.DB, traceStore io.Writer, loadLatest bool, skipUpgradeHeights map[int64]bool,
	homePath string, invCheckPeriod uint, encodingConfig EncodingConfig, appOpts servertypes.AppOptions,
	baseAppOptions ...func(*baseapp.BaseApp),
) *AIApp {
	legacyAmino := encodingConfig.Amino
	codec := encodingConfig.Codec
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(
		Name,
		logger,
		db,
		encodingConfig.TxConfig.TxDecoder(),
		baseAppOptions...,
	)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	keys := sdk.NewKVStoreKeys(
		authtypes.StoreKey,
		authzkeeper.StoreKey,
		banktypes.StoreKey,
		capabilitytypes.StoreKey,
		distrtypes.StoreKey,
		evidencetypes.StoreKey,
		feegrant.StoreKey,
		govtypes.StoreKey,
		paramstypes.StoreKey,
		slashingtypes.StoreKey,
		stakingtypes.StoreKey,
		upgradetypes.StoreKey,
		ibchost.StoreKey,
		ibctransfertypes.StoreKey,
	)
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)

	app := &AIApp{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		Codec:             codec,
		interfaceRegistry: interfaceRegistry,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	// **** create keepers ****

	// address of the gov module account
	authority := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	app.ParamsKeeper = initParamsKeeper(
		codec,
		legacyAmino,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)

	// set the BaseApp's parameter store
	bApp.SetParamStore(
		app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable()),
	)

	// add capability keeper and ScopeToModule for ibc module
	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		codec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)

	// grant capabilities for the ibc and ibc-transfer modules
	app.ScopedIBCKeeper = app.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	app.ScopedIBCTransferKeeper = app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)

	// add keepers
	app.AccountKeeper = authkeeper.NewAccountKeeper(
		codec,
		keys[authtypes.StoreKey],
		getSubspace(app, authtypes.ModuleName),
		authtypes.ProtoBaseAccount,
		maccPerms,
		AccountAddressPrefix,
	)
	app.AuthzKeeper = authzkeeper.NewKeeper(
		keys[authzkeeper.StoreKey],
		codec,
		app.BaseApp.MsgServiceRouter(),
		app.AccountKeeper,
	)
	app.BankKeeper = bankkeeper.NewBaseKeeper(
		codec,
		keys[banktypes.StoreKey],
		app.AccountKeeper,
		getSubspace(app, banktypes.ModuleName),
		getBlockedModuleAccountAddrs(app), // NOTE: fee collector & safety fund are excluded from blocked addresses
	)
	app.CrisisKeeper = crisiskeeper.NewKeeper(
		getSubspace(app, crisistypes.ModuleName),
		invCheckPeriod,
		app.BankKeeper,
		authtypes.FeeCollectorName,
	)
	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		codec,
		keys[feegrant.StoreKey],
		app.AccountKeeper,
	)
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		keys[upgradetypes.StoreKey],
		codec,
		homePath,
		app.BaseApp,
		authority,
	)

	// staking keeper and its dependencies
	// NOTE: the order here (e.g. evidence keeper depends on slashing keeper, so
	// must be defined after it)
	stakingKeeper := stakingkeeper.NewKeeper(
		codec,
		keys[stakingtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		getSubspace(app, stakingtypes.ModuleName),
	)
	app.DistrKeeper = distrkeeper.NewKeeper(
		codec,
		keys[distrtypes.StoreKey],
		getSubspace(app, distrtypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		&stakingKeeper,
		authtypes.FeeCollectorName,
	)
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		codec,
		keys[slashingtypes.StoreKey],
		&stakingKeeper,
		getSubspace(app, slashingtypes.ModuleName),
	)
	app.EvidenceKeeper = *evidencekeeper.NewKeeper(
		codec,
		keys[evidencetypes.StoreKey],
		&app.StakingKeeper,
		app.SlashingKeeper,
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain
	// these hooks.
	app.StakingKeeper = *stakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks()),
	)

	// create IBC Keeper
	app.IBCKeeper = ibckeeper.NewKeeper(
		codec,
		keys[ibchost.StoreKey],
		getSubspace(app, ibchost.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper, app.ScopedIBCKeeper,
	)
	app.IBCTransferKeeper = ibctransferkeeper.NewKeeper(
		codec,
		keys[ibctransfertypes.StoreKey],
		getSubspace(app, ibctransfertypes.ModuleName),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		app.ScopedIBCTransferKeeper,
	)

	// create static IBC router, add transfer route, then set and seal it
	app.IBCKeeper.SetRouter(initIBCRouter(app))

	govRouter := govv1beta1.NewRouter()
	govRouter.AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(distrtypes.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.DistrKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.UpgradeKeeper))
	govConfig := govtypes.DefaultConfig()

	openAIKey := cast.ToString(appOpts.Get(customgov.FlagOpenAIKey))
	aiRules := cast.ToString(appOpts.Get(customgov.FlagAIRules))
	customconfig := customgovtypes.Config{
		MaxMetadataLen: govConfig.MaxMetadataLen,
		OpenAIKey:      openAIKey,
		AIRules:        aiRules,
	}

	app.GovKeeper = customgovkeeper.NewKeeper(
		codec, keys[govtypes.StoreKey], getSubspace(app, govtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
		&stakingKeeper, govRouter, app.MsgServiceRouter(), customconfig,
	)

	// **** module options ****

	// NOTE: We may consider parsing `appOpts` inside module constructors. For
	// the moment we prefer to be more strict in what arguments the modules
	// expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later
	// modified must be passed by reference here.
	app.mm = module.NewManager(
		auth.NewAppModule(codec, app.AccountKeeper, nil),
		authzmodule.NewAppModule(codec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		bank.NewAppModule(codec, app.BankKeeper, app.AccountKeeper),
		capability.NewAppModule(codec, *app.CapabilityKeeper),
		crisis.NewAppModule(&app.CrisisKeeper, skipGenesisInvariants),
		distr.NewAppModule(codec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		feegrantmodule.NewAppModule(codec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		genutil.NewAppModule(app.AccountKeeper, app.StakingKeeper, app.BaseApp.DeliverTx, encodingConfig.TxConfig),
		customgov.NewAppModule(codec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		params.NewAppModule(app.ParamsKeeper),
		slashing.NewAppModule(codec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper),
		staking.NewAppModule(codec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		upgrade.NewAppModule(app.UpgradeKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		ibctransfer.NewAppModule(app.IBCTransferKeeper),
	)

	// During begin block, slashing happens after `distr.BeginBlocker` so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// `CanWithdrawInvariant` invariant.
	// NOTE: staking module is required if `HistoricalEntries` param > 0
	app.mm.SetOrderBeginBlockers(
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		ibchost.ModuleName,
		ibctransfertypes.ModuleName,
	)

	app.mm.SetOrderEndBlockers(
		crisistypes.ModuleName,
		govtypes.ModuleName,
		stakingtypes.ModuleName,
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		ibchost.ModuleName,
		ibctransfertypes.ModuleName,
	)

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: Capability module must occur first so that it can initialize any
	// capabilities so that other modules that want to create or claim
	// capabilities afterwards in `InitChain` can do so safely.
	app.mm.SetOrderInitGenesis(
		capabilitytypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		distrtypes.ModuleName,
		stakingtypes.ModuleName,
		slashingtypes.ModuleName,
		govtypes.ModuleName,
		crisistypes.ModuleName,
		genutiltypes.ModuleName,
		evidencetypes.ModuleName,
		authz.ModuleName,
		paramstypes.ModuleName,
		upgradetypes.ModuleName,
		feegrant.ModuleName,
		ibchost.ModuleName,
		ibctransfertypes.ModuleName,
	)

	app.mm.RegisterInvariants(&app.CrisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter(), encodingConfig.Amino)
	app.mm.RegisterServices(module.NewConfigurator(codec, app.MsgServiceRouter(), app.GRPCQueryRouter()))

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	anteHandler, err := NewAnteHandler(
		HandlerOptions{
			HandlerOptions: ante.HandlerOptions{
				AccountKeeper:   app.AccountKeeper,
				BankKeeper:      app.BankKeeper,
				FeegrantKeeper:  app.FeeGrantKeeper,
				SignModeHandler: encodingConfig.TxConfig.SignModeHandler(),
				SigGasConsumer:  ante.DefaultSigVerificationGasConsumer,
			},
			IBCKeeper: app.IBCKeeper,
		},
	)
	if err != nil {
		panic(err)
	}

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetAnteHandler(anteHandler)
	app.SetEndBlocker(app.EndBlocker)

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(err.Error())
		}
	}

	return app
}

//------------------------------------------------------------------------------
// Implement `aiapptypes.App` interface for AIApp
//
// `ExportAppStateAndValidators` is implemented in ./export.go so no need to
// reimplement here.
//------------------------------------------------------------------------------

// Name returns the app's name
func (app *AIApp) Name() string {
	return app.BaseApp.Name()
}

// LegacyAmino returns the app's legacy amino codec
func (app *AIApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// BeginBlocker application updates every begin block
func (app *AIApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker application updates on every end block
func (app *AIApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer application updates at chain initialization
func (app *AIApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := tmjson.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(err)
	}

	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())

	return app.mm.InitGenesis(ctx, app.Codec, genesisState)
}

// LoadHeight loads a particular height
func (app *AIApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs return the app's module account addresses
func (app *AIApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// NOTE: we don't actually use simulation manager anywhere in this project, so
// we simply return nil.
func (app *AIApp) SimulationManager() *module.SimulationManager {
	return nil
}

//------------------------------------------------------------------------------
// Implement `servertypes.Application` interface for `AIApp`
//
// `RegisterGRPCServer` is already implemented by `BaseApp` and inherited by AIApp
//------------------------------------------------------------------------------

// RegisterAPIRoutes registers all application module routes with the provided
// API server.
func (app *AIApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx

	// register new tx routes from grpc-gateway
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register new tendermint queries routes from grpc-gateway
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register grpc-gateway routes for all modules
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *AIApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(
		app.BaseApp.GRPCQueryRouter(),
		clientCtx,
		app.BaseApp.Simulate,
		app.interfaceRegistry,
	)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *AIApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

//------------------------------------------------------------------------------
// Helpers
//------------------------------------------------------------------------------

// returns a param subspace for a given module name
func getSubspace(app *AIApp, moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// getBlockedModuleAccountAddrs returns all the app's blocked module account
// addresses.
//
// Specifically, we allow the following module accounts to receive funds:
//
//   - `fee_collector` and `safety_fund`, so that protocol revenue can be sent
//     from outposts to the hub via IBC fungible token transfers
//
//   - `incentives`, so that the incentives module can draw funds from the
//     community pool in order to
//     create new incentives schedules upon successful governance proposals
//
// Further note on the 2nd point: the distrkeeper's `DistributeFromFeePool`
// function uses bankkeeper's `SendCoinsFromModuleToAccount` instead of
// `SendCoinsFromModuleToModule`. If it had used `FromModuleToModule`
// then we won't need to allow incentives module account to receive funds here.
//
// Forked from: https://github.com/cosmos/gaia/pull/1493
func getBlockedModuleAccountAddrs(app *AIApp) map[string]bool {
	modAccAddrs := app.ModuleAccountAddrs()

	delete(modAccAddrs, authtypes.NewModuleAddress(authtypes.FeeCollectorName).String())

	return modAccAddrs
}

// initializes params keeper and its subspaces
func initParamsKeeper(codec codec.BinaryCodec, legacyAmino *codec.LegacyAmino, key, tkey storetypes.StoreKey) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(codec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govv1.ParamKeyTable())
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(ibchost.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)

	return paramsKeeper
}

// initialzies IBC router
func initIBCRouter(app *AIApp) *ibcporttypes.Router {
	ibcRouter := ibcporttypes.NewRouter()

	ibcRouter.AddRoute(ibctransfertypes.ModuleName, ibctransfer.NewIBCModule(app.IBCTransferKeeper))

	return ibcRouter
}
