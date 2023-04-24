package customtypes

// Config is a config struct used for intialising the gov module to avoid using globals.
type Config struct {
	// MaxMetadataLen defines the maximum proposal metadata length.
	MaxMetadataLen uint64

	// OpenAIKey is the API key for OpenAI
	OpenAIKey string
	// AIRules is the rules that it should follow
	AIRules string
}

// DefaultConfig returns the default config for gov.
func DefaultConfig() Config {
	return Config{
		OpenAIKey:      "",
		AIRules:        "",
		MaxMetadataLen: 4096,
	}
}
