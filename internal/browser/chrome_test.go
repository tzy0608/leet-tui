package browser

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha1"
	"testing"

	"golang.org/x/crypto/pbkdf2"
)

func TestPkcs7Unpad(t *testing.T) {
	tests := []struct {
		name    string
		input   []byte
		want    []byte
		wantErr bool
	}{
		{
			name:  "valid padding 4",
			input: []byte("hello\x04\x04\x04\x04"),
			want:  []byte("hello"),
		},
		{
			name:  "valid padding 1",
			input: []byte("hi\x01"),
			want:  []byte("hi"),
		},
		{
			name:    "empty input",
			input:   []byte{},
			wantErr: true,
		},
		{
			name:    "invalid padding",
			input:   []byte("hi\x03\x04"), // last byte says 4 but only 1 pad byte follows
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := pkcs7Unpad(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("pkcs7Unpad() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && string(got) != string(tt.want) {
				t.Errorf("pkcs7Unpad() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestEncryptDecryptRoundtrip verifies that our decryption logic works
// by encrypting a known value and then decrypting it.
func TestEncryptDecryptRoundtrip(t *testing.T) {
	passphrase := "test-passphrase"
	plaintext := "LEETCODE_SESSION=abcdef12345"

	// Derive key (same as chromiumKey)
	key := pbkdf2.Key([]byte(passphrase), []byte("saltysalt"), 1003, 16, sha1.New)

	// Encrypt (simulate Chrome's encryption)
	encrypted, err := encryptChromium(key, plaintext)
	if err != nil {
		t.Fatalf("encrypt: %v", err)
	}

	// Prepend v10 prefix
	withPrefix := append([]byte("v10"), encrypted...)

	// Decrypt using our function
	got, err := decryptChromium(key, withPrefix)
	if err != nil {
		t.Fatalf("decryptChromium: %v", err)
	}

	if got != plaintext {
		t.Errorf("roundtrip failed: got %q, want %q", got, plaintext)
	}
}

// encryptChromium is a test-only helper that encrypts data using the same
// scheme Chrome uses, so we can test our decryption logic.
func encryptChromium(key []byte, plaintext string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	// PKCS7 pad to block size
	padLen := aes.BlockSize - (len(plaintext) % aes.BlockSize)
	padded := append([]byte(plaintext), bytes.Repeat([]byte{byte(padLen)}, padLen)...)

	ciphertext := make([]byte, len(padded))
	iv := bytes.Repeat([]byte{' '}, aes.BlockSize)
	mode := cipher.NewCBCEncrypter(block, iv)
	mode.CryptBlocks(ciphertext, padded)

	return ciphertext, nil
}

func TestDecryptUnencrypted(t *testing.T) {
	// Old format: no v10/v11 prefix, value is plaintext
	val := []byte("plain-cookie-value")
	got, err := decryptChromium(nil, val)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != string(val) {
		t.Errorf("got %q, want %q", got, val)
	}
}

func TestMaskCookie(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"short", "*****"},
		{"abcdefghijklmnopqrstuvwxyz", "abcdefgh**************wxyz"},
	}
	for _, tt := range tests {
		got := maskCookie(tt.input)
		if got != tt.want {
			t.Errorf("maskCookie(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestInstalledBrowsers(t *testing.T) {
	// Just ensure the function doesn't panic and returns a valid list.
	browsers := InstalledBrowsers()
	t.Logf("Installed browsers: %v", browsers)
	// Can't assert specific browsers since it depends on the test machine.
}

func TestHostFilter(t *testing.T) {
	if hostFilter("us") != "%leetcode.com%" {
		t.Error("us site should filter for leetcode.com")
	}
	if hostFilter("cn") != "%leetcode.cn%" {
		t.Error("cn site should filter for leetcode.cn")
	}
}
