package api

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const (
	// ContextPrincipal holds parsed token claims in request context.
	ContextPrincipal contextKey = "principal"
)

// GenerateToken creates a signed JWT with the given subject and role and TTL.
func GenerateToken(subject string, role string, ttl time.Duration) (string, error) {
	secret := os.Getenv("AUTH_JWT_SECRET")
	if secret == "" {
		secret = "dev-secret" // fallback for local dev only
	}
	now := time.Now()
	claims := jwt.MapClaims{
		"sub":  subject,
		"role": role,
		"iat":  now.Unix(),
		"exp":  now.Add(ttl).Unix(),
	}
	t := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return t.SignedString([]byte(secret))
}

// parseAndValidate parses token string and returns claims if valid.
// Returns a generic map to avoid exposing jwt types in public signatures.
func parseAndValidate(tokenStr string) (map[string]interface{}, error) {
	secret := os.Getenv("AUTH_JWT_SECRET")
	if secret == "" {
		secret = "dev-secret"
	}
	tok, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		// only allow HMAC
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenUnverifiable
		}
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := tok.Claims.(jwt.MapClaims); ok && tok.Valid {
		// convert jwt.MapClaims to map[string]interface{}
		out := make(map[string]interface{}, len(claims))
		for k, v := range claims {
			out[k] = v
		}
		return out, nil
	}
	return nil, jwt.ErrTokenInvalidClaims
}

// RequireAdmin is a middleware that enforces the token has role == "admin".
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "missing authorization", http.StatusUnauthorized)
			return
		}
		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			http.Error(w, "invalid authorization header", http.StatusUnauthorized)
			return
		}
		claims, err := parseAndValidate(parts[1])
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}
		role, _ := claims["role"].(string)
		if role != "admin" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		// attach claims into context for downstream handlers
		ctx := context.WithValue(r.Context(), ContextPrincipal, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// ExtractPrincipal returns token claims from request context if present.
func ExtractPrincipal(r *http.Request) (map[string]interface{}, bool) {
	v := r.Context().Value(ContextPrincipal)
	if v == nil {
		return nil, false
	}
	if m, ok := v.(map[string]interface{}); ok {
		return m, true
	}
	// fallback: if stored as jwt.MapClaims convert
	if m, ok := v.(jwt.MapClaims); ok {
		out := make(map[string]interface{}, len(m))
		for k, vv := range m {
			out[k] = vv
		}
		return out, true
	}
	return nil, false
}

