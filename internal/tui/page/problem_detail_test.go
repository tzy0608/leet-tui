package page

import (
	"strings"
	"testing"
)

func TestHtmlToMarkdown_BasicTags(t *testing.T) {
	input := "<p>Hello <strong>world</strong></p>"
	got := htmlToMarkdown(input)
	if !strings.Contains(got, "**world**") {
		t.Errorf("expected bold markdown, got: %s", got)
	}
}

func TestHtmlToMarkdown_CodeBlock(t *testing.T) {
	input := "<pre><code>x = 1</code></pre>"
	got := htmlToMarkdown(input)
	if !strings.Contains(got, "```") {
		t.Errorf("expected code block, got: %s", got)
	}
	if !strings.Contains(got, "x = 1") {
		t.Errorf("expected code content, got: %s", got)
	}
}

func TestHtmlToMarkdown_MathLessThan(t *testing.T) {
	// a < b should NOT be treated as HTML tag and should not cause infinite loop
	input := "<p>Given a &lt; b and c &gt; d</p>"
	got := htmlToMarkdown(input)
	if !strings.Contains(got, "a < b") {
		t.Errorf("expected decoded HTML entities, got: %s", got)
	}
	if !strings.Contains(got, "c > d") {
		t.Errorf("expected decoded HTML entities, got: %s", got)
	}
}

func TestHtmlToMarkdown_BareLessThan(t *testing.T) {
	// A bare < that doesn't look like a tag should be preserved
	input := "x < 5 is true"
	got := htmlToMarkdown(input)
	if got != "x < 5 is true" {
		t.Errorf("expected preserved <, got: %s", got)
	}
}

func TestHtmlToMarkdown_NestedTags(t *testing.T) {
	input := "<div class=\"example\"><p>test</p></div>"
	got := htmlToMarkdown(input)
	if strings.Contains(got, "<") {
		t.Errorf("expected all HTML tags removed, got: %s", got)
	}
	if !strings.Contains(got, "test") {
		t.Errorf("expected content preserved, got: %s", got)
	}
}

func TestHtmlToMarkdown_SupSub(t *testing.T) {
	input := "2<sup>10</sup> and H<sub>2</sub>O"
	got := htmlToMarkdown(input)
	if !strings.Contains(got, "^(10)") {
		t.Errorf("expected superscript conversion, got: %s", got)
	}
	if !strings.Contains(got, "_(2)") {
		t.Errorf("expected subscript conversion, got: %s", got)
	}
}

func TestHtmlToMarkdown_Entities(t *testing.T) {
	input := "&amp; &quot;hello&#39;s&quot; &nbsp;end"
	got := htmlToMarkdown(input)
	if !strings.Contains(got, `& "hello's"`) {
		t.Errorf("expected decoded entities, got: %s", got)
	}
}

func TestTruncate_Unicode(t *testing.T) {
	s := "两数之和是一个经典问题"
	got := truncate(s, 6)
	runes := []rune(got)
	if len(runes) > 6 {
		t.Errorf("expected at most 6 runes, got %d: %s", len(runes), got)
	}
	if !strings.HasSuffix(got, "..") {
		t.Errorf("expected truncation suffix '..', got: %s", got)
	}
}

func TestTruncate_Short(t *testing.T) {
	s := "hello"
	got := truncate(s, 10)
	if got != "hello" {
		t.Errorf("expected no truncation, got: %s", got)
	}
}
