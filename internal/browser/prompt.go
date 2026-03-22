package browser

import (
	"fmt"
	"strings"
)

// PickBrowser interactively prompts the user to choose a browser.
// Returns the selected browser type and whether the user cancelled.
func PickBrowser() (BrowserType, bool) {
	installed := InstalledBrowsers()
	if len(installed) == 0 {
		fmt.Println("No supported browsers found.")
		return 0, true
	}

	fmt.Println("\nDetected browsers:")
	for i, b := range installed {
		fmt.Printf("  [%d] %s\n", i+1, b)
	}
	fmt.Printf("  [0] Enter manually\n")
	fmt.Print("\nSelect browser (default: 1): ")

	var input string
	fmt.Scanln(&input)
	input = strings.TrimSpace(input)

	if input == "" || input == "1" {
		return installed[0], false
	}
	if input == "0" {
		return 0, true
	}

	idx := 0
	fmt.Sscanf(input, "%d", &idx)
	if idx < 1 || idx > len(installed) {
		fmt.Println("Invalid selection.")
		return 0, true
	}
	return installed[idx-1], false
}

// LoginInteractive guides the user through browser cookie extraction.
// Returns the extracted cookies, or nil if the user chose manual entry.
func LoginInteractive(site string) (*LeetCodeCookies, error) {
	fmt.Println("=== leet-tui Login ===")
	fmt.Printf("Reading LeetCode%s cookies from your browser...\n",
		map[string]string{"cn": ".cn", "us": ".com"}[site])

	installed := InstalledBrowsers()
	if len(installed) == 0 {
		return nil, fmt.Errorf("no supported browsers installed")
	}

	fmt.Printf("\nFound %d browser(s):\n", len(installed))
	for i, b := range installed {
		fmt.Printf("  [%d] %s\n", i+1, b)
	}
	fmt.Printf("  [0] Enter cookie manually\n")
	fmt.Print("\nSelect [1]: ")

	var input string
	fmt.Scanln(&input)
	input = strings.TrimSpace(input)
	if input == "" {
		input = "1"
	}

	if input == "0" {
		return nil, nil // caller handles manual entry
	}

	idx := 0
	fmt.Sscanf(input, "%d", &idx)
	if idx < 1 || idx > len(installed) {
		return nil, fmt.Errorf("invalid selection")
	}

	chosen := installed[idx-1]
	fmt.Printf("\nReading from %s", chosen)

	// Chromium browsers require Keychain access — warn the user.
	if chosen != BrowserFirefox {
		fmt.Printf("\n⚠  macOS may show a dialog asking permission to access Keychain.\n")
		fmt.Printf("   Please click \"Allow\" when prompted.\n")
	}

	fmt.Printf("Reading cookies...")
	cookies, err := ReadFromBrowser(chosen, site)
	if err != nil {
		fmt.Println(" ✗")
		return nil, fmt.Errorf("read %s cookies: %w", chosen, err)
	}
	fmt.Println(" ✓")

	return cookies, nil
}

// PrintCookieStatus prints a summary of the found cookies (masking sensitive data).
func PrintCookieStatus(cookies *LeetCodeCookies) {
	fmt.Printf("Browser:       %s\n", cookies.Browser)
	fmt.Printf("Session:       %s\n", maskCookie(cookies.Session))
	fmt.Printf("CSRF Token:    %s\n", maskCookie(cookies.CSRFToken))
}

// maskCookie shows only the first 8 and last 4 characters of a cookie value.
func maskCookie(s string) string {
	if len(s) <= 12 {
		return strings.Repeat("*", len(s))
	}
	return s[:8] + strings.Repeat("*", len(s)-12) + s[len(s)-4:]
}
