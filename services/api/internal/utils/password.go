package utils

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidPassword = errors.New("invalid password")
)

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
