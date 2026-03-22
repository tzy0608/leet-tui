package browser

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	_ "modernc.org/sqlite"
	"golang.org/x/crypto/pbkdf2"
)

// readChromium reads LeetCode cookies from a Chromium-based browser.
// Chrome/Brave/Edge store cookies in an SQLite DB with AES-128-CBC encryption.
// The encryption key is derived via PBKDF2 from a passphrase stored in macOS Keychain.
func readChromium(b BrowserType, site string) (*LeetCodeCookies, error) {
	dbPath := cookieDBPath(b)
	if dbPath == "" {
		return nil, ErrBrowserNotInstalled
	}
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		return nil, ErrBrowserNotInstalled
	}

	// Copy DB to temp file: Chrome may have it locked.
	tmp, err := copyToTemp(dbPath)
	if err != nil {
		return nil, fmt.Errorf("copy cookie db: %w", err)
	}
	defer os.Remove(tmp)

	// Fetch decryption key from macOS Keychain.
	key, err := chromiumKey(b)
	if err != nil {
		return nil, fmt.Errorf("get keychain key: %w", err)
	}

	db, err := sql.Open("sqlite", tmp+"?mode=ro")
	if err != nil {
		return nil, fmt.Errorf("open cookie db: %w", err)
	}
	defer db.Close()

	// Query for LEETCODE_SESSION and csrftoken.
	query := `
		SELECT name, value, encrypted_value
		FROM cookies
		WHERE host_key LIKE ?
		  AND name IN ('LEETCODE_SESSION', 'csrftoken')
		ORDER BY last_access_utc DESC`

	rows, err := db.Query(query, hostFilter(site))
	if err != nil {
		return nil, fmt.Errorf("query cookies: %w", err)
	}
	defer rows.Close()

	cookies := &LeetCodeCookies{Browser: b.String()}
	seen := map[string]bool{}

	for rows.Next() {
		var name string
		var value string
		var encryptedValue []byte

		if err := rows.Scan(&name, &value, &encryptedValue); err != nil {
			continue
		}
		if seen[name] {
			continue // take first (most recent) occurrence
		}
		seen[name] = true

		decrypted := value
		if len(encryptedValue) > 3 {
			d, err := decryptChromium(key, encryptedValue)
			if err == nil {
				decrypted = d
			}
		}

		switch name {
		case "LEETCODE_SESSION":
			cookies.Session = decrypted
		case "csrftoken":
			cookies.CSRFToken = decrypted
		}
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate cookies: %w", err)
	}

	if cookies.Session == "" {
		return nil, ErrNotFound
	}

	return cookies, nil
}

// chromiumKey retrieves and derives the AES key for cookie decryption.
func chromiumKey(b BrowserType) ([]byte, error) {
	service := keychainService(b)
	if service == "" {
		return nil, fmt.Errorf("no keychain service for %s", b)
	}

	// `security find-generic-password -w -s <service>` prints the password to stdout.
	// macOS may show a dialog asking for permission the first time.
	out, err := exec.Command("security", "find-generic-password", "-w", "-s", service).Output()
	if err != nil {
		return nil, fmt.Errorf("keychain lookup failed (%s): %w", service, err)
	}

	passphrase := strings.TrimSpace(string(out))
	if passphrase == "" {
		return nil, fmt.Errorf("empty passphrase from keychain")
	}

	// PBKDF2-SHA1: salt="saltysalt", iterations=1003, keyLen=16
	key := pbkdf2.Key([]byte(passphrase), []byte("saltysalt"), 1003, 16, sha1.New)
	return key, nil
}

