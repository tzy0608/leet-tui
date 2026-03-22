package leetcode

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"
)

func (c *graphQLClient) FetchProblemList(ctx context.Context, skip, limit int) ([]ProblemListItem, int, error) {
	vars := map[string]any{
		"categorySlug": "all-code-essentials",
		"skip":         skip,
		"limit":        limit,
		"filters":      map[string]any{},
	}

	data, err := c.graphQL(ctx, queryProblemList, vars)
	if err != nil {
		return nil, 0, fmt.Errorf("fetch problem list: %w", err)
	}

	pqList, ok := data["problemsetQuestionList"].(map[string]any)
	if !ok {
		return nil, 0, fmt.Errorf("unexpected response format")
	}

	total := intVal(pqList, "total")
	questions, ok := pqList["questions"].([]any)
	if !ok {
		return nil, 0, fmt.Errorf("unexpected questions format")
	}

	var items []ProblemListItem
	for _, q := range questions {
		qm, ok := q.(map[string]any)
		if !ok {
			continue
		}
		item := ProblemListItem{
			Title:      str(qm, "title"),
			TitleSlug:  str(qm, "titleSlug"),
			FrontendID: str(qm, "questionFrontendId"),
			Difficulty: str(qm, "difficulty"),
			IsPaidOnly: boolVal(qm, "isPaidOnly"),
			AcRate:     floatVal(qm, "acRate"),
			Status:     str(qm, "status"),
		}
		if id, err := strconv.Atoi(str(qm, "questionId")); err == nil {
			item.ID = id
		}
		if tags, ok := qm["topicTags"].([]any); ok {
			for _, t := range tags {
				if tm, ok := t.(map[string]any); ok {
					item.TopicTags = append(item.TopicTags, str(tm, "name"))
				}
			}
		}
		items = append(items, item)
	}

	return items, total, nil
}

