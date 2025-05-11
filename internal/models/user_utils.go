package models

import (
	"crypto/sha256"
	"encoding/hex"
)

func HashPassword(password string) string {
	h := sha256.New()
	h.Write([]byte(password))
	return hex.EncodeToString(h.Sum(nil))
}

func CheckPassword(hash, password string) bool {
	return hash == HashPassword(password)
}
