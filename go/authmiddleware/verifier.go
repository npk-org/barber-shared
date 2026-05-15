package authmiddleware

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v2/jwk"
)

type Verifier struct {
	jwksURL  string
	issuer   string
	audience string

	mu       sync.RWMutex
	keys     jwk.Set
	fetched  time.Time
	cacheTTL time.Duration
}

func NewVerifier(jwksURL, issuer, audience string) *Verifier {
	return &Verifier{
		jwksURL:  jwksURL,
		issuer:   issuer,
		audience: audience,
		cacheTTL: 10 * time.Minute,
	}
}

func (v *Verifier) refresh(ctx context.Context) error {
	set, err := jwk.Fetch(ctx, v.jwksURL)
	if err != nil {
		return fmt.Errorf("fetch jwks: %w", err)
	}
	v.mu.Lock()
	v.keys = set
	v.fetched = time.Now()
	v.mu.Unlock()
	return nil
}

func (v *Verifier) RefreshNow(ctx context.Context) error {
	return v.refresh(ctx)
}

func (v *Verifier) Verify(ctx context.Context, tokenStr string) (User, error) {
	v.mu.RLock()
	stale := v.keys == nil || time.Since(v.fetched) > v.cacheTTL
	v.mu.RUnlock()
	if stale {
		if err := v.refresh(ctx); err != nil {
			return User{}, err
		}
	}

	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		kid, _ := t.Header["kid"].(string)
		v.mu.RLock()
		defer v.mu.RUnlock()
		key, ok := v.keys.LookupKeyID(kid)
		if !ok {
			return nil, errors.New("unknown kid")
		}
		var raw interface{}
		if err := key.Raw(&raw); err != nil {
			return nil, err
		}
		return raw, nil
	}, jwt.WithIssuer(v.issuer), jwt.WithAudience(v.audience))
	if err != nil {
		return User{}, err
	}
	if !tok.Valid {
		return User{}, errors.New("invalid token")
	}

	claims, ok := tok.Claims.(jwt.MapClaims)
	if !ok {
		return User{}, errors.New("unexpected claims type")
	}

	sub, _ := claims["sub"].(string)
	role, _ := claims["role"].(string)
	shopID, _ := claims["shop_id"].(string)
	jti, _ := claims["jti"].(string)
	if sub == "" || role == "" {
		return User{}, errors.New("missing sub or role")
	}

	return User{
		ID:     sub,
		Role:   Role(role),
		ShopID: shopID,
		JTI:    jti,
	}, nil
}
