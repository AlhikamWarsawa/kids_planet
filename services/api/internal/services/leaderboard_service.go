package services

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/ZygmaCore/kids_planet/services/api/internal/clients"
	"github.com/ZygmaCore/kids_planet/services/api/internal/models"
	"github.com/ZygmaCore/kids_planet/services/api/internal/repos"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type LeaderboardService struct {
	valkey         *clients.Valkey
	submissionRepo *repos.SubmissionRepo
}

const (
	maxLeaderboardScore   = 1_000_000
	suspiciousBurstLimit  = 8
	suspiciousBurstWindow = 10 * time.Second
)

func NewLeaderboardService(valkey *clients.Valkey, submissionRepo *repos.SubmissionRepo) *LeaderboardService {
	return &LeaderboardService{
		valkey:         valkey,
		submissionRepo: submissionRepo,
	}
}

func (s *LeaderboardService) SubmitScore(
	ctx context.Context,
	tokenGameID int64,
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

	sessionID = strings.TrimSpace(sessionID)

	flagReasons := make([]string, 0, 2)
	if req.Score > maxLeaderboardScore {
		flagReasons = append(flagReasons, "score_out_of_bounds")
	}
	if sessionID == "" {
		flagReasons = append(flagReasons, "missing_session_id")
	}

	if sessionID != "" && s.valkey != nil {
		key := "ac:leaderboard:submit:burst:" + sessionID
		count, err := s.valkey.IncrWithTTL(ctx, key, suspiciousBurstWindow)
		if err == nil && count > int64(suspiciousBurstLimit) {
			flagReasons = append(flagReasons, "rate_suspicious")
		}
	}

	flagged := len(flagReasons) > 0
	flagReason := strings.Join(flagReasons, ",")

	member := "g:" + guestID
	now := time.Now().UTC()

	sub := &repos.LeaderboardSubmission{
		GameID:        req.GameID,
		PlayerID:      sql.NullInt64{Valid: false},
		SessionID:     nullString(sessionID),
		Score:         req.Score,
		IPHash:        nullString(ipHash),
		UserAgentHash: nullString(userAgentHash),
		Flagged:       flagged,
		FlagReason:    nullString(flagReason),
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
	if gameID <= 0 {
		return nil, utils.ErrBadRequest("game_id must be an integer >= 1")
	}

	period = strings.ToLower(strings.TrimSpace(period))
	if period == "" {
		period = "daily"
	}
	if period != "daily" && period != "weekly" {
		return nil, utils.ErrBadRequest("period must be 'daily' or 'weekly'")
	}

	scope = strings.ToLower(strings.TrimSpace(scope))
	if scope == "" {
		scope = "game"
	}
	if scope != "game" && scope != "global" {
		return nil, utils.ErrBadRequest("scope must be 'game' or 'global'")
	}

	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		return nil, utils.ErrBadRequest("limit must be an integer between 1 and 100")
	}

	now := time.Now().UTC()

	var key string
	if scope == "game" {
		if period == "daily" {
			key = clients.KeyGameDaily(gameID, now)
		} else {
			key = clients.KeyGameWeekly(gameID, now)
		}
	} else {
		if period == "daily" {
			key = clients.KeyGlobalDaily(now)
		} else {
			key = clients.KeyGlobalWeekly(now)
		}
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

func nullString(v string) sql.NullString {
	v = strings.TrimSpace(v)
	if v == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: v, Valid: true}
}
