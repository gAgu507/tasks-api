package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gAgu507/tasks-api/internal/auth"
)

func TestRequireAuth(t *testing.T) {
	valid, _ := auth.NewToken(testSecret, 1, time.Hour)
	expired, _ := auth.NewToken(testSecret, 1, -time.Minute)
	foreign, _ := auth.NewToken([]byte("another-secret-another-secret-123"), 1, time.Hour)

	tests := []struct {
		name     string
		header   string
		wantCode int
	}{
		{"valid token", "Bearer " + valid, http.StatusOK},
		{"no header", "", http.StatusUnauthorized},
		{"wrong scheme", "Basic abc", http.StatusUnauthorized},
		{"empty token", "Bearer ", http.StatusUnauthorized},
		{"garbage", "Bearer not.a.token", http.StatusUnauthorized},
		{"expired", "Bearer " + expired, http.StatusUnauthorized},
		{"wrong secret", "Bearer " + foreign, http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/tasks", nil)
			if tt.header != "" {
				req.Header.Set("Authorization", tt.header)
			}
			rec := httptest.NewRecorder()
			NewServer(&fakeStore{}, &fakeUsers{}, testSecret).Routes().ServeHTTP(rec, req)

			if rec.Code != tt.wantCode {
				t.Errorf("status = %d, want %d", rec.Code, tt.wantCode)
			}
		})
	}
}

func TestMe_ReturnsUserFromToken(t *testing.T) {
	tok, _ := auth.NewToken(testSecret, 42, time.Hour)
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("Authorization", "Bearer "+tok)
	rec := httptest.NewRecorder()
	NewServer(&fakeStore{}, &fakeUsers{}, testSecret).Routes().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rec.Code)
	}
	var got struct{ ID int }
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("bad json: %v", err)
	}
	if got.ID != 42 {
		t.Errorf("id = %d, want 42", got.ID)
	}
}

func TestPublicRoutesNeedNoToken(t *testing.T) {
	srv := NewServer(&fakeStore{}, &fakeUsers{}, testSecret).Routes()

	rec := httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusOK {
		t.Errorf("GET / status = %d, want 200", rec.Code)
	}

	// плохой JSON должен дать 400, а не 401: сюда токен не нужен
	rec = httptest.NewRecorder()
	srv.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("{")))
	if rec.Code != http.StatusBadRequest {
		t.Errorf("POST /login status = %d, want 400", rec.Code)
	}
}
