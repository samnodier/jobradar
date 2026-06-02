// Package crypto
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
)

type Service struct {
	gcm cipher.AEAD
}

func New(base64Key string) (*Service, error) {
	if base64Key == "" {
		return nil, errors.New("encryption key environment variable is empty")
	}

	// Decode base64 string into raw bytes
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 key: %w", err)
	}

	// Validate the lenght - strictly 32
	if len(key) != 32 {
		return nil, fmt.Errorf("invalid encryption key length: got %d, want 32 bytes", len(key))
	}

	// create AES cipher block
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("failed to cipher the block: %w", err)
	}

	// wrap the block in galois/counter mode (GCM)
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("failed to create GCM: %w", err)
	}

	return &Service{gcm: gcm}, nil
}

// Encrypt locks the plaintext and return a base64-encoded string for the DB
func (s *Service) Encrypt(plainText string) (string, error) {
	// A "nonce" is a Number Used Once. GCM absolutely requires a unique nonce per encryption
	nonce := make([]byte, s.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Seal encrypts the data
	// Append the encrypted data to the nonce byte slice
	// We need the nonce later to decrypt the data
	// The data doesn't need to be secret
	ciphertext := s.gcm.Seal(nonce, nonce, []byte(plainText), nil)

	// Return as base64 so it fits cleanly into the database
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

func (s *Service) Decrypt(dbString string) (string, error) {
	// Decode the base64 string back to raw bytes
	data, err := base64.StdEncoding.DecodeString(dbString)
	if err != nil {
		return "", fmt.Errorf("failed to base64 decode: %w", err)
	}

	nonceSize := s.gcm.NonceSize()
	if len(data) < nonceSize {
		return "", errors.New("ciphertext too short")
	}

	// Split the bytes into the nonce and the encrypted message
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	// Open (decrypt) the message
	plaintext, err := s.gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", fmt.Errorf("failed to decrypt: %w", err)
	}
	return string(plaintext), nil
}
