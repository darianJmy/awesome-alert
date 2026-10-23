package utils

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"guanlan/internal/config"
	"io"
	"strings"
)

func EncryptSecret(plain string) (string, error) {
	key := sha256.Sum256([]byte(config.AppConfig.JWTSecret))
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nil, nonce, []byte(strings.TrimSpace(plain)), nil)
	return base64.StdEncoding.EncodeToString(append(nonce, ciphertext...)), nil
}

func DecryptSecret(ciphertext string) (string, error) {
	key := sha256.Sum256([]byte(config.AppConfig.JWTSecret))
	raw, err := base64.StdEncoding.DecodeString(ciphertext)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(key[:])
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(raw) < gcm.NonceSize() {
		return "", errors.New("密文格式错误")
	}
	nonce := raw[:gcm.NonceSize()]
	plain, err := gcm.Open(nil, nonce, raw[gcm.NonceSize():], nil)
	if err != nil {
		return "", err
	}
	return string(plain), nil
}
