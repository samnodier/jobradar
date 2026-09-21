// Package password implements bounded Argon2id password storage.
package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"golang.org/x/crypto/argon2"
	"strings"
	"unicode/utf8"
)

const memory = 19 * 1024
const iterations = 2

// A small bound keeps concurrent hashing predictable on an 8 GB home server.
var slots = make(chan struct{}, 2)
var ErrBusy = errors.New("password service is busy; try again")

func Validate(value string) error {
	if !utf8.ValidString(value) || utf8.RuneCountInString(value) < 15 || len(value) > 1024 {
		return errors.New("use at least 15 characters and no more than 1024 bytes")
	}
	return nil
}
func Hash(value string) (string, error) {
	if err := Validate(value); err != nil {
		return "", err
	}
	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := derive(value, salt)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("$argon2id$v=19$m=19456,t=2,p=1$%s$%s", base64.RawStdEncoding.EncodeToString(salt), base64.RawStdEncoding.EncodeToString(key)), nil
}
func Verify(value, encoded string) (bool, error) {
	if len(value) > 1024 {
		return false, nil
	}
	parts := strings.Split(encoded, "$")
	// Accept only supported bounded parameters; corrupt hashes cannot request huge allocations.
	if len(parts) != 6 || parts[1] != "argon2id" || parts[2] != "v=19" || parts[3] != "m=19456,t=2,p=1" {
		return false, errors.New("invalid password hash")
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil || len(salt) != 16 {
		return false, errors.New("invalid salt")
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil || len(want) != 32 {
		return false, errors.New("invalid hash")
	}
	got, err := derive(value, salt)
	if err != nil {
		return false, err
	}
	return subtle.ConstantTimeCompare(got, want) == 1, nil
}
func derive(value string, salt []byte) ([]byte, error) {
	select {
	case slots <- struct{}{}:
		defer func() { <-slots }()
	default:
		return nil, ErrBusy
	}
	return argon2.IDKey([]byte(value), salt, iterations, memory, 1, 32), nil
}

// DummyHash permits equal-cost checking when the account has no password.
const DummyHash = "$argon2id$v=19$m=19456,t=2,p=1$AAAAAAAAAAAAAAAAAAAAAA$AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
