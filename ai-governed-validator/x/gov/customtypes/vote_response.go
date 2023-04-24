package customtypes

// VoteResponse is the result returned from the OpenAI API
type VoteResponse struct {
	Vote   string `json:"vote"`
	Reason string `json:"reason"`
}
