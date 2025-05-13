package dao

import (
	"github.com/cosmos/cosmos-sdk/codec"

	"github.com/unicorn-reseaerch/chain/x/dao/keeper"
	"github.com/unicorn-reseaerch/chain/x/dao/module"
)

// AppModule implements an application module for the DAO module.
type AppModule = module.AppModule

// NewAppModule creates and returns a new DAO AppModule.
func NewAppModule(cdc codec.Codec, k keeper.Keeper) AppModule {
	return module.NewAppModule(cdc, k)
}
