package client

import (
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"

	"github.com/nstchain/nstchain-core/x/earn/client/cli"
)

// community-pool deposit/withdraw proposal handlers
var (
	DepositProposalHandler  = govclient.NewProposalHandler(cli.GetCmdSubmitCommunityPoolDepositProposal)
	WithdrawProposalHandler = govclient.NewProposalHandler(cli.GetCmdSubmitCommunityPoolWithdrawProposal)
)
