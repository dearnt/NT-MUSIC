package security

import (
	"errors"
	"strings"
	"unicode"
)

var (
	ErrInvalidInput    = errors.New("invalid input")
	ErrInvalidRoomCode = errors.New("invalid room code")
)

func CleanText(value string, maxLength int) (string, error) {
	value = strings.TrimSpace(value)

	if value == "" || len(value) > maxLength {
		return "", ErrInvalidInput
	}

	for _, r := range value {
		if unicode.IsControl(r) {
			return "", ErrInvalidInput
		}
	}

	return value, nil
}

func ValidateRoomCode(code string) error {
	code = strings.TrimSpace(code)

	if code == "" || len(code) > MaxRoomCodeLength {
		return ErrInvalidRoomCode
	}

	for _, r := range code {
		if !((r >= 'A' && r <= 'Z') ||
			(r >= 'a' && r <= 'z') ||
			(r >= '0' && r <= '9')) {
			return ErrInvalidRoomCode
		}
	}

	return nil
}

func ValidateName(name string) (string, error) {
	return CleanText(name, MaxNameLength)
}

func ValidateSongTitle(title string) (string, error) {
	return CleanText(title, MaxSongTitleLength)
}

func ValidateSongURL(value string) (string, error) {
	value = strings.TrimSpace(value)

	if value == "" || len(value) > MaxSongURLLength {
		return "", ErrInvalidInput
	}

	for _, r := range value {
		if unicode.IsControl(r) {
			return "", ErrInvalidInput
		}
	}

	return value, nil
}
