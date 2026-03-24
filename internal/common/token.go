// Package common provides shared utilities.
package common

import (
	"crypto/sha256"
	"encoding/hex"
)

// HashToken creates a SHA-256 hash of a token.
// Used for secure token storage - tokens are stored by hash, not plaintext.
func HashToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return hex.EncodeToString(h[:])
}
