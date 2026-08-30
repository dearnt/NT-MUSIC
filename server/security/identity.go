package security

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
)

var ErrInvalidIdentity = errors.New("invalid identity")

const sessionIDBytes = 16

func NewSessionID() (string, error) {
	data := make([]byte, sessionIDBytes)

	if _, err := rand.Read(data); err != nil {
		return "", err
	}

	return hex.EncodeToString(data), nil
}

func ValidateSessionID(id string) error {
	if len(id) != sessionIDBytes*2 {
		return ErrInvalidIdentity
	}

	for _, char := range id {
		if !((char >= 'a' && char <= 'f') || (char >= '0' && char <= '9')) {
			return ErrInvalidIdentity
		}
	}

	return nil
}
