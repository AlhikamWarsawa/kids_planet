package utils

import (
	"errors"
	"regexp"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword = errors.New("invalid password")
	ErrInvalidPIN      = errors.New("invalid pin")
)

var pinRe = regexp.MustCompile(`^\d{6}$`)

func HashPassword(plain string) (string, error) {
	if len(plain) <= 8 {
		return "", errors.New("password min 8 characters")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func ComparePassword(hash, plain string) error {
	if len(hash) == 0 || len(plain) == 0 {
		return ErrInvalidPassword
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)); err != nil {
		return ErrInvalidPassword
	}
	return nil
}

func ValidatePIN(pin string) error {
	pin = strings.TrimSpace(pin)
	if !pinRe.MatchString(pin) {
		return ErrInvalidPIN
	}
	return nil
}

func HashPIN(pin string) (string, error) {
	if err := ValidatePIN(pin); err != nil {
		return "", err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(strings.TrimSpace(pin)), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hash), nil
}

func ComparePIN(hash, pin string) error {
	hash = strings.TrimSpace(hash)
	if hash == "" {
		return ErrInvalidPIN
	}

	if err := ValidatePIN(pin); err != nil {
		return ErrInvalidPIN
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(strings.TrimSpace(pin))); err != nil {
		return ErrInvalidPIN
	}
	return nil
}
