package keeper

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	"github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

//------------------------------------------------------------------------------
// msgServer
//------------------------------------------------------------------------------

type msgServer struct{ k Keeper }

// NewMsgServerImpl creates an implementation of the gov v1 MsgServer interface
// for the given keeper.
func NewMsgServerImpl(k Keeper) govv1.MsgServer {
	return &msgServer{k}
}

func (ms msgServer) SubmitProposal(goCtx context.Context, msg *govv1.MsgSubmitProposal) (*govv1.MsgSubmitProposalResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	fmt.Println("SUUUBMITTT PROP!!")

	proposalMsgs, err := msg.GetMsgs()
	if err != nil {
		return nil, err
	}

	proposal, err := ms.k.Keeper.SubmitProposal(ctx, proposalMsgs, msg.Metadata)
	if err != nil {
		return nil, err
	}

	bytes, err := proposal.Marshal()
	if err != nil {
		return nil, err
	}

	// ref: https://github.com/cosmos/cosmos-sdk/issues/9683
	ctx.GasMeter().ConsumeGas(
		3*ctx.KVGasConfig().WriteCostPerByte*uint64(len(bytes)),
		"submit proposal",
	)

	defer telemetry.IncrCounter(1, types.ModuleName, "proposal")

	proposer, _ := sdk.AccAddressFromBech32(msg.GetProposer())
	votingStarted, err := ms.k.AddDeposit(ctx, proposal.Id, proposer, msg.GetInitialDeposit())
	if err != nil {
		return nil, err
	}

	ctx.EventManager().EmitEvent(
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.AttributeValueCategory),
			sdk.NewAttribute(sdk.AttributeKeySender, msg.GetProposer()),
		),
	)

	if votingStarted {
		submitEvent := sdk.NewEvent(types.EventTypeSubmitProposal,
			sdk.NewAttribute(types.AttributeKeyVotingPeriodStart, fmt.Sprintf("%d", proposal.Id)),
		)

		ctx.EventManager().EmitEvent(submitEvent)
	}

	return &v1.MsgSubmitProposalResponse{
		ProposalId: proposal.Id,
	}, nil
}

func (ms msgServer) Vote(goCtx context.Context, msg *govv1.MsgVote) (*govv1.MsgVoteResponse, error) {
	return govkeeper.NewMsgServerImpl(ms.k.Keeper).Vote(goCtx, msg)
}

func (ms msgServer) ExecLegacyContent(goCtx context.Context, msg *govv1.MsgExecLegacyContent) (*govv1.MsgExecLegacyContentResponse, error) {
	return govkeeper.NewMsgServerImpl(ms.k.Keeper).ExecLegacyContent(goCtx, msg)
}

func (ms msgServer) VoteWeighted(goCtx context.Context, msg *govv1.MsgVoteWeighted) (*govv1.MsgVoteWeightedResponse, error) {
	return govkeeper.NewMsgServerImpl(ms.k.Keeper).VoteWeighted(goCtx, msg)
}

func (ms msgServer) Deposit(goCtx context.Context, msg *govv1.MsgDeposit) (*govv1.MsgDepositResponse, error) {
	return govkeeper.NewMsgServerImpl(ms.k.Keeper).Deposit(goCtx, msg)
}

//------------------------------------------------------------------------------
// legacyMsgServer
//------------------------------------------------------------------------------

type legacyMsgServer struct {
	govAcct string
	server  govv1.MsgServer
}

// NewMsgServerImpl creates an implementation of the gov v1 MsgServer interface
// for the given keeper.
func NewLegacyMsgServerImpl(govAcct string, server govv1.MsgServer) govv1beta1.MsgServer {
	return &legacyMsgServer{
		govAcct: govAcct,
		server:  server,
	}
}

func (k legacyMsgServer) SubmitProposal(goCtx context.Context, msg *govv1beta1.MsgSubmitProposal) (*govv1beta1.MsgSubmitProposalResponse, error) {
	content := msg.GetContent()

	contentMsg, err := govv1.NewLegacyContent(content, k.govAcct)
	if err != nil {
		return nil, fmt.Errorf("error converting legacy content into proposal message: %w", err)
	}

	// this part is different from the vanilla gov module:
	//
	// we compose the metadata string based on the legacy content, instead of
	// simply leaving it empty.
	//
	// this is necessary because of the metadata type check we implemented.
	metadata := types.ProposalMetadata{
		Title:   content.GetTitle(),
		Summary: content.GetDescription(),
	}
	metadataStr, err := json.Marshal(&metadata)
	if err != nil {
		return nil, err
	}

	proposal, err := govv1.NewMsgSubmitProposal(
		[]sdk.Msg{contentMsg},
		msg.InitialDeposit,
		msg.Proposer,
		string(metadataStr),
	)
	if err != nil {
		return nil, err
	}

	resp, err := k.server.SubmitProposal(goCtx, proposal)
	if err != nil {
		return nil, err
	}

	return &govv1beta1.MsgSubmitProposalResponse{ProposalId: resp.ProposalId}, nil
}

func (k legacyMsgServer) Vote(goCtx context.Context, msg *govv1beta1.MsgVote) (*govv1beta1.MsgVoteResponse, error) {
	_, err := k.server.Vote(goCtx, &govv1.MsgVote{
		ProposalId: msg.ProposalId,
		Voter:      msg.Voter,
		Option:     govv1.VoteOption(msg.Option),
	})
	if err != nil {
		return nil, err
	}

	return &govv1beta1.MsgVoteResponse{}, nil
}

func (k legacyMsgServer) VoteWeighted(goCtx context.Context, msg *govv1beta1.MsgVoteWeighted) (*govv1beta1.MsgVoteWeightedResponse, error) {
	opts := make([]*govv1.WeightedVoteOption, len(msg.Options))
	for idx, opt := range msg.Options {
		opts[idx] = &govv1.WeightedVoteOption{
			Option: govv1.VoteOption(opt.Option),
			Weight: opt.Weight.String(),
		}
	}

	_, err := k.server.VoteWeighted(goCtx, &govv1.MsgVoteWeighted{
		ProposalId: msg.ProposalId,
		Voter:      msg.Voter,
		Options:    opts,
	})
	if err != nil {
		return nil, err
	}

	return &govv1beta1.MsgVoteWeightedResponse{}, nil
}

func (k legacyMsgServer) Deposit(goCtx context.Context, msg *govv1beta1.MsgDeposit) (*govv1beta1.MsgDepositResponse, error) {
	_, err := k.server.Deposit(goCtx, &govv1.MsgDeposit{
		ProposalId: msg.ProposalId,
		Depositor:  msg.Depositor,
		Amount:     msg.Amount,
	})
	if err != nil {
		return nil, err
	}

	return &govv1beta1.MsgDepositResponse{}, nil
}
