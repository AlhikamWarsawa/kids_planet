package clients

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/ZygmaCore/kids_planet/services/api/internal/config"
	"github.com/redis/go-redis/v9"
)

type Valkey struct {
	rdb *redis.Client
}

type ZMemberScore struct {
	Member string
	Score  float64
}

const (
	DailyTTL  = 8 * 24 * time.Hour
	WeeklyTTL = 6 * 7 * 24 * time.Hour
)

var incrWithTTLScript = redis.NewScript(`
local current = redis.call("INCR", KEYS[1])
if current == 1 then
  redis.call("PEXPIRE", KEYS[1], ARGV[1])
end
return current
`)

func NewValkey(cfg config.ValkeyConfig) (*Valkey, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		_ = rdb.Close()
		return nil, err
	}

	return &Valkey{rdb: rdb}, nil
}

func (v *Valkey) Close() error {
	return v.rdb.Close()
}

func (v *Valkey) ZScore(ctx context.Context, key, member string) (float64, bool, error) {
	score, err := v.rdb.ZScore(ctx, key, member).Result()
	if err == redis.Nil {
		return 0, false, nil
	}
	if err != nil {
		return 0, false, err
	}
	return score, true, nil
}

func (v *Valkey) ZAdd(ctx context.Context, key, member string, score float64) error {
	return v.rdb.ZAdd(ctx, key, redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

func (v *Valkey) ZRevRangeWithScores(ctx context.Context, key string, start, stop int64) ([]ZMemberScore, error) {
	res, err := v.rdb.ZRevRangeWithScores(ctx, key, start, stop).Result()
	if err != nil {
		return nil, err
	}

	out := make([]ZMemberScore, 0, len(res))
	for _, z := range res {
		member, ok := z.Member.(string)
		if !ok {
			member = fmt.Sprint(z.Member)
		}
		out = append(out, ZMemberScore{
			Member: member,
			Score:  z.Score,
		})
	}
	return out, nil
}

func (v *Valkey) Expire(ctx context.Context, key string, ttl time.Duration) error {
	return v.rdb.Expire(ctx, key, ttl).Err()
}

func (v *Valkey) ZRem(ctx context.Context, key string, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	return v.rdb.ZRem(ctx, key, members).Err()
}

func (v *Valkey) ZRemPipeline(ctx context.Context, keys []string, member string) error {
	if v == nil {
		return errors.New("valkey client is nil")
	}
	member = strings.TrimSpace(member)
	if member == "" || len(keys) == 0 {
		return nil
	}

	pipe := v.rdb.Pipeline()
	for _, key := range keys {
		key = strings.TrimSpace(key)
		if key == "" {
			continue
		}
		pipe.ZRem(ctx, key, member)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (v *Valkey) IncrWithTTL(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	if ttl <= 0 {
		return v.rdb.Incr(ctx, key).Result()
	}

	ms := ttl.Milliseconds()
	if ms <= 0 {
		ms = 1
	}

	res, err := incrWithTTLScript.Run(ctx, v.rdb, []string{key}, ms).Result()
	if err != nil {
		return 0, err
	}

	switch value := res.(type) {
	case int64:
		return value, nil
	case float64:
		return int64(value), nil
	case string:
		parsed, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("valkey.incr_with_ttl: parse result: %w", err)
		}
		return parsed, nil
	default:
		return 0, fmt.Errorf("valkey.incr_with_ttl: unexpected result type %T", res)
	}
}

func KeyGameDaily(gameID int64, t time.Time) string {
	return fmt.Sprintf("lb:game:%d:d:%s", gameID, t.UTC().Format("20060102"))
}

func KeyGameWeekly(gameID int64, t time.Time) string {
	year, week := t.UTC().ISOWeek()
	return fmt.Sprintf("lb:game:%d:w:%04d%02d", gameID, year, week)
}

func KeyGlobalDaily(t time.Time) string {
	return fmt.Sprintf("lb:global:d:%s", t.UTC().Format("20060102"))
}

func KeyGlobalWeekly(t time.Time) string {
	year, week := t.UTC().ISOWeek()
	return fmt.Sprintf("lb:global:w:%04d%02d", year, week)
}
