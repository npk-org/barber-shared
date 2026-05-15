package authmiddleware

import (
	"context"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type RevocationChecker struct {
	rdb       *redis.Client
	keyPrefix string

	mu    sync.RWMutex
	cache map[string]time.Time
	ttl   time.Duration
}

func NewRevocationChecker(rdb *redis.Client) *RevocationChecker {
	return &RevocationChecker{
		rdb:       rdb,
		keyPrefix: "revoked:",
		cache:     make(map[string]time.Time),
		ttl:       60 * time.Second,
	}
}

func (r *RevocationChecker) IsRevoked(ctx context.Context, jti string) (bool, error) {
	if jti == "" {
		return false, nil
	}

	r.mu.RLock()
	if exp, ok := r.cache[jti]; ok && time.Now().Before(exp) {
		r.mu.RUnlock()
		return true, nil
	}
	r.mu.RUnlock()

	n, err := r.rdb.Exists(ctx, r.keyPrefix+jti).Result()
	if err != nil {
		return false, err
	}
	revoked := n > 0
	if revoked {
		r.mu.Lock()
		r.cache[jti] = time.Now().Add(r.ttl)
		r.mu.Unlock()
	}
	return revoked, nil
}