// decryptChromium decrypts a Chrome/Arc/Brave/Edge v10/v11 encrypted cookie value.
//
// Standard Chrome format (macOS):
//   [v10|v11] + AES-128-CBC(key, IV=16_spaces, plaintext + PKCS7)
//
// Newer Arc format (and potentially newer Chromium versions):
//   [v10|v11] + 16-byte embedded IV + AES-128-CBC(key, embeddedIV, nonce(16) + plaintext + PKCS7)
//
// We try the standard format first; if the result isn't printable ASCII we
// fall back to the Arc/embedded-IV format.
func decryptChromium(key, encrypted []byte) (string, error) {
	if len(encrypted) < 3 {
		return "", fmt.Errorf("too short")
	}

	prefix := string(encrypted[:3])
	if prefix != "v10" && prefix != "v11" {
		// Old unencrypted format: value stored as plaintext.
		return string(encrypted), nil
	}

	// --- Attempt 1: Standard Chrome/macOS format ---
	// Format: "v10" + AES-CBC(key, IV=16_spaces, plaintext)
	if result, err := decryptCBC(key, encrypted[3:], bytes.Repeat([]byte{' '}, aes.BlockSize), 0); err == nil {
		if isValidCookieValue(result) {
			return result, nil
		}
	}

	// --- Attempt 2: Arc/embedded-IV format ---
	// Format: "v10" + embeddedIV(16) + AES-CBC(key, embeddedIV, nonce(16) + plaintext)
	if len(encrypted) >= 3+aes.BlockSize*2 {
		embeddedIV := encrypted[3 : 3+aes.BlockSize]
		ct := encrypted[3+aes.BlockSize:]
		if result, err := decryptCBC(key, ct, embeddedIV, aes.BlockSize); err == nil {
			if isValidCookieValue(result) {
				return result, nil
			}
		}
	}

	return "", fmt.Errorf("could not decrypt with any known format")
}

// decryptCBC performs AES-128-CBC decryption, strips PKCS7, and optionally
// skips skipPrefix bytes from the beginning of the plaintext (for the nonce).
func decryptCBC(key, ciphertext, iv []byte, skipPrefix int) (string, error) {
	if len(ciphertext) == 0 || len(ciphertext)%aes.BlockSize != 0 {
		return "", fmt.Errorf("invalid ciphertext length")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", fmt.Errorf("create cipher: %w", err)
	}

	plaintext := make([]byte, len(ciphertext))
	cipher.NewCBCDecrypter(block, iv).CryptBlocks(plaintext, ciphertext)

	// Strip leading nonce if requested.
	if skipPrefix > 0 {
		if len(plaintext) <= skipPrefix {
			return "", fmt.Errorf("plaintext too short after nonce skip")
		}
		plaintext = plaintext[skipPrefix:]
	}

	// Remove PKCS7 padding.
	plaintext, err = pkcs7Unpad(plaintext)
	if err != nil {
		return "", fmt.Errorf("unpad: %w", err)
	}

	return string(plaintext), nil
}

// isValidCookieValue checks whether a decrypted value looks like a real cookie.
// A valid cookie must consist entirely of printable ASCII characters.
// We inspect the first 32 bytes (or the full string if shorter) to avoid
// false positives where only the very first byte happens to be ASCII.
func isValidCookieValue(s string) bool {
	if len(s) == 0 {
		return false
	}
	check := s
	if len(check) > 32 {
		check = check[:32]
	}
	for _, b := range []byte(check) {
		if b < 0x20 || b > 0x7E {
			return false
		}
	}
	return true
}

// pkcs7Unpad removes PKCS7 padding from plaintext.
func pkcs7Unpad(data []byte) ([]byte, error) {
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data")
	}
	padLen := int(data[len(data)-1])
	if padLen == 0 || padLen > aes.BlockSize || padLen > len(data) {
		return nil, fmt.Errorf("invalid padding length: %d", padLen)
	}
	for _, b := range data[len(data)-padLen:] {
		if int(b) != padLen {
			return nil, fmt.Errorf("invalid padding bytes")
		}
	}
	return data[:len(data)-padLen], nil
}

// copyToTemp copies the cookie DB to a temp file to avoid SQLite lock issues
// when the browser is running.
func copyToTemp(src string) (string, error) {
	f, err := os.Open(src)
	if err != nil {
		return "", err
	}
	defer f.Close()

	tmp, err := os.CreateTemp("", "leet-tui-cookies-*.db")
	if err != nil {
		return "", err
	}
	defer tmp.Close()

	if _, err := io.Copy(tmp, f); err != nil {
		os.Remove(tmp.Name())
		return "", err
	}
	return tmp.Name(), nil
}
