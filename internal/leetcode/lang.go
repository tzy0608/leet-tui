package leetcode

import "strings"

// LangSlug maps a user-friendly language name to the LeetCode API langSlug.
func LangSlug(lang string) string {
	switch strings.ToLower(lang) {
	case "go":
		return "golang"
	case "python":
		return "python3"
	case "c++":
		return "cpp"
	case "c#":
		return "csharp"
	default:
		return strings.ToLower(lang)
	}
}
