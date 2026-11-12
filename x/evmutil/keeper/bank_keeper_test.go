package keeper_test

import (
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	tmtime "github.com/cometbft/cometbft/types/time"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	vesting "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	evmtypes "github.com/evmos/ethermint/x/evm/types"

	"github.com/kava-labs/kava/x/evmutil/keeper"
	"github.com/kava-labs/kava/x/evmutil/testutil"
	"github.com/kava-labs/kava/x/evmutil/types"
)

type evmBankKeeperTestSuite struct {
	testutil.Suite
}

func (suite *evmBankKeeperTestSuite) SetupTest() {
	suite.Suite.SetupTest()
}

func (suite *evmBankKeeperTestSuite) TestGetBalance_ReturnsSpendable() {
	startingCoins := sdk.NewCoins(sdk.NewInt64Coin("unst", 10))
	startingAist := sdkmath.NewInt(100)

	now := tmtime.Now()
	endTime := now.Add(24 * time.Hour)
	bacc := authtypes.NewBaseAccountWithAddress(suite.Addrs[0])
	vacc := vesting.NewContinuousVestingAccount(bacc, startingCoins, now.Unix(), endTime.Unix())
	suite.AccountKeeper.SetAccount(suite.Ctx, vacc)

	err := suite.App.FundAccount(suite.Ctx, suite.Addrs[0], startingCoins)
	suite.Require().NoError(err)
	err = suite.Keeper.SetBalance(suite.Ctx, suite.Addrs[0], startingAist)
	suite.Require().NoError(err)

	coin := suite.EvmBankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "anst")
	suite.Require().Equal(startingAist, coin.Amount)

	ctx := suite.Ctx.WithBlockTime(now.Add(12 * time.Hour))
	coin = suite.EvmBankKeeper.GetBalance(ctx, suite.Addrs[0], "anst")
	suite.Require().Equal(sdkmath.NewIntFromUint64(5_000_000_000_100), coin.Amount)
}

func (suite *evmBankKeeperTestSuite) TestGetBalance_NotEvmDenom() {
	suite.Require().Panics(func() {
		suite.EvmBankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "unst")
	})
	suite.Require().Panics(func() {
		suite.EvmBankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "busd")
	})
}

