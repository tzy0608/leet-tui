// Package browser provides utilities to read LeetCode session cookies
// directly from installed browsers on macOS.
package browser

import (
	"errors"
	"fmt"
	"os"
)

// LeetCodeCookies holds the two cookies needed for LeetCode API access.
type LeetCodeCookies struct {
	Session   string // LEETCODE_SESSION
	CSRFToken string // csrftoken
	Browser   string // source browser name
}

// BrowserType identifies a supported browser.
type BrowserType int

const (
	BrowserArc     BrowserType = iota // Arc is first: highest priority for this user
	BrowserChrome
	BrowserBrave
	BrowserEdge
	BrowserFirefox
)

func (b BrowserType) String() string {
	switch b {
	case BrowserArc:
		return "Arc"
	case BrowserChrome:
		return "Chrome"
	case BrowserBrave:
		return "Brave"
	case BrowserEdge:
		return "Edge"
	case BrowserFirefox:
		return "Firefox"
	default:
		return "Unknown"
	}
}

// ErrNotFound is returned when no LeetCode cookies are found in a browser.
var ErrNotFound = errors.New("leetcode session not found in browser")

// ErrBrowserNotInstalled is returned when the browser is not installed.
var ErrBrowserNotInstalled = errors.New("browser not installed")

// ReadFromBrowser reads LeetCode cookies from the specified browser.
func ReadFromBrowser(b BrowserType, site string) (*LeetCodeCookies, error) {
	switch b {
	case BrowserArc, BrowserChrome, BrowserBrave, BrowserEdge:
		return readChromium(b, site)
	case BrowserFirefox:
		return readFirefox(site)
	default:
		return nil, fmt.Errorf("unsupported browser: %v", b)
	}
}

// AutoRead tries each installed browser in order and returns the first
// successful LeetCode session found.
// Priority: Arc → Chrome → Brave → Edge → Firefox
func AutoRead(site string) (*LeetCodeCookies, error) {
	browsers := []BrowserType{BrowserArc, BrowserChrome, BrowserBrave, BrowserEdge, BrowserFirefox}

	var errs []error
	for _, b := range browsers {
		cookies, err := ReadFromBrowser(b, site)
		if err == nil && cookies.Session != "" {
			return cookies, nil
		}
		if err != nil && !errors.Is(err, ErrBrowserNotInstalled) {
			errs = append(errs, fmt.Errorf("%s: %w", b, err))
		}
	}

	if len(errs) > 0 {
		return nil, fmt.Errorf("no browser with LeetCode session found: %v", errs)
	}
	return nil, ErrNotFound
}

// InstalledBrowsers returns a list of browsers that appear to be installed.
func InstalledBrowsers() []BrowserType {
	var result []BrowserType
	for _, b := range []BrowserType{BrowserArc, BrowserChrome, BrowserBrave, BrowserEdge, BrowserFirefox} {
		path := cookieDBPath(b)
		if path != "" {
			if _, err := os.Stat(path); err == nil {
				result = append(result, b)
			}
		}
	}
	return result
}

// cookieDBPath returns the SQLite cookie database path for each browser on macOS.
func cookieDBPath(b BrowserType) string {
	home, _ := os.UserHomeDir()
	base := home + "/Library/Application Support/"
	switch b {
	case BrowserArc:
		return base + "Arc/User Data/Default/Cookies"
	case BrowserChrome:
		return base + "Google/Chrome/Default/Cookies"
	case BrowserBrave:
		return base + "BraveSoftware/Brave-Browser/Default/Cookies"
	case BrowserEdge:
		return base + "Microsoft Edge/Default/Cookies"
	case BrowserFirefox:
		return firefoxCookiePath(home)
	}
	return ""
}

// keychainService returns the macOS Keychain service name for each Chromium browser.
func keychainService(b BrowserType) string {
	switch b {
	case BrowserArc:
		return "Arc Safe Storage"
	case BrowserChrome:
		return "Chrome Safe Storage"
	case BrowserBrave:
		return "Brave Safe Storage"
	case BrowserEdge:
		return "Microsoft Edge Safe Storage"
	}
	return ""
}

// hostFilter returns the SQL host_key pattern for the given LeetCode site.
func hostFilter(site string) string {
	if site == "cn" {
		return "%leetcode.cn%"
	}
	return "%leetcode.com%"
}
