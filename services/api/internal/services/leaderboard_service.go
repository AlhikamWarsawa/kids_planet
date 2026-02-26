package services

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/ZygmaCore/kids_planet/services/api/internal/clients"
	"github.com/ZygmaCore/kids_planet/services/api/internal/models"
	"github.com/ZygmaCore/kids_planet/services/api/internal/repos"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type LeaderboardService struct {
	valkey         *clients.Valkey
	submissionRepo *repos.SubmissionRepo
}

func NewLeaderboardService(valkey *clients.Valkey, submissionRepo *repos.SubmissionRepo) *LeaderboardService {
	return &LeaderboardService{
		valkey:         valkey,
		submissionRepo: submissionRepo,
	}
}

func (s *LeaderboardService) SubmitScore(
	ctx context.Context,
	tokenGameID int64,
	tokenPlayerID string,
	guestID string,
	req models.SubmitScoreRequest,
	sessionID string,
	ipHash string,
	userAgentHash string,
) (*models.SubmitScoreResponse, *utils.AppError) {
	if req.GameID <= 0 {
		e := utils.ErrBadRequest("game_id must be a positive integer")
		return nil, &e
	}
	if req.Score < 0 {
		e := utils.ErrBadRequest("score must be >= 0")
		return nil, &e
	}

	if tokenGameID <= 0 {
		e := utils.ErrUnauthorized()
		return nil, &e
	}
	if req.GameID != tokenGameID {
		e := utils.ErrForbidden()
		return nil, &e
	}

	guestID = strings.TrimSpace(guestID)
	if guestID == "" {
		e := utils.ErrBadRequest("missing guest id")
		return nil, &e
	}

	tokenPlayerID = normalizePlayerID(tokenPlayerID)
	sessionID = strings.TrimSpace(sessionID)

	member := resolveSubmissionMember(tokenPlayerID, sessionID, guestID)
	if member == "" {
		e := utils.ErrUnauthorized()
		return nil, &e
	}
	now := time.Now().UTC()

	sub := &repos.LeaderboardSubmission{
		GameID:        req.GameID,
		PlayerID:      sql.NullInt64{Valid: false},
		SessionID:     nullString(sessionID),
		Member:        nullString(member),
		Score:         req.Score,
		IPHash:        nullString(ipHash),
		UserAgentHash: nullString(userAgentHash),
	}

	if _, err := s.submissionRepo.CreateSubmission(ctx, sub); err != nil {
		e := utils.ErrInternal()
		return nil, &e
	}

	dKey := clients.KeyGameDaily(req.GameID, now)
	wKey := clients.KeyGameWeekly(req.GameID, now)

	bestDaily, errApp := s.upsertIfHigher(ctx, dKey, member, req.Score, clients.DailyTTL)
	if errApp != nil {
		return nil, errApp
	}
	bestWeekly, errApp := s.upsertIfHigher(ctx, wKey, member, req.Score, clients.WeeklyTTL)
	if errApp != nil {
		return nil, errApp
	}

	best := bestWeekly
	if bestDaily > best {
		best = bestDaily
	}

	return &models.SubmitScoreResponse{
		Accepted:  true,
		BestScore: best,
	}, nil
}

func (s *LeaderboardService) GetTop(
	ctx context.Context,
	gameID int64,
	period string,
	scope string,
	limit int,
) ([]models.LeaderboardItem, error) {
	_, _, key, err := resolveLeaderboardKey(gameID, period, scope, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		return nil, utils.ErrBadRequest("limit must be an integer between 1 and 100")
	}

	rows, err := s.valkey.ZRevRangeWithScores(ctx, key, 0, int64(limit-1))
	if err != nil {
		return nil, utils.ErrInternal()
	}

	items := make([]models.LeaderboardItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, models.LeaderboardItem{
			Member: r.Member,
			Score:  int(r.Score),
		})
	}

	return items, nil
}

