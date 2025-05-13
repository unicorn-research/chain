package module

import (
	"encoding/json"
	"fmt"

	"github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/spf13/cobra"

	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"

	"github.com/unicorn-reseaerch/chain/x/dao/keeper"
	"github.com/unicorn-reseaerch/chain/x/dao/types"
)

var (
	_ module.AppModule      = AppModule{}
	_ module.AppModuleBasic = AppModuleBasic{}
)

// AppModuleBasic defines the basic application module for the DAO module.
type AppModuleBasic struct {
	cdc codec.Codec
}

// Name returns the DAO module's name.
func (AppModuleBasic) Name() string {
	return types.ModuleName
}

// RegisterLegacyAminoCodec registers the DAO module's types for the given codec.
func (AppModuleBasic) RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	// types.RegisterLegacyAminoCodec(cdc)
}

// RegisterInterfaces registers interfaces and implementations of the DAO module.
func (AppModuleBasic) RegisterInterfaces(registry codectypes.InterfaceRegistry) {
	// types.RegisterInterfaces(registry)
}

// DefaultGenesis returns default genesis state as raw bytes for the DAO module.
func (AppModuleBasic) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(types.DefaultGenesis())
}

// ValidateGenesis performs genesis state validation for the DAO module.
func (AppModuleBasic) ValidateGenesis(cdc codec.JSONCodec, _ client.TxEncodingConfig, bz json.RawMessage) error {
	var data types.GenesisState
	if err := cdc.UnmarshalJSON(bz, &data); err != nil {
		return fmt.Errorf("failed to unmarshal %s genesis state: %w", types.ModuleName, err)
	}
	return data.Validate()
}

// RegisterGRPCGatewayRoutes registers the gRPC Gateway routes for the DAO module.
func (AppModuleBasic) RegisterGRPCGatewayRoutes(clientCtx client.Context, mux *runtime.ServeMux) {
	// types.RegisterQueryHandlerClient(context.Background(), mux, types.NewQueryClient(clientCtx))
}

// GetTxCmd returns the root tx command for the DAO module.
func (AppModuleBasic) GetTxCmd() *cobra.Command {
	return nil // To be implemented
}

// GetQueryCmd returns the root query command for the DAO module.
func (AppModuleBasic) GetQueryCmd() *cobra.Command {
	return nil // To be implemented
}

// AppModule implements an application module for the DAO module.
type AppModule struct {
	AppModuleBasic
	keeper keeper.Keeper
}

// NewAppModule creates a new AppModule object
func NewAppModule(cdc codec.Codec, k keeper.Keeper) AppModule {
	return AppModule{
		AppModuleBasic: AppModuleBasic{cdc: cdc},
		keeper:         k,
	}
}

// IsAppModule implements the appmodule.AppModule interface.
func (AppModule) IsAppModule() {}

// Name returns the DAO module's name.
func (AppModule) Name() string {
	return types.ModuleName
}

// RegisterInvariants registers the DAO module invariants.
func (am AppModule) RegisterInvariants(_ sdk.InvariantRegistry) {}

// Route returns the message routing key for the DAO module.
func (am AppModule) Route() sdk.Route {
	return sdk.Route{}
}

// QuerierRoute returns the DAO module's querier route name.
func (AppModule) QuerierRoute() string {
	return types.ModuleName
}

// RegisterServices registers module services.
func (am AppModule) RegisterServices(cfg module.Configurator) {
	// To be implemented: register module services
}

// InitGenesis performs genesis initialization for the DAO module.
func (am AppModule) InitGenesis(ctx sdk.Context, cdc codec.JSONCodec, data json.RawMessage) {
	var genesisState types.GenesisState
	cdc.MustUnmarshalJSON(data, &genesisState)
	am.keeper.InitGenesis(ctx, genesisState)
}

// ExportGenesis returns the exported genesis state as raw bytes for the DAO module.
func (am AppModule) ExportGenesis(ctx sdk.Context, cdc codec.JSONCodec) json.RawMessage {
	gs := am.keeper.ExportGenesis(ctx)
	return cdc.MustMarshalJSON(gs)
}

// ConsensusVersion implements AppModule/ConsensusVersion.
func (AppModule) ConsensusVersion() uint64 { return 1 }

// BeginBlock returns the begin blocker for the DAO module.
func (am AppModule) BeginBlock(ctx sdk.Context) {}

// EndBlock returns the end blocker for the DAO module.
func (am AppModule) EndBlock(ctx sdk.Context) []abci.ValidatorUpdate {
	// Check for DAOs that haven't had any actions in a year and mark them for dissolution
	am.keeper.CheckForInactiveDaos(ctx)
	return []abci.ValidatorUpdate{}
}
