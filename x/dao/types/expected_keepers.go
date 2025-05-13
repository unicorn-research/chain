package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// BankKeeper defines the expected interface for the Bank keeper
type BankKeeper interface {
	// Methods required from the bank keeper
	SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error
	SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error
	GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin
	GetAllBalances(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins
}

// AccountKeeper defines the expected interface for the Account keeper
type AccountKeeper interface {
	// Methods required from account keeper
	GetAccount(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI
	SetAccount(ctx sdk.Context, account sdk.AccountI)
	NewAccountWithAddress(ctx sdk.Context, addr sdk.AccAddress) sdk.AccountI
}

// StakingKeeper defines the expected interface for the Staking keeper (if needed)
type StakingKeeper interface {
	// Any staking methods required
}
