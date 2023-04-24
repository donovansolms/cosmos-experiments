package customtypes

// Metadata is the proposal metadata
type Metadata struct {
	Title             string `json:"title"`
	Authors           string `json:"authors"`
	Summary           string `json:"summary"`
	Details           string `json:"details"`
	ProposalForumURL  string `json:"proposal_forum_url"`
	VoteOptionContext string `json:"vote_option_context"`
}
