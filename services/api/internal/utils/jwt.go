package utils

import (
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ZygmaCore/kids_planet/services/api/internal/config"
	"github.com/ZygmaCore/kids_planet/services/api/internal/repos"
)

type AdminClaims struct {
	Role string `json:"role"`
	jwt.RegisteredClaims
}

type ParsedToken struct {
	UserID int64
	Role   string
	Issuer string
	Iat    time.Time
	Exp    time.Time
}

func GenerateAdminToken(jwtCfg config.JWTConfig, user *repos.User) (tokenString string, expiresInSec int64, err error) {
	if user == nil || user.ID <= 0 {
		return "", 0, errors.New("invalid user")
	}
	if user.Role != "admin" {
		return "", 0, errors.New("user is not admin")
	}

	now := time.Now()
	exp := now.Add(jwtCfg.ExpiresIn)

	claims := AdminClaims{
		Role: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(user.ID, 10),
			Issuer:    jwtCfg.Issuer,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	signed, err := tok.SignedString([]byte(jwtCfg.Secret))
	if err != nil {
		return "", 0, err
	}

	return signed, int64(jwtCfg.ExpiresIn.Seconds()), nil
}

func ParseToken(jwtCfg config.JWTConfig, tokenString string) (ParsedToken, error) {
	claims := &AdminClaims{}

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuer(jwtCfg.Issuer),
	)

	_, err := parser.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (any, error) {
		return []byte(jwtCfg.Secret), nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return ParsedToken{}, ErrUnauthorized()
		}
		if errors.Is(err, jwt.ErrTokenSignatureInvalid) ||
			errors.Is(err, jwt.ErrTokenMalformed) ||
			errors.Is(err, jwt.ErrTokenInvalidIssuer) ||
			errors.Is(err, jwt.ErrTokenUnverifiable) {
			return ParsedToken{}, ErrUnauthorized()
		}
		return ParsedToken{}, ErrUnauthorized()
	}

	uid, convErr := strconv.ParseInt(claims.Subject, 10, 64)
	if convErr != nil || uid <= 0 {
		return ParsedToken{}, ErrUnauthorized()
	}

	if claims.Role == "" {
		return ParsedToken{}, ErrUnauthorized()
	}

	var iat, exp time.Time
	if claims.IssuedAt != nil {
		iat = claims.IssuedAt.Time
	}
	if claims.ExpiresAt != nil {
		exp = claims.ExpiresAt.Time
	}

	return ParsedToken{
		UserID: uid,
		Role:   claims.Role,
		Issuer: claims.Issuer,
		Iat:    iat,
		Exp:    exp,
	}, nil
}
