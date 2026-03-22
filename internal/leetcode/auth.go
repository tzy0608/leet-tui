package leetcode

import (
	"context"
	"fmt"
	"net/http"
)

// ValidateSession checks if the current cookie is valid by attempting a query.
func ValidateSession(ctx context.Context, client Client) (bool, error) {
	// Try fetching a single problem to validate credentials
	_, _, err := client.FetchProblemList(ctx, 0, 1)
	if err != nil {
		return false, err
	}
	return true, nil
}

// ExtractCSRF extracts CSRF token from leetcode.com by visiting the homepage.
func ExtractCSRF(site string) (string, error) {
	baseURL := "https://leetcode.com"
	if site == "cn" {
		baseURL = "https://leetcode.cn"
	}

	resp, err := http.Get(baseURL)
	if err != nil {
		return "", fmt.Errorf("fetch homepage: %w", err)
	}
	defer resp.Body.Close()

	for _, cookie := range resp.Cookies() {
		if cookie.Name == "csrftoken" {
			return cookie.Value, nil
		}
	}

	return "", fmt.Errorf("csrftoken not found in cookies")
}
