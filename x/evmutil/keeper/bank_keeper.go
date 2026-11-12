package keeper

import (
	"fmt"

	errorsmod "cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/nstchain/nstchain-core/x/evmutil/types"
)

const (
	// EvmDenom is the gas denom used by the evm
	EvmDenom = "anst"

	// CosmosDenom is the gas denom used by the ist app
	CosmosDenom = "unst"
)

// ConversionMultiplier is the conversion multiplier between anst and unst
var ConversionMultiplier = sdkmath.NewInt(1_000_000_000_000)

var _ evmtypes.BankKeeper = EvmBankKeeper{}

// EvmBankKeeper is a BankKeeper wrapper for the x/evm module to allow the use
// of the 18 decimal aist coin on the evm.
// x/evm consumes gas and send coins by minting and burning aist coins in its module
// account and then sending the funds to the target account.
// This keeper uses both the unst coin and a separate aist balance to manage the
// extra precision needed by the evm.
type EvmBankKeeper struct {
	aistKeeper Keeper
	bk         types.BankKeeper
	ak         types.AccountKeeper
}

func NewEvmBankKeeper(aistKeeper Keeper, bk types.BankKeeper, ak types.AccountKeeper) EvmBankKeeper {
	return EvmBankKeeper{
		aistKeeper: aistKeeper,
		bk:         bk,
		ak:         ak,
	}
}

// GetBalance returns the total **spendable** balance of aist for a given account by address.
func (k EvmBankKeeper) GetBalance(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	if denom != EvmDenom {
		panic(fmt.Errorf("only evm denom %s is supported by EvmBankKeeper", EvmDenom))
	}

	spendableCoins := k.bk.SpendableCoins(ctx, addr)
	unst := spendableCoins.AmountOf(CosmosDenom)
	anst := k.aistKeeper.GetBalance(ctx, addr)
	total := unst.Mul(ConversionMultiplier).Add(anst)
	return sdk.NewCoin(EvmDenom, total)
}

// SendCoins transfers aist coins from a AccAddress to an AccAddress.
func (k EvmBankKeeper) SendCoins(ctx sdk.Context, senderAddr sdk.AccAddress, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	// SendCoins method is not used by the evm module, but is required by the
	// evmtypes.BankKeeper interface. This must be updated if the evm module
	// is updated to use SendCoins.
	panic("not implemented")
}

// SendCoinsFromModuleToAccount transfers aist coins from a ModuleAccount to an AccAddress.
// It will panic if the module account does not exist. An error is returned if the recipient
// address is black-listed or if sending the tokens fails.
func (k EvmBankKeeper) SendCoinsFromModuleToAccount(ctx sdk.Context, senderModule string, recipientAddr sdk.AccAddress, amt sdk.Coins) error {
	unst, anst, err := SplitAistCoins(amt)
	if err != nil {
		return err
	}

	if unst.Amount.IsPositive() {
		if err := k.bk.SendCoinsFromModuleToAccount(ctx, senderModule, recipientAddr, sdk.NewCoins(unst)); err != nil {
			return err
		}
	}

	senderAddr := k.GetModuleAddress(senderModule)
	if err := k.ConvertOneunstToAistIfNeeded(ctx, senderAddr, anst); err != nil {
		return err
	}

	if err := k.aistKeeper.SendBalance(ctx, senderAddr, recipientAddr, anst); err != nil {
		return err
	}

	return k.ConvertAistTounst(ctx, recipientAddr)
}

// SendCoinsFromAccountToModule transfers aist coins from an AccAddress to a ModuleAccount.
// It will panic if the module account does not exist.
func (k EvmBankKeeper) SendCoinsFromAccountToModule(ctx sdk.Context, senderAddr sdk.AccAddress, recipientModule string, amt sdk.Coins) error {
	unst, aistNeeded, err := SplitAistCoins(amt)
	if err != nil {
		return err
	}

	if unst.IsPositive() {
		if err := k.bk.SendCoinsFromAccountToModule(ctx, senderAddr, recipientModule, sdk.NewCoins(unst)); err != nil {
			return err
		}
	}

	if err := k.ConvertOneunstToAistIfNeeded(ctx, senderAddr, aistNeeded); err != nil {
		return err
	}

	recipientAddr := k.GetModuleAddress(recipientModule)
	if err := k.aistKeeper.SendBalance(ctx, senderAddr, recipientAddr, aistNeeded); err != nil {
		return err
	}

	return k.ConvertAistTounst(ctx, recipientAddr)
}

// MintCoins mints aist coins by minting the equivalent unst coins and any remaining aist coins.
// It will panic if the module account does not exist or is unauthorized.
func (k EvmBankKeeper) MintCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	unst, anst, err := SplitAistCoins(amt)
	if err != nil {
		return err
	}

	if unst.IsPositive() {
		if err := k.bk.MintCoins(ctx, moduleName, sdk.NewCoins(unst)); err != nil {
			return err
		}
	}

	recipientAddr := k.GetModuleAddress(moduleName)
	if err := k.aistKeeper.AddBalance(ctx, recipientAddr, anst); err != nil {
		return err
	}

	return k.ConvertAistTounst(ctx, recipientAddr)
}

