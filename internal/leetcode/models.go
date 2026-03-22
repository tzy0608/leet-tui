package leetcode

// Problem represents a LeetCode problem.
type Problem struct {
	ID               int
	Title            string
	TitleSlug        string
	FrontendID       string
	Difficulty       string
	Content          string
	TopicTags        []string
	CodeSnippets     map[string]string // lang -> code
	IsPaidOnly       bool
	AcRate           float64
	LastAcceptedCode string // most recent accepted submission code
	SampleTestCase   string
}

// ProblemListItem is a compact representation for listing.
type ProblemListItem struct {
	ID         int
	Title      string
	TitleSlug  string
	FrontendID string
	Difficulty string
	IsPaidOnly bool
	AcRate     float64
	TopicTags  []string
	Status     string // "ac", "notac", ""
}

// SubmitResult holds the result of a code submission.
type SubmitResult struct {
	StatusCode        int
	StatusMsg         string
	RunSuccess        bool
	TotalCorrect      int
	TotalTestcase     int
	RuntimeMs         string
	MemoryMB          string
	CompileError      string
	RuntimeError      string
	RuntimePercentile float64
	MemoryPercentile  float64
}

// RunResult holds the result of running code against test cases.
type RunResult struct {
	StatusCode    int
	StatusMsg     string
	RunSuccess    bool
	CodeAnswer    []string
	ExpectedAns   []string
	CompileError  string
	RuntimeError  string
}

// TopicTag represents a topic/tag.
type TopicTag struct {
	Name string
	Slug string
}

// GraphQL response types
type graphQLResponse struct {
	Data   map[string]any `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}