func (suite *evmBankKeeperTestSuite) TestGetBalance() {
	tests := []struct {
		name           string
		startingAmount sdk.Coins
		expAmount      sdkmath.Int
	}{
		{
			"unst with anst",
			sdk.NewCoins(
				sdk.NewInt64Coin("anst", 100),
				sdk.NewInt64Coin("unst", 10),
			),
			sdkmath.NewInt(10_000_000_000_100),
		},
		{
			"just anst",
			sdk.NewCoins(
				sdk.NewInt64Coin("anst", 100),
				sdk.NewInt64Coin("busd", 100),
			),
			sdkmath.NewInt(100),
		},
		{
			"just unst",
			sdk.NewCoins(
				sdk.NewInt64Coin("unst", 10),
				sdk.NewInt64Coin("busd", 100),
			),
			sdkmath.NewInt(10_000_000_000_000),
		},
		{
			"no unst or anst",
			sdk.NewCoins(),
			sdk.ZeroInt(),
		},
		{
			"with anst that is more than 1 unst",
			sdk.NewCoins(
				sdk.NewInt64Coin("anst", 20_000_000_000_220),
				sdk.NewInt64Coin("unst", 11),
			),
			sdkmath.NewInt(31_000_000_000_220),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			suite.FundAccountWithKava(suite.Addrs[0], tt.startingAmount)
			coin := suite.EvmBankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "aist")
			suite.Require().Equal(tt.expAmount, coin.Amount)
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestSendCoinsFromModuleToAccount() {
	startingModuleCoins := sdk.NewCoins(
		sdk.NewInt64Coin("aist", 200),
		sdk.NewInt64Coin("unst", 100),
	)
	tests := []struct {
		name           string
		sendCoins      sdk.Coins
		startingAccBal sdk.Coins
		expAccBal      sdk.Coins
		hasErr         bool
	}{
		{
			"send more than 1 unst",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 12_000_000_000_010)),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin("aist", 10),
				sdk.NewInt64Coin("unst", 12),
			),
			false,
		},
		{
			"send less than 1 unst",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 122)),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin("aist", 122),
				sdk.NewInt64Coin("unst", 0),
			),
			false,
		},
		{
			"send an exact amount of unst",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 98_000_000_000_000)),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin("aist", 0o0),
				sdk.NewInt64Coin("unst", 98),
			),
			false,
		},
		{
			"send no aist",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 0)),
			sdk.Coins{},
			sdk.NewCoins(
				sdk.NewInt64Coin("aist", 0),
				sdk.NewInt64Coin("unst", 0),
			),
			false,
		},
		{
			"errors if sending other coins",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 500), sdk.NewInt64Coin("busd", 1000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if not enough total aist to cover",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 100_000_000_001_000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if not enough unst to cover",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 200_000_000_000_000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"converts receiver's aist to unst if there's enough aist after the transfer",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 99_000_000_000_200)),
			sdk.NewCoins(
				sdk.NewInt64Coin("aist", 999_999_999_900),
				sdk.NewInt64Coin("unst", 1),
			),
			sdk.NewCoins(
				sdk.NewInt64Coin("aist", 100),
				sdk.NewInt64Coin("unst", 101),
			),
			false,
		},
		{
			"converts all of receiver's aist to unst even if somehow receiver has more than 1unst of aist",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 12_000_000_000_100)),
			sdk.NewCoins(
				sdk.NewInt64Coin("aist", 5_999_999_999_990),
				sdk.NewInt64Coin("unst", 1),
			),
			sdk.NewCoins(
				sdk.NewInt64Coin("aist", 90),
				sdk.NewInt64Coin("unst", 19),
			),
			false,
		},
		{
			"swap 1 unst for aist if module account doesn't have enough aist",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 99_000_000_001_000)),
			sdk.NewCoins(
				sdk.NewInt64Coin("aist", 200),
				sdk.NewInt64Coin("unst", 1),
			),
			sdk.NewCoins(
				sdk.NewInt64Coin("aist", 1200),
				sdk.NewInt64Coin("unst", 100),
			),
			false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			suite.FundAccountWithKava(suite.Addrs[0], tt.startingAccBal)
			suite.FundModuleAccountWithKava(evmtypes.ModuleName, startingModuleCoins)

			// fund our module with some unst to account for converting extra aist back to unst
			suite.FundModuleAccountWithKava(types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("unst", 10)))

			err := suite.EvmBankKeeper.SendCoinsFromModuleToAccount(suite.Ctx, evmtypes.ModuleName, suite.Addrs[0], tt.sendCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check unst
			unstSender := suite.BankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "unst")
			suite.Require().Equal(tt.expAccBal.AmountOf("unst").Int64(), unstSender.Amount.Int64())

			// check aist
			actualAkava := suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[0])
			suite.Require().Equal(tt.expAccBal.AmountOf("aist").Int64(), actualAkava.Int64())
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestSendCoinsFromAccountToModule() {
	startingAccCoins := sdk.NewCoins(
		sdk.NewInt64Coin("aist", 200),
		sdk.NewInt64Coin("unst", 100),
	)
	startingModuleCoins := sdk.NewCoins(
		sdk.NewInt64Coin("aist", 100_000_000_000),
	)
	tests := []struct {
		name           string
		sendCoins      sdk.Coins
		expSenderCoins sdk.Coins
		expModuleCoins sdk.Coins
		hasErr         bool
	}{
		{
			"send more than 1 unst",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 12_000_000_000_010)),
			sdk.NewCoins(sdk.NewInt64Coin("aist", 190), sdk.NewInt64Coin("unst", 88)),
			sdk.NewCoins(sdk.NewInt64Coin("aist", 100_000_000_010), sdk.NewInt64Coin("unst", 12)),
			false,
		},
		{
			"send less than 1 unst",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 122)),
			sdk.NewCoins(sdk.NewInt64Coin("aist", 78), sdk.NewInt64Coin("unst", 100)),
			sdk.NewCoins(sdk.NewInt64Coin("aist", 100_000_000_122), sdk.NewInt64Coin("unst", 0)),
			false,
		},
		{
			"send an exact amount of unst",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 98_000_000_000_000)),
			sdk.NewCoins(sdk.NewInt64Coin("aist", 200), sdk.NewInt64Coin("unst", 2)),
			sdk.NewCoins(sdk.NewInt64Coin("aist", 100_000_000_000), sdk.NewInt64Coin("unst", 98)),
			false,
		},
		{
			"send no aist",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 0)),
			sdk.NewCoins(sdk.NewInt64Coin("aist", 200), sdk.NewInt64Coin("unst", 100)),
			sdk.NewCoins(sdk.NewInt64Coin("aist", 100_000_000_000), sdk.NewInt64Coin("unst", 0)),
			false,
		},
		{
			"errors if sending other coins",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 500), sdk.NewInt64Coin("busd", 1000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if have dup coins",
			sdk.Coins{
				sdk.NewInt64Coin("aist", 12_000_000_000_000),
				sdk.NewInt64Coin("aist", 2_000_000_000_000),
			},
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if not enough total aist to cover",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 100_000_000_001_000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"errors if not enough unst to cover",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 200_000_000_000_000)),
			sdk.Coins{},
			sdk.Coins{},
			true,
		},
		{
			"converts 1 unst to aist if not enough aist to cover",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 99_001_000_000_000)),
			sdk.NewCoins(sdk.NewInt64Coin("aist", 999_000_000_200), sdk.NewInt64Coin("unst", 0)),
			sdk.NewCoins(sdk.NewInt64Coin("aist", 101_000_000_000), sdk.NewInt64Coin("unst", 99)),
			false,
		},
		{
			"converts receiver's aist to unst if there's enough aist after the transfer",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 5_900_000_000_200)),
			sdk.NewCoins(sdk.NewInt64Coin("aist", 100_000_000_000), sdk.NewInt64Coin("unst", 94)),
			sdk.NewCoins(sdk.NewInt64Coin("aist", 200), sdk.NewInt64Coin("unst", 6)),
			false,
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()
			suite.FundAccountWithKava(suite.Addrs[0], startingAccCoins)
			suite.FundModuleAccountWithKava(evmtypes.ModuleName, startingModuleCoins)

			err := suite.EvmBankKeeper.SendCoinsFromAccountToModule(suite.Ctx, suite.Addrs[0], evmtypes.ModuleName, tt.sendCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check sender balance
			unstSender := suite.BankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "unst")
			suite.Require().Equal(tt.expSenderCoins.AmountOf("unst").Int64(), unstSender.Amount.Int64())
			actualAkava := suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[0])
			suite.Require().Equal(tt.expSenderCoins.AmountOf("aist").Int64(), actualAkava.Int64())

			// check module balance
			moduleAddr := suite.AccountKeeper.GetModuleAddress(evmtypes.ModuleName)
			unstSender = suite.BankKeeper.GetBalance(suite.Ctx, moduleAddr, "unst")
			suite.Require().Equal(tt.expModuleCoins.AmountOf("unst").Int64(), unstSender.Amount.Int64())
			actualAkava = suite.Keeper.GetBalance(suite.Ctx, moduleAddr)
			suite.Require().Equal(tt.expModuleCoins.AmountOf("aist").Int64(), actualAkava.Int64())
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestBurnCoins() {
	startingUkava := sdkmath.NewInt(100)
	tests := []struct {
		name      string
		burnCoins sdk.Coins
		expUkava  sdkmath.Int
		expAkava  sdkmath.Int
		hasErr    bool
		aistStart sdkmath.Int
	}{
		{
			"burn more than 1 unst",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 12_021_000_000_002)),
			sdkmath.NewInt(88),
			sdkmath.NewInt(100_000_000_000),
			false,
			sdkmath.NewInt(121_000_000_002),
		},
		{
			"burn less than 1 unst",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 122)),
			sdkmath.NewInt(100),
			sdkmath.NewInt(878),
			false,
			sdkmath.NewInt(1000),
		},
		{
			"burn an exact amount of unst",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 98_000_000_000_000)),
			sdkmath.NewInt(2),
			sdkmath.NewInt(10),
			false,
			sdkmath.NewInt(10),
		},
		{
			"burn no aist",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 0)),
			startingUkava,
			sdk.ZeroInt(),
			false,
			sdk.ZeroInt(),
		},
		{
			"errors if burning other coins",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 500), sdk.NewInt64Coin("busd", 1000)),
			startingUkava,
			sdkmath.NewInt(100),
			true,
			sdkmath.NewInt(100),
		},
		{
			"errors if have dup coins",
			sdk.Coins{
				sdk.NewInt64Coin("aist", 12_000_000_000_000),
				sdk.NewInt64Coin("aist", 2_000_000_000_000),
			},
			startingUkava,
			sdk.ZeroInt(),
			true,
			sdk.ZeroInt(),
		},
		{
			"errors if burn amount is negative",
			sdk.Coins{sdk.Coin{Denom: "aist", Amount: sdkmath.NewInt(-100)}},
			startingUkava,
			sdkmath.NewInt(50),
			true,
			sdkmath.NewInt(50),
		},
		{
			"errors if not enough aist to cover burn",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 100_999_000_000_000)),
			sdkmath.NewInt(0),
			sdkmath.NewInt(99_000_000_000),
			true,
			sdkmath.NewInt(99_000_000_000),
		},
		{
			"errors if not enough unst to cover burn",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 200_000_000_000_000)),
			sdkmath.NewInt(100),
			sdk.ZeroInt(),
			true,
			sdk.ZeroInt(),
		},
		{
			"converts 1 unst to aist if not enough aist to cover",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 12_021_000_000_002)),
			sdkmath.NewInt(87),
			sdkmath.NewInt(980_000_000_000),
			false,
			sdkmath.NewInt(1_000_000_002),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()
			startingCoins := sdk.NewCoins(
				sdk.NewCoin("unst", startingUkava),
				sdk.NewCoin("aist", tt.aistStart),
			)
			suite.FundModuleAccountWithKava(evmtypes.ModuleName, startingCoins)

			err := suite.EvmBankKeeper.BurnCoins(suite.Ctx, evmtypes.ModuleName, tt.burnCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check unst
			unstActual := suite.BankKeeper.GetBalance(suite.Ctx, suite.EvmModuleAddr, "unst")
			suite.Require().Equal(tt.expUkava, unstActual.Amount)

			// check aist
			aistActual := suite.Keeper.GetBalance(suite.Ctx, suite.EvmModuleAddr)
			suite.Require().Equal(tt.expAkava, aistActual)
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestMintCoins() {
	tests := []struct {
		name      string
		mintCoins sdk.Coins
		unst      sdkmath.Int
		aist      sdkmath.Int
		hasErr    bool
		aistStart sdkmath.Int
	}{
		{
			"mint more than 1 unst",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 12_021_000_000_002)),
			sdkmath.NewInt(12),
			sdkmath.NewInt(21_000_000_002),
			false,
			sdk.ZeroInt(),
		},
		{
			"mint less than 1 unst",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 901_000_000_001)),
			sdk.ZeroInt(),
			sdkmath.NewInt(901_000_000_001),
			false,
			sdk.ZeroInt(),
		},
		{
			"mint an exact amount of unst",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 123_000_000_000_000_000)),
			sdkmath.NewInt(123_000),
			sdk.ZeroInt(),
			false,
			sdk.ZeroInt(),
		},
		{
			"mint no aist",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 0)),
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			false,
			sdk.ZeroInt(),
		},
		{
			"errors if minting other coins",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 500), sdk.NewInt64Coin("busd", 1000)),
			sdk.ZeroInt(),
			sdkmath.NewInt(100),
			true,
			sdkmath.NewInt(100),
		},
		{
			"errors if have dup coins",
			sdk.Coins{
				sdk.NewInt64Coin("aist", 12_000_000_000_000),
				sdk.NewInt64Coin("aist", 2_000_000_000_000),
			},
			sdk.ZeroInt(),
			sdk.ZeroInt(),
			true,
			sdk.ZeroInt(),
		},
		{
			"errors if mint amount is negative",
			sdk.Coins{sdk.Coin{Denom: "aist", Amount: sdkmath.NewInt(-100)}},
			sdk.ZeroInt(),
			sdkmath.NewInt(50),
			true,
			sdkmath.NewInt(50),
		},
		{
			"adds to existing aist balance",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 12_021_000_000_002)),
			sdkmath.NewInt(12),
			sdkmath.NewInt(21_000_000_102),
			false,
			sdkmath.NewInt(100),
		},
		{
			"convert aist balance to unst if it exceeds 1 unst",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 10_999_000_000_000)),
			sdkmath.NewInt(12),
			sdkmath.NewInt(1_200_000_001),
			false,
			sdkmath.NewInt(1_002_200_000_001),
		},
	}

	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()
			suite.FundModuleAccountWithKava(types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("unst", 10)))
			suite.FundModuleAccountWithKava(evmtypes.ModuleName, sdk.NewCoins(sdk.NewCoin("aist", tt.aistStart)))

			err := suite.EvmBankKeeper.MintCoins(suite.Ctx, evmtypes.ModuleName, tt.mintCoins)
			if tt.hasErr {
				suite.Require().Error(err)
				return
			} else {
				suite.Require().NoError(err)
			}

			// check unst
			unstActual := suite.BankKeeper.GetBalance(suite.Ctx, suite.EvmModuleAddr, "unst")
			suite.Require().Equal(tt.unst, unstActual.Amount)

			// check aist
			aistActual := suite.Keeper.GetBalance(suite.Ctx, suite.EvmModuleAddr)
			suite.Require().Equal(tt.aist, aistActual)
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestValidateEvmCoins() {
	tests := []struct {
		name      string
		coins     sdk.Coins
		shouldErr bool
	}{
		{
			"valid coins",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 500)),
			false,
		},
		{
			"dup coins",
			sdk.Coins{sdk.NewInt64Coin("aist", 500), sdk.NewInt64Coin("aist", 500)},
			true,
		},
		{
			"not evm coins",
			sdk.NewCoins(sdk.NewInt64Coin("unst", 500)),
			true,
		},
		{
			"negative coins",
			sdk.Coins{sdk.Coin{Denom: "aist", Amount: sdkmath.NewInt(-500)}},
			true,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			err := keeper.ValidateEvmCoins(tt.coins)
			if tt.shouldErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
			}
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestConvertOneUkavaToAkavaIfNeeded() {
	aistNeeded := sdkmath.NewInt(200)
	tests := []struct {
		name          string
		startingCoins sdk.Coins
		expectedCoins sdk.Coins
		success       bool
	}{
		{
			"not enough unst for conversion",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 100)),
			sdk.NewCoins(sdk.NewInt64Coin("aist", 100)),
			false,
		},
		{
			"converts 1 unst to aist",
			sdk.NewCoins(sdk.NewInt64Coin("unst", 10), sdk.NewInt64Coin("aist", 100)),
			sdk.NewCoins(sdk.NewInt64Coin("unst", 9), sdk.NewInt64Coin("aist", 1_000_000_000_100)),
			true,
		},
		{
			"conversion not needed",
			sdk.NewCoins(sdk.NewInt64Coin("unst", 10), sdk.NewInt64Coin("aist", 200)),
			sdk.NewCoins(sdk.NewInt64Coin("unst", 10), sdk.NewInt64Coin("aist", 200)),
			true,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			suite.FundAccountWithKava(suite.Addrs[0], tt.startingCoins)
			err := suite.EvmBankKeeper.ConvertOneunstToAistIfNeeded(suite.Ctx, suite.Addrs[0], aistNeeded)
			moduleKava := suite.BankKeeper.GetBalance(suite.Ctx, suite.AccountKeeper.GetModuleAddress(types.ModuleName), "unst")
			if tt.success {
				suite.Require().NoError(err)
				if tt.startingCoins.AmountOf("aist").LT(aistNeeded) {
					suite.Require().Equal(sdk.OneInt(), moduleKava.Amount)
				}
			} else {
				suite.Require().Error(err)
				suite.Require().Equal(sdk.ZeroInt(), moduleKava.Amount)
			}

			aist := suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[0])
			suite.Require().Equal(tt.expectedCoins.AmountOf("aist"), aist)
			unst := suite.BankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "unst")
			suite.Require().Equal(tt.expectedCoins.AmountOf("unst"), unst.Amount)
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestConvertAkavaToUkava() {
	tests := []struct {
		name          string
		startingCoins sdk.Coins
		expectedCoins sdk.Coins
	}{
		{
			"not enough unst",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 100)),
			sdk.NewCoins(sdk.NewInt64Coin("aist", 100), sdk.NewInt64Coin("unst", 0)),
		},
		{
			"converts aist for 1 unst",
			sdk.NewCoins(sdk.NewInt64Coin("unst", 10), sdk.NewInt64Coin("aist", 1_000_000_000_003)),
			sdk.NewCoins(sdk.NewInt64Coin("unst", 11), sdk.NewInt64Coin("aist", 3)),
		},
		{
			"converts more than 1 unst of aist",
			sdk.NewCoins(sdk.NewInt64Coin("unst", 10), sdk.NewInt64Coin("aist", 8_000_000_000_123)),
			sdk.NewCoins(sdk.NewInt64Coin("unst", 18), sdk.NewInt64Coin("aist", 123)),
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			suite.SetupTest()

			err := suite.App.FundModuleAccount(suite.Ctx, types.ModuleName, sdk.NewCoins(sdk.NewInt64Coin("unst", 10)))
			suite.Require().NoError(err)
			suite.FundAccountWithKava(suite.Addrs[0], tt.startingCoins)
			err = suite.EvmBankKeeper.ConvertAistTounst(suite.Ctx, suite.Addrs[0])
			suite.Require().NoError(err)
			aist := suite.Keeper.GetBalance(suite.Ctx, suite.Addrs[0])
			suite.Require().Equal(tt.expectedCoins.AmountOf("aist"), aist)
			unst := suite.BankKeeper.GetBalance(suite.Ctx, suite.Addrs[0], "unst")
			suite.Require().Equal(tt.expectedCoins.AmountOf("unst"), unst.Amount)
		})
	}
}

