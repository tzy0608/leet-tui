package leetcode

import (
	"context"
	"fmt"
)

// cnClient implements Client for leetcode.cn (not yet implemented).
type cnClient struct {
	graphQLClient
}

var errCNNotImplemented = fmt.Errorf("leetcode.cn site is not yet implemented; please use site=us")

func (c *cnClient) FetchProblemList(ctx context.Context, skip, limit int) ([]ProblemListItem, int, error) {
	return nil, 0, errCNNotImplemented
}

func (c *cnClient) FetchProblemDetail(ctx context.Context, titleSlug string) (*Problem, error) {
	return nil, errCNNotImplemented
}

func (c *cnClient) RunCode(ctx context.Context, titleSlug string, questionID int, lang, code, testInput string) (*RunResult, error) {
	return nil, errCNNotImplemented
}

func (c *cnClient) SubmitCode(ctx context.Context, titleSlug string, questionID int, lang, code string) (*SubmitResult, error) {
	return nil, errCNNotImplemented
}

func (c *cnClient) FetchLatestAcceptedCode(ctx context.Context, titleSlug, lang string) (string, error) {
	return "", errCNNotImplemented
}

func (c *cnClient) FetchRecentAccepted(ctx context.Context, titleSlug string, withinDays int) (bool, error) {
	return false, nil
}
