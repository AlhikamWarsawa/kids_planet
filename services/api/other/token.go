package main

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type AdminClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

func main() {
	secret := os.Getenv("JWT_SECRET")
	issuer := os.Getenv("JWT_ISSUER")

	if secret == "" || issuer == "" {
		fmt.Println("ERROR: JWT_SECRET dan JWT_ISSUER wajib di-set di environment")
		os.Exit(1)
	}

	sub := os.Getenv("JWT_SUBJECT")
	if sub == "" {
		sub = "999"
	}
	if _, err := strconv.ParseInt(sub, 10, 64); err != nil {
		fmt.Println("ERROR: JWT_SUBJECT harus angka (string int64)")
		os.Exit(1)
	}

	now := time.Now()
	exp := now.Add(15 * time.Minute)

	claims := AdminClaims{
		Role: "player",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			Issuer:    issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(secret))
	if err != nil {
		fmt.Println("ERROR:", err)
		os.Exit(1)
	}

	fmt.Println(signed)
}
