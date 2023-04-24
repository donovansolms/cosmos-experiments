package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func (keeper Keeper) ActivateVotingPeriod(ctx sdk.Context, proposal v1.Proposal) {

	fmt.Println("\n\nTHIS PROPOSAL IS NOW ACTIVE, LETS VOTE\n\n")
	fmt.Println(keeper.config.OpenAIKey)
	fmt.Println(keeper.config.AIRules)

	startTime := ctx.BlockHeader().Time
	proposal.VotingStartTime = &startTime
	votingPeriod := keeper.GetVotingParams(ctx).VotingPeriod
	endTime := proposal.VotingStartTime.Add(*votingPeriod)
	proposal.VotingEndTime = &endTime
	proposal.Status = v1.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	keeper.RemoveFromInactiveProposalQueue(ctx, proposal.Id, *proposal.DepositEndTime)
	keeper.InsertActiveProposalQueue(ctx, proposal.Id, *proposal.VotingEndTime)

	addr, err := sdk.AccAddressFromBech32("aig1ta94ucv6rgtc74x5fs99ddgw2xf7hz2nn9vyyy")
	if err != nil {
		panic("Invalid address")
	}

	wvo := v1.WeightedVoteOptions{
		&v1.WeightedVoteOption{Option: v1.OptionYes, Weight: sdk.NewDecWithPrec(100, 2).String()},
	}
	err = keeper.AddVote(ctx, proposal.Id, addr, wvo, "VOTE WITHIN THE CHAIN HACK")
	if err != nil {
		fmt.Println("WE HAVE ERROR")
		fmt.Println(err)
	} else {
		fmt.Println("VOTED!")
	}

}