func (c *graphQLClient) FetchProblemDetail(ctx context.Context, titleSlug string) (*Problem, error) {
	vars := map[string]any{"titleSlug": titleSlug}
	data, err := c.graphQL(ctx, queryProblemDetail, vars)
	if err != nil {
		return nil, fmt.Errorf("fetch problem detail: %w", err)
	}

	q, ok := data["question"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	p := &Problem{
		Title:      str(q, "title"),
		TitleSlug:  str(q, "titleSlug"),
		FrontendID: str(q, "questionFrontendId"),
		Difficulty: str(q, "difficulty"),
		Content:    str(q, "content"),
		IsPaidOnly:     boolVal(q, "isPaidOnly"),
		SampleTestCase: str(q, "sampleTestCase"),
	}
	if id, err := strconv.Atoi(str(q, "questionId")); err == nil {
		p.ID = id
	}

	if tags, ok := q["topicTags"].([]any); ok {
		for _, t := range tags {
			if tm, ok := t.(map[string]any); ok {
				p.TopicTags = append(p.TopicTags, str(tm, "name"))
			}
		}
	}

	p.CodeSnippets = make(map[string]string)
	if snippets, ok := q["codeSnippets"].([]any); ok {
		for _, s := range snippets {
			if sm, ok := s.(map[string]any); ok {
				p.CodeSnippets[str(sm, "langSlug")] = str(sm, "code")
			}
		}
	}

	return p, nil
}

func (c *graphQLClient) RunCode(ctx context.Context, titleSlug string, questionID int, lang, code, testInput string) (*RunResult, error) {
	// RunCode uses REST API, not GraphQL
	runResp, err := c.postJSON(ctx, fmt.Sprintf("/problems/%s/interpret_solution/", titleSlug), map[string]any{
		"lang":        lang,
		"question_id": questionID,
		"typed_code":  code,
		"data_input":  testInput,
	})
	if err != nil {
		return nil, fmt.Errorf("run code: %w", err)
	}

	interpretID := str(runResp, "interpret_id")
	if interpretID == "" {
		return nil, fmt.Errorf("no interpret_id in response")
	}

	return c.pollRunResult(ctx, interpretID)
}

func (c *graphQLClient) SubmitCode(ctx context.Context, titleSlug string, questionID int, lang, code string) (*SubmitResult, error) {
	submitResp, err := c.postJSON(ctx, fmt.Sprintf("/problems/%s/submit/", titleSlug), map[string]any{
		"lang":        lang,
		"question_id": questionID,
		"typed_code":  code,
	})
	if err != nil {
		return nil, fmt.Errorf("submit code: %w", err)
	}

	submissionID := int(floatVal(submitResp, "submission_id"))
	if submissionID == 0 {
		return nil, fmt.Errorf("no submission_id in response")
	}

	return c.pollSubmitResult(ctx, submissionID)
}

func (c *graphQLClient) FetchLatestAcceptedCode(ctx context.Context, titleSlug, lang string) (string, error) {
	// Fetch recent submissions for this problem, filter client-side
	vars := map[string]any{
		"questionSlug": titleSlug,
		"offset":       0,
		"limit":        20,
	}
	data, err := c.graphQL(ctx, querySubmissionList, vars)
	if err != nil {
		return "", fmt.Errorf("fetch submission list: %w", err)
	}

	qsl, ok := data["questionSubmissionList"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("unexpected submission list format")
	}
	submissions, ok := qsl["submissions"].([]any)
	if !ok || len(submissions) == 0 {
		return "", nil
	}

	// Find the first Accepted submission matching the requested language
	var subIDStr string
	for _, s := range submissions {
		sm, ok := s.(map[string]any)
		if !ok {
			continue
		}
		if str(sm, "statusDisplay") == "Accepted" && str(sm, "lang") == lang {
			subIDStr = str(sm, "id")
			break
		}
	}
	if subIDStr == "" {
		return "", nil
	}

	subID, err := strconv.Atoi(subIDStr)
	if err != nil {
		return "", fmt.Errorf("parse submission id: %w", err)
	}

	// Fetch submission detail to get the code
	detailVars := map[string]any{"submissionId": subID}
	detailData, err := c.graphQL(ctx, querySubmissionDetail, detailVars)
	if err != nil {
		return "", fmt.Errorf("fetch submission detail: %w", err)
	}

	detail, ok := detailData["submissionDetails"].(map[string]any)
	if !ok {
		return "", fmt.Errorf("unexpected submission detail format")
	}

	return str(detail, "code"), nil
}

func (c *graphQLClient) FetchRecentAccepted(ctx context.Context, titleSlug string, withinDays int) (bool, error) {
	vars := map[string]any{
		"questionSlug": titleSlug,
		"offset":       0,
		"limit":        20,
	}
	data, err := c.graphQL(ctx, querySubmissionList, vars)
	if err != nil {
		return false, fmt.Errorf("fetch submission list: %w", err)
	}

	qsl, ok := data["questionSubmissionList"].(map[string]any)
	if !ok {
		return false, fmt.Errorf("unexpected submission list format")
	}
	submissions, ok := qsl["submissions"].([]any)
	if !ok || len(submissions) == 0 {
		return false, nil
	}

	cutoff := time.Now().AddDate(0, 0, -withinDays)
	for _, s := range submissions {
		sm, ok := s.(map[string]any)
		if !ok {
			continue
		}
		if str(sm, "statusDisplay") != "Accepted" {
			continue
		}
		ts, err := strconv.ParseInt(str(sm, "timestamp"), 10, 64)
		if err != nil {
			continue
		}
		if time.Unix(ts, 0).After(cutoff) {
			return true, nil
		}
	}
	return false, nil
}

func (c *graphQLClient) postJSON(ctx context.Context, path string, body map[string]any) (map[string]any, error) {
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request body: %w", err)
	}
	req, err := newJSONRequest(ctx, "POST", c.baseURL+path, jsonBody)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	req.Header.Set("Referer", c.baseURL+path)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}

	var result map[string]any
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}
	return result, nil
}

func (c *graphQLClient) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", c.baseURL)
	req.Header.Set("Origin", c.baseURL)
	if c.cookie != "" {
		req.Header.Set("Cookie", fmt.Sprintf("LEETCODE_SESSION=%s; csrftoken=%s", c.cookie, c.csrfToken))
	}
	if c.csrfToken != "" {
		req.Header.Set("X-CSRFToken", c.csrfToken)
	}
}

