package browser

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite"
)

// readFirefox reads LeetCode cookies from Firefox.
// Firefox stores cookies in plaintext in an SQLite database.
func readFirefox(site string) (*LeetCodeCookies, error) {
	dbPath := cookieDBPath(BrowserFirefox)
	if dbPath == "" {
		return nil, ErrBrowserNotInstalled
	}
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, ErrBrowserNotInstalled
	}

	tmp, err := copyToTemp(dbPath)
	if err != nil {
		return nil, fmt.Errorf("copy cookie db: %w", err)
	}
	defer os.Remove(tmp)

	db, err := sql.Open("sqlite", tmp+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("open cookie db: %w", err)
	}
	defer db.Close()

	// Firefox schema: moz_cookies(name, value, host, ...)
	query := `
		SELECT name, value
		FROM moz_cookies
		WHERE host LIKE ?
		  AND name IN ('LEETCODE_SESSION', 'csrftoken')
		ORDER BY lastAccessed DESC`

	rows, err := db.Query(query, hostFilter(site))
	if err != nil {
		return nil, fmt.Errorf("query cookies: %w", err)
	}
	defer rows.Close()

	cookies := &LeetCodeCookies{Browser: "Firefox"}
	seen := map[string]bool{}

	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			continue
		}
		if seen[name] {
			continue
		}
		seen[name] = true

		switch name {
		case "LEETCODE_SESSION":
			cookies.Session = value
		case "csrftoken":
			cookies.CSRFToken = value
		}
	}

	if cookies.Session == "" {
		return nil, ErrNotFound
	}

	return cookies, nil
}

// firefoxCookiePath finds the Firefox profile directory and returns the cookie DB path.
func firefoxCookiePath(home string) string {
	profilesDir := filepath.Join(home, "Library", "Application Support", "Firefox", "Profiles")

	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return ""
	}

	// Prefer the default release profile (*.default-release) over other profiles.
	var fallback string
	for _, e := range entries {
		if !e.IsDir() {
			continue
		}
		cookiePath := filepath.Join(profilesDir, e.Name(), "cookies.sqlite")
		if _, err := os.Stat(cookiePath); err == nil {
			if strings.HasSuffix(e.Name(), ".default-release") {
				return cookiePath
			}
			if strings.Contains(e.Name(), ".default") {
				return cookiePath
			}
			fallback = cookiePath
		}
	}
	return fallback
}
