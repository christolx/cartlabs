package identity

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	client *redis.Client
}

func NewRedisRateLimiter(client *redis.Client) *RedisRateLimiter {
	return &RedisRateLimiter{client: client}
}

var incrementWindow = redis.NewScript(`
local count = redis.call('INCR', KEYS[1])
if count == 1 then
  redis.call('EXPIRE', KEYS[1], ARGV[1])
end
return count
`)

func (l *RedisRateLimiter) Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error) {
	if limit < 1 || window <= 0 {
		return false, fmt.Errorf("invalid rate limit configuration")
	}
	count, err := incrementWindow.Run(ctx, l.client, []string{"cartlabs:auth:" + key}, int(window.Seconds())).Int64()
	if err != nil {
		return false, fmt.Errorf("apply authentication rate limit: %w", err)
	}
	return count <= int64(limit), nil
}

func RateLimitKey(remoteAddress, subject string) string {
	normalized := strings.ToLower(strings.TrimSpace(remoteAddress)) + "\x00" + strings.ToLower(strings.TrimSpace(subject))
	sum := sha256.Sum256([]byte(normalized))
	return hex.EncodeToString(sum[:16])
}