// BurnCoins burns aist coins by burning the equivalent unst coins and any remaining aist coins.
// It will panic if the module account does not exist or is unauthorized.
func (k EvmBankKeeper) BurnCoins(ctx sdk.Context, moduleName string, amt sdk.Coins) error {
	unst, anst, err := SplitAistCoins(amt)
	if err != nil {
		return err
	}

	if unst.IsPositive() {
		if err := k.bk.BurnCoins(ctx, moduleName, sdk.NewCoins(unst)); err != nil {
			return err
		}
	}

	moduleAddr := k.GetModuleAddress(moduleName)
	if err := k.ConvertOneunstToAistIfNeeded(ctx, moduleAddr, anst); err != nil {
		return err
	}

	return k.aistKeeper.RemoveBalance(ctx, moduleAddr, anst)
}

// IsSendEnabledCoins checks the coins provided and returns an ErrSendDisabled
// if any of the coins are not configured for sending. Returns nil if sending is
// enabled for all provided coins.
func (k EvmBankKeeper) IsSendEnabledCoins(ctx sdk.Context, coins ...sdk.Coin) error {
	// IsSendEnabledCoins method is not used by the evm module, but is required by the
	// evmtypes.BankKeeper interface. This must be updated if the evm module
	// is updated to use IsSendEnabledCoins.
	panic("not implemented")
}

// ConvertOneunstToAistIfNeeded converts 1 unst to aist for an address if
// its aist balance is smaller than the aistNeeded amount.
func (k EvmBankKeeper) ConvertOneunstToAistIfNeeded(ctx sdk.Context, addr sdk.AccAddress, aistNeeded sdkmath.Int) error {
	aistBal := k.aistKeeper.GetBalance(ctx, addr)
	if aistBal.GTE(aistNeeded) {
		return nil
	}

	unstToStore := sdk.NewCoins(sdk.NewCoin(CosmosDenom, sdk.OneInt()))
	if err := k.bk.SendCoinsFromAccountToModule(ctx, addr, types.ModuleName, unstToStore); err != nil {
		return err
	}

	// add 1unst equivalent of aist to addr
	aistToReceive := ConversionMultiplier
	if err := k.aistKeeper.AddBalance(ctx, addr, aistToReceive); err != nil {
		return err
	}

	return nil
}

// ConvertAistTounst converts all available aist to unst for a given AccAddress.
func (k EvmBankKeeper) ConvertAistTounst(ctx sdk.Context, addr sdk.AccAddress) error {
	totalAist := k.aistKeeper.GetBalance(ctx, addr)
	unst, _, err := SplitAistCoins(sdk.NewCoins(sdk.NewCoin(EvmDenom, totalAist)))
	if err != nil {
		return err
	}

	// do nothing if account does not have enough aist for a single unst
	unstToReceive := unst.Amount
	if !unstToReceive.IsPositive() {
		return nil
	}

	// remove aist used for converting to unst
	aistToBurn := unstToReceive.Mul(ConversionMultiplier)
	finalBal := totalAist.Sub(aistToBurn)
	if err := k.aistKeeper.SetBalance(ctx, addr, finalBal); err != nil {
		return err
	}

	fromAddr := k.GetModuleAddress(types.ModuleName)
	if err := k.bk.SendCoins(ctx, fromAddr, addr, sdk.NewCoins(unst)); err != nil {
		return err
	}

	return nil
}

func (k EvmBankKeeper) GetModuleAddress(moduleName string) sdk.AccAddress {
	addr := k.ak.GetModuleAddress(moduleName)
	if addr == nil {
		panic(errorsmod.Wrapf(sdkerrors.ErrUnknownAddress, "module account %s does not exist", moduleName))
	}
	return addr
}

// SplitAistCoins splits aist coins to the equivalent unst coins and any remaining aist balance.
// An error will be returned if the coins are not valid or if the coins are not the aist denom.
func SplitAistCoins(coins sdk.Coins) (sdk.Coin, sdkmath.Int, error) {
	anst := sdk.ZeroInt()
	unst := sdk.NewCoin(CosmosDenom, sdk.ZeroInt())

	if len(coins) == 0 {
		return unst, anst, nil
	}

	if err := ValidateEvmCoins(coins); err != nil {
		return unst, anst, err
	}

	// note: we should always have len(coins) == 1 here since coins cannot have dup denoms after we validate.
	coin := coins[0]
	remainingBalance := coin.Amount.Mod(ConversionMultiplier)
	if remainingBalance.IsPositive() {
		anst = remainingBalance
	}
	unstAmount := coin.Amount.Quo(ConversionMultiplier)
	if unstAmount.IsPositive() {
		unst = sdk.NewCoin(CosmosDenom, unstAmount)
	}

	return unst, anst, nil
}

// ValidateEvmCoins validates the coins from evm is valid and is the EvmDenom (aist).
func ValidateEvmCoins(coins sdk.Coins) error {
	if len(coins) == 0 {
		return nil
	}

	// validate that coins are non-negative, sorted, and no dup denoms
	if err := coins.Validate(); err != nil {
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, coins.String())
	}

	// validate that coin denom is aist
	if len(coins) != 1 || coins[0].Denom != EvmDenom {
		errMsg := fmt.Sprintf("invalid evm coin denom, only %s is supported", EvmDenom)
		return errorsmod.Wrap(sdkerrors.ErrInvalidCoins, errMsg)
	}

	return nil
}
