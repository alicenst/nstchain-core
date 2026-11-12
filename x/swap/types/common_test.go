package types_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/nstchain/nstchain-core/app"
)

func init() {
	kavaConfig := sdk.GetConfig()
	app.SetBech32AddressPrefixes(kavaConfig)
	app.SetBip44CoinType(kavaConfig)
	kavaConfig.Seal()
}
