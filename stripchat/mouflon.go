package stripchat

import (
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

// reToken matches _NUMBER_TOKEN_NUMBER patterns in MOUFLON v2 segment URIs.
// The token (group 2) is the encrypted portion sandwiched between two numeric fields.
var reToken = regexp.MustCompile(`_(\d+)_([^_]+)_(\d+)`)

// ParsePKeyFromMaster extracts the pkey from a master playlist's
// #EXT-X-MOUFLON:PSCH:v2:{pkey} line. Returns empty string if not found.
func ParsePKeyFromMaster(masterBody string) string {
	for _, line := range strings.Split(masterBody, "\n") {
		line = strings.TrimRight(line, "\r\n ")
		if strings.HasPrefix(line, "#EXT-X-MOUFLON:PSCH:") {
			parts := strings.SplitN(line, ":", 4)
			if len(parts) == 4 {
				return parts[3]
			}
		}
	}
	return ""
}

// DecryptMouflonURI decrypts the encrypted token in a MOUFLON v2 segment URI.
// Algorithm: reverse token → base64-decode → XOR with cyclic SHA256(pdkey).
func DecryptMouflonURI(uri, pdkey string) (string, error) {
	m := reToken.FindStringSubmatch(uri)
	if m == nil {
		return uri, nil
	}

	result, err := decryptToken(uri, pdkey)
	if err != nil {
		return "", err
	}

	if !isPrintableASCII(result) {
		return "", fmt.Errorf("decryption produced non-printable bytes (hex=%x); pdkey is likely wrong", result)
	}

	encryptedPart := m[2]
	decryptedPart := string(result)
	decryptedURI := strings.Replace(uri, encryptedPart, decryptedPart, 1)
	return decryptedURI, nil
}

// decryptToken extracts, reverses, base64-decodes, and XOR-decrypts the token.
func decryptToken(uri, pdkey string) ([]byte, error) {
	m := reToken.FindStringSubmatch(uri)
	if m == nil {
		return nil, fmt.Errorf("no encrypted token found in URI")
	}
	encryptedPart := m[2]

	runes := []rune(encryptedPart)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	reversed := string(runes)

	decoded, err := base64.StdEncoding.DecodeString(padBase64(reversed))
	if err != nil {
		decoded, err = base64.URLEncoding.DecodeString(padBase64(reversed))
		if err != nil {
			return nil, fmt.Errorf("base64 decode: %w", err)
		}
	}

	hash := sha256.Sum256([]byte(pdkey))
	result := make([]byte, len(decoded))
	for i, b := range decoded {
		result[i] = b ^ hash[i%32]
	}
	return result, nil
}

func isPrintableASCII(b []byte) bool {
	if len(b) == 0 {
		return false
	}
	for _, c := range b {
		if c < 0x20 || c > 0x7E {
			return false
		}
	}
	return true
}

func padBase64(s string) string {
	switch len(s) % 4 {
	case 2:
		return s + "=="
	case 3:
		return s + "="
	default:
		return s
	}
}
