package access

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

const iterations = 600000

func Random() string { return hex.EncodeToString(randBytes(32)) }
func randBytes(n int) []byte {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}
	return b
}
func Digest(s string) string { h := sha256.Sum256([]byte(s)); return hex.EncodeToString(h[:]) }
func Password(password string) (string, error) {
	if len(password) < 14 || len(password) > 256 {
		return "", errors.New("use a password of 14–256 bytes")
	}
	if strings.Contains(strings.ToLower(password), "password") || strings.Count(password, string(password[0])) == len(password) {
		return "", errors.New("choose a less predictable password")
	}
	salt := randBytes(16)
	key, err := pbkdf2.Key(sha256.New, password, salt, iterations, 32)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("pbkdf2-sha256$%d$%x$%x", iterations, salt, key), nil
}
func CheckPassword(encoded, password string) bool {
	var count int
	var saltHex, keyHex string
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != "pbkdf2-sha256" {
		return false
	}
	if _, err := fmt.Sscanf(parts[1], "%d", &count); err != nil || count < iterations || count > 2000000 {
		return false
	}
	saltHex, keyHex = parts[2], parts[3]
	salt, e1 := hex.DecodeString(saltHex)
	want, e2 := hex.DecodeString(keyHex)
	if e1 != nil || e2 != nil || len(want) != 32 {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, password, salt, count, 32)
	return err == nil && subtle.ConstantTimeCompare(got, want) == 1
}
func Seal(key []byte, secret string) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	g, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	nonce := randBytes(g.NonceSize())
	return g.Seal(nonce, nonce, []byte(secret), []byte("totp-v1")), nil
}
func Open(key, data []byte) (string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", err
	}
	g, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(data) < g.NonceSize() {
		return "", errors.New("invalid secret")
	}
	p, err := g.Open(nil, data[:g.NonceSize()], data[g.NonceSize():], []byte("totp-v1"))
	return string(p), err
}
func Equal(a, b string) bool { return hmac.Equal([]byte(a), []byte(b)) }
