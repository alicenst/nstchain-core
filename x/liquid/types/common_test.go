package types_test

import (
	"os"
	"testing"

	"github.com/nstchain/nstchain-core/app"
)

func TestMain(m *testing.M) {
	app.SetSDKConfig()
	os.Exit(m.Run())
}