func (s *LeaderboardService) GetSelf(
	ctx context.Context,
	gameID int64,
	period string,
	scope string,
	member string,
) (*models.LeaderboardSelfDTO, error) {
	member = strings.TrimSpace(member)
	if member == "" {
		return nil, utils.ErrUnauthorized()
	}

	period, scope, key, err := resolveLeaderboardKey(gameID, period, scope, time.Now().UTC())
	if err != nil {
		return nil, err
	}

	rank, ranked, err := s.valkey.ZRevRank(ctx, key, member)
	if err != nil {
		return nil, utils.ErrInternal()
	}

	score, scoreFound, err := s.valkey.ZScore(ctx, key, member)
	if err != nil {
		return nil, utils.ErrInternal()
	}
	if !ranked || !scoreFound {
		return &models.LeaderboardSelfDTO{
			GameID: gameID,
			Rank:   nil,
			Score:  nil,
			Period: period,
			Scope:  scope,
		}, nil
	}

	rankOneBased := rank + 1
	scoreInt := int64(score)

	return &models.LeaderboardSelfDTO{
		GameID: gameID,
		Rank:   &rankOneBased,
		Score:  &scoreInt,
		Period: period,
		Scope:  scope,
	}, nil
}

func (s *LeaderboardService) upsertIfHigher(
	ctx context.Context,
	key string,
	member string,
	score int,
	ttl time.Duration,
) (int, *utils.AppError) {
	old, exists, err := s.valkey.ZScore(ctx, key, member)
	if err != nil {
		e := utils.ErrInternal()
		return 0, &e
	}

	best := score
	if exists {
		oldInt := int(old)
		if score <= oldInt {
			best = oldInt
			if err := s.valkey.Expire(ctx, key, ttl); err != nil {
				e := utils.ErrInternal()
				return 0, &e
			}
			return best, nil
		}
	}

	if err := s.valkey.ZAdd(ctx, key, member, float64(score)); err != nil {
		e := utils.ErrInternal()
		return 0, &e
	}
	if err := s.valkey.Expire(ctx, key, ttl); err != nil {
		e := utils.ErrInternal()
		return 0, &e
	}
	return best, nil
}

func resolveLeaderboardKey(
	gameID int64,
	period string,
	scope string,
	now time.Time,
) (normalizedPeriod string, normalizedScope string, key string, err error) {
	if gameID <= 0 {
		return "", "", "", utils.ErrBadRequest("game_id must be an integer >= 1")
	}

	normalizedPeriod = strings.ToLower(strings.TrimSpace(period))
	if normalizedPeriod == "" {
		normalizedPeriod = "daily"
	}
	if normalizedPeriod != "daily" && normalizedPeriod != "weekly" {
		return "", "", "", utils.ErrBadRequest("period must be 'daily' or 'weekly'")
	}

	normalizedScope = strings.ToLower(strings.TrimSpace(scope))
	if normalizedScope == "" {
		normalizedScope = "game"
	}
	if normalizedScope != "game" && normalizedScope != "global" {
		return "", "", "", utils.ErrBadRequest("scope must be 'game' or 'global'")
	}

	now = now.UTC()

	if normalizedScope == "game" {
		if normalizedPeriod == "daily" {
			return normalizedPeriod, normalizedScope, clients.KeyGameDaily(gameID, now), nil
		}
		return normalizedPeriod, normalizedScope, clients.KeyGameWeekly(gameID, now), nil
	}

	if normalizedPeriod == "daily" {
		return normalizedPeriod, normalizedScope, clients.KeyGlobalDaily(now), nil
	}
	return normalizedPeriod, normalizedScope, clients.KeyGlobalWeekly(now), nil
}

func resolveSubmissionMember(playerID string, sessionID string, guestID string) string {
	playerID = normalizePlayerID(playerID)
	if playerID != "" {
		return "p:" + playerID
	}

	sessionID = strings.TrimSpace(sessionID)
	if sessionID != "" {
		return "s:" + sessionID
	}

	guestID = strings.TrimSpace(guestID)
	if guestID != "" {
		return "g:" + guestID
	}

	return ""
}

func normalizePlayerID(playerID string) string {
	playerID = strings.TrimSpace(playerID)
	if playerID == "" {
		return ""
	}
	if _, err := uuid.Parse(playerID); err != nil {
		return ""
	}
	return playerID
}

func nullString(v string) sql.NullString {
	v = strings.TrimSpace(v)
	if v == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: v, Valid: true}
}
