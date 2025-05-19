package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func Hash(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

func Compare(hash1, hash2 string) bool {
	return hmac.Equal([]byte(hash1), []byte(hash2))
}
