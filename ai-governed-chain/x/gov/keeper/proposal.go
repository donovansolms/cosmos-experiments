package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
)

func (keeper Keeper) ActivateVotingPeriod(ctx sdk.Context, proposal v1.Proposal) {

	startTime := ctx.BlockHeader().Time
	proposal.VotingStartTime = &startTime
	votingPeriod := keeper.GetVotingParams(ctx).VotingPeriod
	endTime := proposal.VotingStartTime.Add(*votingPeriod)
	proposal.VotingEndTime = &endTime
	proposal.Status = v1.StatusVotingPeriod
	keeper.SetProposal(ctx, proposal)

	keeper.RemoveFromInactiveProposalQueue(ctx, proposal.Id, *proposal.DepositEndTime)
	keeper.InsertActiveProposalQueue(ctx, proposal.Id, *proposal.VotingEndTime)

	// We're hacking in a vote from our validator address
	// This will fail consensus and is done on purpose so that no one gets
	// any ideas for integrating this into a real chain
	addr, err := sdk.AccAddressFromBech32("aig1ta94ucv6rgtc74x5fs99ddgw2xf7hz2nn9vyyy")
	if err != nil {
		panic("Invalid address")
	}

	vote, reason, err := keeper.DetermineProposalVote(proposal)
	if err != nil {
		panic(fmt.Sprintf("Unable to determine vote: %s", err))
	}

	weightedVote := v1.WeightedVoteOptions{
		&v1.WeightedVoteOption{Option: vote, Weight: sdk.NewDecWithPrec(100, 2).String()},
	}
	err = keeper.AddVote(ctx, proposal.Id, addr, weightedVote, reason)
	if err != nil {
		panic(fmt.Sprintf("Unable to vote: %s", err))
	}
	fmt.Println("\n=====\nVoting completed\n=====")
}