func (suite *evmBankKeeperTestSuite) TestSplitAkavaCoins() {
	tests := []struct {
		name          string
		coins         sdk.Coins
		expectedCoins sdk.Coins
		shouldErr     bool
	}{
		{
			"invalid coins",
			sdk.NewCoins(sdk.NewInt64Coin("unst", 500)),
			nil,
			true,
		},
		{
			"empty coins",
			sdk.NewCoins(),
			sdk.NewCoins(),
			false,
		},
		{
			"unst & aist coins",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 8_000_000_000_123)),
			sdk.NewCoins(sdk.NewInt64Coin("unst", 8), sdk.NewInt64Coin("aist", 123)),
			false,
		},
		{
			"only aist",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 10_123)),
			sdk.NewCoins(sdk.NewInt64Coin("aist", 10_123)),
			false,
		},
		{
			"only unst",
			sdk.NewCoins(sdk.NewInt64Coin("aist", 5_000_000_000_000)),
			sdk.NewCoins(sdk.NewInt64Coin("unst", 5)),
			false,
		},
	}
	for _, tt := range tests {
		suite.Run(tt.name, func() {
			unst, aist, err := keeper.SplitAistCoins(tt.coins)
			if tt.shouldErr {
				suite.Require().Error(err)
			} else {
				suite.Require().NoError(err)
				suite.Require().Equal(tt.expectedCoins.AmountOf("unst"), unst.Amount)
				suite.Require().Equal(tt.expectedCoins.AmountOf("aist"), aist)
			}
		})
	}
}

func TestEvmBankKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(evmBankKeeperTestSuite))
}
