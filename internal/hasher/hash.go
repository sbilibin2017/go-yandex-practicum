package hasher

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// Hash вычисляет HMAC-SHA256 хеш для переданных данных с использованием указанного ключа.
// data — данные для хеширования.
// key — секретный ключ для HMAC.
// Возвращает строковое представление хеша в шестнадцатеричном формате.
func Hash(data []byte, key string) string {
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// Compare сравнивает две хеш-строки в безопасном режиме, чтобы предотвратить атаки по времени.
// hash1, hash2 — строки хешей для сравнения.
// Возвращает true, если хеши совпадают, иначе false.
func Compare(hash1, hash2 string) bool {
	return hmac.Equal([]byte(hash1), []byte(hash2))
}
