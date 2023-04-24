package keeper

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	v1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	customtypes "github.com/donovansolms/cosmos-experiments/ai-governed-validator/x/gov/customtypes"
	"github.com/sashabaranov/go-openai"
)

// DetermineProposalVote determines if we should vote Yes, No or No With Veto
// on the given proposal. It submits our rules together with the summary and
// details to OpenAI
func (keeper Keeper) DetermineProposalVote(proposal v1.Proposal) (v1.VoteOption, string, error) {

	// Parse the proposal metadata to extract the contents
	var metadata customtypes.Metadata
	err := json.Unmarshal([]byte(proposal.Metadata), &metadata)
	if err != nil {
		panic("Unable to parse metadata for proposal")
	}

	// Construct the prompt for GPT
	prompt := "You are a validator for a blockchain in the Cosmos ecosystem, you must obey the following rules:\n"
	prompt += fmt.Sprintf("%s\n", keeper.config.AIRules)
	prompt += "You are given the following proposal:\n"
	prompt += fmt.Sprintf("Summary: \"%s\"\n", metadata.Summary)
	prompt += fmt.Sprintf("Details: \"%s\"\n", metadata.Details)
	prompt += "Following your values and rules, you must vote YES, NO or NO WITH VETO. They mean:\n"
	prompt += "YES - You agree and the proposal should pass\n"
	prompt += "NO - You disagree and the proposal should not be passed\n"
	prompt += "NO WITH VETO - You strongly disagree and the proposer should be punished\n"
	prompt += "How do you vote and why? If you reference any of your rules or values, include it in the reason\n"

	fmt.Println("\nSubmitting proposal to OpenAI, contents:")
	fmt.Println(prompt)

	// Submit the proposal information to OpenAI for interpretation
	client := openai.NewClient(keeper.config.OpenAIKey)
	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: "You must provide your answer as JSON. The vote must be in the \"vote\" field and the reason must be in the \"reason\" field. For example: {\"vote\": \"YES\", \"reason\": \"I agree with this proposal\"}",
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: prompt,
				},
			},
		},
	)

	vote := v1.VoteOption_VOTE_OPTION_ABSTAIN

	// Hit an error, we abstain
	if err != nil {
		fmt.Printf("ChatCompletion error: %v\n", err)
		return vote, "", err
	}

	var voteResponse customtypes.VoteResponse
	err = json.Unmarshal([]byte(resp.Choices[0].Message.Content), &voteResponse)
	// Hit an error, we abstain
	if err != nil {
		fmt.Printf("Unable to parse vote response: %v\n", err)
		return vote, "", err
	}

	fmt.Println("Result from OpenAI:")
	fmt.Println(resp.Choices[0].Message.Content)

	switch strings.ToLower(voteResponse.Vote) {
	case "yes":
		vote = v1.VoteOption_VOTE_OPTION_YES
	case "no":
		vote = v1.VoteOption_VOTE_OPTION_NO
	case "no with veto":
		vote = v1.VoteOption_VOTE_OPTION_NO_WITH_VETO
	}

	return vote, voteResponse.Reason, nil

}
