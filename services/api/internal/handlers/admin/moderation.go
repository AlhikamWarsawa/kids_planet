package admin

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/ZygmaCore/kids_planet/services/api/internal/middleware"
	"github.com/ZygmaCore/kids_planet/services/api/internal/repos"
	"github.com/ZygmaCore/kids_planet/services/api/internal/services"
	"github.com/ZygmaCore/kids_planet/services/api/internal/utils"
)

type ModerationHandler struct {
	submissionRepo *repos.SubmissionRepo
	leaderboardSvc *services.LeaderboardService
}

func NewModerationHandler(submissionRepo *repos.SubmissionRepo, leaderboardSvc *services.LeaderboardService) *ModerationHandler {
	return &ModerationHandler{
		submissionRepo: submissionRepo,
		leaderboardSvc: leaderboardSvc,
	}
}

type flaggedSubmissionDTO struct {
	ID         int64   `json:"id"`
	GameID     int64   `json:"game_id"`
	PlayerName string  `json:"player_name"`
	Score      int     `json:"score"`
	Flagged    bool    `json:"flagged"`
	FlagReason *string `json:"flag_reason,omitempty"`
	SessionID  *string `json:"session_id,omitempty"`
	CreatedAt  string  `json:"created_at"`
}

func (h *ModerationHandler) ListFlagged(c *fiber.Ctx) error {
	limit := 50
	if v := strings.TrimSpace(c.Query("limit")); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return utils.Fail(c, utils.ErrBadRequest("limit must be a positive integer"))
		}
		limit = n
	}

	ctx := c.UserContext()
	if ctx == nil {
		ctx = context.Background()
	}

	items, err := h.submissionRepo.ListFlagged(ctx, limit)
	if err != nil {
		return utils.Fail(c, utils.ErrInternal())
	}

	out := make([]flaggedSubmissionDTO, 0, len(items))
	for _, it := range items {
		var flagReason *string
		if it.FlagReason.Valid {
			v := it.FlagReason.String
			flagReason = &v
		}
		var sessionID *string
		if it.SessionID.Valid {
			v := it.SessionID.String
			sessionID = &v
		}
		out = append(out, flaggedSubmissionDTO{
			ID:         it.ID,
			GameID:     it.GameID,
			PlayerName: strings.TrimSpace(it.PlayerName),
			Score:      it.Score,
			Flagged:    it.Flagged,
			FlagReason: flagReason,
			SessionID:  sessionID,
			CreatedAt:  it.CreatedAt.UTC().Format(time.RFC3339),
		})
	}

	return utils.Success(c, fiber.Map{
		"items": out,
	})
}

func parseSubmissionID(body []byte) (int64, error) {
	if len(body) == 0 {
		return 0, errors.New("submission_id is required")
	}

	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
		return 0, errors.New("invalid json body")
	}
	rawID, ok := payload["submission_id"]
	if !ok {
		return 0, errors.New("submission_id is required")
	}

	switch v := rawID.(type) {
	case float64:
		if v <= 0 {
			return 0, errors.New("submission_id must be a positive integer")
		}
		return int64(v), nil
	case string:
		id, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
		if err != nil || id <= 0 {
			return 0, errors.New("submission_id must be a positive integer")
		}
		return id, nil
	default:
		return 0, errors.New("submission_id must be a positive integer")
	}
}

func (h *ModerationHandler) RemoveScore(c *fiber.Ctx) error {
	uidAny := c.Locals(middleware.LocalUserID)
	adminID, ok := uidAny.(int64)
	if !ok || adminID <= 0 {
		return utils.Fail(c, utils.ErrUnauthorized())
	}

	body := c.Body()
	submissionID, err := parseSubmissionID(body)
	if err != nil {
		return utils.Fail(c, utils.ErrBadRequest(err.Error()))
	}

	ctx := c.UserContext()
	if ctx == nil {
		ctx = context.Background()
	}

	sub, err := h.submissionRepo.GetByID(ctx, submissionID)
	if err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return utils.Fail(c, utils.ErrNotFound("submission not found"))
		}
		return utils.Fail(c, utils.ErrInternal())
	}

	now := time.Now().UTC()
	if err := h.submissionRepo.MarkRemovedByAdmin(ctx, submissionID, adminID, now); err != nil {
		if errors.Is(err, repos.ErrNotFound) {
			return utils.Fail(c, utils.ErrNotFound("submission not found"))
		}
		return utils.Fail(c, utils.ErrInternal())
	}

	if err := h.leaderboardSvc.RemoveSubmissionFromLeaderboards(ctx, sub); err != nil {
		return utils.Fail(c, utils.ErrInternal())
	}

	return utils.Success(c, fiber.Map{"ok": true})
}