func newJSONRequest(ctx context.Context, method, url string, body []byte) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	return req, nil
}

func (c *graphQLClient) pollRunResult(ctx context.Context, interpretID string) (*RunResult, error) {
	url := fmt.Sprintf("%s/submissions/detail/%s/check/", c.baseURL, interpretID)
	for i := 0; i < 30; i++ {
		time.Sleep(time.Second)

		result, err := func() (map[string]any, error) {
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return nil, err
			}
			c.setHeaders(req)

			resp, err := c.httpClient.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("read response: %w", err)
			}

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
			}

			var r map[string]any
			if err := json.Unmarshal(respBody, &r); err != nil {
				return nil, fmt.Errorf("decode response: %w", err)
			}
			return r, nil
		}()
		if err != nil {
			return nil, err
		}

		state := str(result, "state")
		if state == "SUCCESS" {
			return &RunResult{
				StatusCode:   intVal(result, "status_code"),
				StatusMsg:    str(result, "status_msg"),
				RunSuccess:   boolVal(result, "run_success"),
				CompileError: str(result, "full_compile_error"),
				RuntimeError: str(result, "full_runtime_error"),
				CodeAnswer:   strSlice(result, "code_answer"),
				ExpectedAns:  strSlice(result, "expected_code_answer"),
			}, nil
		}
	}
	return nil, fmt.Errorf("polling timed out")
}

func (c *graphQLClient) pollSubmitResult(ctx context.Context, submissionID int) (*SubmitResult, error) {
	url := fmt.Sprintf("%s/submissions/detail/%d/check/", c.baseURL, submissionID)
	for i := 0; i < 30; i++ {
		time.Sleep(time.Second)

		result, err := func() (map[string]any, error) {
			req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
			if err != nil {
				return nil, err
			}
			c.setHeaders(req)

			resp, err := c.httpClient.Do(req)
			if err != nil {
				return nil, err
			}
			defer resp.Body.Close()

			respBody, err := io.ReadAll(resp.Body)
			if err != nil {
				return nil, fmt.Errorf("read response: %w", err)
			}

			if resp.StatusCode != http.StatusOK {
				return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
			}

			var r map[string]any
			if err := json.Unmarshal(respBody, &r); err != nil {
				return nil, fmt.Errorf("decode response: %w", err)
			}
			return r, nil
		}()
		if err != nil {
			return nil, err
		}

		state := str(result, "state")
		if state == "SUCCESS" {
			return &SubmitResult{
				StatusCode:    intVal(result, "status_code"),
				StatusMsg:     str(result, "status_msg"),
				RunSuccess:    boolVal(result, "run_success"),
				TotalCorrect:  intVal(result, "total_correct"),
				TotalTestcase: intVal(result, "total_testcases"),
				RuntimeMs:     str(result, "status_runtime"),
				MemoryMB:      str(result, "status_memory"),
				CompileError:  str(result, "full_compile_error"),
				RuntimeError:      str(result, "full_runtime_error"),
			RuntimePercentile: floatVal(result, "runtime_percentile"),
			MemoryPercentile:  floatVal(result, "memory_percentile"),
			}, nil
		}
	}
	return nil, fmt.Errorf("polling timed out")
}

// Helper functions for safe type assertions
func str(m map[string]any, key string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func intVal(m map[string]any, key string) int {
	if v, ok := m[key]; ok && v != nil {
		if f, ok := v.(float64); ok {
			return int(f)
		}
	}
	return 0
}

func floatVal(m map[string]any, key string) float64 {
	if v, ok := m[key]; ok && v != nil {
		if f, ok := v.(float64); ok {
			return f
		}
	}
	return 0
}

func boolVal(m map[string]any, key string) bool {
	if v, ok := m[key]; ok && v != nil {
		if b, ok := v.(bool); ok {
			return b
		}
	}
	return false
}

func strSlice(m map[string]any, key string) []string {
	if v, ok := m[key]; ok && v != nil {
		if arr, ok := v.([]any); ok {
			var result []string
			for _, a := range arr {
				if s, ok := a.(string); ok {
					result = append(result, s)
				}
			}
			return result
		}
	}
	return nil
}
