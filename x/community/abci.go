package community

import (
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/nstchain/nstchain-core/x/community/keeper"
	"github.com/nstchain/nstchain-core/x/community/types"
)

// BeginBlocker runs the community module begin blocker logic.
func BeginBlocker(ctx sdk.Context, k keeper.Keeper) {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyBeginBlocker)

	// This exact call order is required to allow payout on the upgrade block
	k.CheckAndDisableMintAndKavaDistInflation(ctx)
	k.PayoutAccumulatedStakingRewards(ctx)
}
