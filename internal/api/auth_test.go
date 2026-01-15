package api

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

func TestGenerateAndValidateToken_Admin(t *testing.T) {
	os.Setenv("AUTH_JWT_SECRET", "test-secret")
	defer os.Unsetenv("AUTH_JWT_SECRET")

	token, err := GenerateToken("alice", "admin", 24*time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}

	// valid token should pass middleware
	nextCalled := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()

	handler := RequireAdmin(next)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}
	if !nextCalled {
		t.Fatalf("next handler was not called")
	}
}

func TestRequireAdmin_RejectsNonAdmin(t *testing.T) {
	os.Setenv("AUTH_JWT_SECRET", "test-secret")
	defer os.Unsetenv("AUTH_JWT_SECRET")
	token, err := GenerateToken("bob", "user", 24*time.Hour)
	if err != nil {
		t.Fatalf("GenerateToken error: %v", err)
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rr := httptest.NewRecorder()
	handler := RequireAdmin(next)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 for non-admin, got %d", rr.Code)
	}
}

func TestRequireAdmin_MissingOrInvalidToken(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	// missing token
	req := httptest.NewRequest("GET", "/", nil)
	rr := httptest.NewRecorder()
	handler := RequireAdmin(next)
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing token, got %d", rr.Code)
	}

	// invalid token
	req2 := httptest.NewRequest("GET", "/", nil)
	req2.Header.Set("Authorization", "Bearer invalid.token.value")
	rr2 := httptest.NewRecorder()
	handler.ServeHTTP(rr2, req2)
	if rr2.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for invalid token, got %d", rr2.Code)
	}
}

