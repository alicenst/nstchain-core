package community_test

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"

	"github.com/nstchain/nstchain-core/x/community"
	"github.com/nstchain/nstchain-core/x/community/keeper"
	"github.com/nstchain/nstchain-core/x/community/testutil"
)

func TestABCIDisableInflation(t *testing.T) {
	testFunc := func(ctx sdk.Context, k keeper.Keeper) {
		community.BeginBlocker(ctx, k)
	}
	suite.Run(t, testutil.NewDisableInflationTestSuite(testFunc))
}
