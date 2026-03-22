package leetcode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client defines the LeetCode API interface.
type Client interface {
	// FetchProblemList retrieves problems in batches.
	FetchProblemList(ctx context.Context, skip, limit int) ([]ProblemListItem, int, error)
	// FetchProblemDetail retrieves full problem details.
	FetchProblemDetail(ctx context.Context, titleSlug string) (*Problem, error)
	// RunCode runs code against test cases.
	RunCode(ctx context.Context, titleSlug string, questionID int, lang, code, testInput string) (*RunResult, error)
	// SubmitCode submits a solution.
	SubmitCode(ctx context.Context, titleSlug string, questionID int, lang, code string) (*SubmitResult, error)
	// FetchLatestAcceptedCode retrieves the code from the most recent Accepted submission for the given problem and language.
	FetchLatestAcceptedCode(ctx context.Context, titleSlug, lang string) (string, error)
	// FetchRecentAccepted checks if the problem has an Accepted submission within the given number of days.
	FetchRecentAccepted(ctx context.Context, titleSlug string, withinDays int) (bool, error)
}

type graphQLClient struct {
	httpClient *http.Client
	baseURL    string
	cookie     string
	csrfToken  string
}

// NewClient creates a LeetCode API client for the given site.
func NewClient(site, cookie, csrfToken string) Client {
	base := &graphQLClient{
		httpClient: &http.Client{Timeout: 30 * time.Second},
		baseURL:    "https://leetcode.com",
		cookie:     cookie,
		csrfToken:  csrfToken,
	}
	if site == "cn" {
		base.baseURL = "https://leetcode.cn"
		return &cnClient{graphQLClient: *base}
	}
	return base
}

func (c *graphQLClient) graphQL(ctx context.Context, query string, variables map[string]any) (map[string]any, error) {
	body := map[string]any{
		"query":     query,
		"variables": variables,
	}
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/graphql", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", c.baseURL)
	req.Header.Set("Origin", c.baseURL)
	if c.cookie != "" {
		req.Header.Set("Cookie", fmt.Sprintf("LEETCODE_SESSION=%s; csrftoken=%s", c.cookie, c.csrfToken))
	}
	if c.csrfToken != "" {
		req.Header.Set("X-CSRFToken", c.csrfToken)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var gqlResp graphQLResponse
	if err := json.Unmarshal(respBody, &gqlResp); err != nil {
		return nil, fmt.Errorf("unmarshal response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("graphql error: %s", gqlResp.Errors[0].Message)
	}

	return gqlResp.Data, nil
}
