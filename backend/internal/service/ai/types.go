package ai

type HexagramInfo struct {
	Name       string
	FullName   string
	Summary    string
	BinaryCode string
}

type GenerateInput struct {
	DivinationID    int64
	Question        string
	CategoryName    string
	PrimaryHexagram HexagramInfo
	ChangedHexagram HexagramInfo
	MovingLines     []int
	LineSnapshot    string
	FreeContent     string
}

type GenerateOutput struct {
	Provider              string
	Content               string
	RawResponse           string
	ModelName             string
	DurationMs            int64
	FallbackUsed          int
	ErrorMessage          string
	PromptCacheHitTokens  int
	PromptCacheMissTokens int
}
