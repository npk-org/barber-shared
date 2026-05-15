package authmiddleware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"
)

type Config struct {
	Verifier   *Verifier
	Revocation *RevocationChecker
}

func New(cfg Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tok, err := extractBearer(r)
			if err != nil {
				writeUnauthorized(w, "missing_or_malformed_token")
				return
			}

			user, err := cfg.Verifier.Verify(r.Context(), tok)
			if err != nil {
				writeUnauthorized(w, "invalid_token")
				return
			}

			if cfg.Revocation != nil {
				revoked, err := cfg.Revocation.IsRevoked(r.Context(), user.JTI)
				if err == nil && revoked {
					writeUnauthorized(w, "token_revoked")
					return
				}
			}

			ctx := WithUser(r.Context(), user)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func RequireRole(roles ...Role) func(http.Handler) http.Handler {
	allowed := make(map[Role]struct{}, len(roles))
	for _, r := range roles {
		allowed[r] = struct{}{}
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			u, ok := FromContext(r.Context())
			if !ok {
				writeUnauthorized(w, "no_user_in_context")
				return
			}
			if _, ok := allowed[u.Role]; !ok {
				writeForbidden(w, "wrong_role")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func extractBearer(r *http.Request) (string, error) {
	h := r.Header.Get("Authorization")
	if h == "" {
		return "", errors.New("missing authorization header")
	}
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return "", errors.New("not a bearer token")
	}
	return strings.TrimSpace(h[len(prefix):]), nil
}

type errBody struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func writeUnauthorized(w http.ResponseWriter, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	body := errBody{}
	body.Error.Code = code
	body.Error.Message = "unauthorized"
	_ = json.NewEncoder(w).Encode(body)
}

func writeForbidden(w http.ResponseWriter, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	body := errBody{}
	body.Error.Code = code
	body.Error.Message = "forbidden"
	_ = json.NewEncoder(w).Encode(body)
}

// Sanity ensures we use context import on Go versions where some toolchains warn.
var _ = context.Background
