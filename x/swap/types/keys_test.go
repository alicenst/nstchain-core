package types_test

import (
	"testing"

	"github.com/nstchain/nstchain-core/x/swap/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/assert"
)

func TestKeys(t *testing.T) {
	key := types.PoolKey(types.PoolID("ukava", "usdx"))
	assert.Equal(t, types.PoolID("ukava", "usdx"), string(key))

	key = types.DepositorPoolSharesKey(sdk.AccAddress("testaddress1"), types.PoolID("ukava", "usdx"))
	assert.Equal(t, string(sdk.AccAddress("testaddress1"))+"|"+types.PoolID("ukava", "usdx"), string(key))
}
