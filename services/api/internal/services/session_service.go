package services

import (
	"context"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/ZygmaCore/kids_planet/services/api/internal/config"
	"github.com/ZygmaCore/kids_planet/services/api/internal/models"
	"github.com/ZygmaCore/kids_planet/services/api/internal/repos"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type SessionService struct {
	cfg      config.Config
	gameRepo *repos.GameRepo
	ttl      time.Duration
}

func NewSessionService(cfg config.Config, gameRepo *repos.GameRepo) *SessionService {
	return &SessionService{
		cfg:      cfg,
		gameRepo: gameRepo,
		ttl:      2 * time.Hour,
	}
}

type PlayTokenClaims struct {
	GameID int64  `json:"game_id"`
	Typ    string `json:"typ"`
	jwt.RegisteredClaims
}

func (s *SessionService) StartSession(ctx context.Context, gameID int64, sub string) (*models.StartSessionResponse, *utils.AppError) {
	if gameID <= 0 {
		e := utils.ErrBadRequest("game_id must be a positive integer")
		return nil, &e
	}

	g, err := s.gameRepo.GetByID(ctx, gameID)
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			e := utils.ErrBadRequest("game not found")
			return nil, &e
		}
		e := utils.ErrInternal()
		return nil, &e
	}

	if g.Status != "active" {
		e := utils.ErrBadRequest("game is not active")
		return nil, &e
	}

	now := time.Now().UTC()
	exp := now.Add(s.ttl)

	claims := PlayTokenClaims{
		GameID: gameID,
		Typ:    "play",
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.cfg.JWT.Issuer,
			Subject:   sub,
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}

	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, err := t.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		e := utils.ErrInternal()
		return nil, &e
	}

	return &models.StartSessionResponse{
		PlayToken: tokenStr,
		ExpiresAt: exp,
	}, nil
}
