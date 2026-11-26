package handlers

import (
	"net/http"
	"strings"

	"github.com/gAgu507/tasks-api/internal/auth"
)

func (s *Server) requireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenStr, ok := strings.CutPrefix(r.Header.Get("Authorization"), "Bearer ")
		if !ok || tokenStr == "" {
			w.Header().Set("WWW-Authenticate", "Bearer")
			http.Error(w, "missing or malformed token", http.StatusUnauthorized)
			return
		}

		userID, err := auth.ParseToken(s.jwtSecret, tokenStr)
		if err != nil {
			http.Error(w, "invalid or expired token", http.StatusUnauthorized)
			return
		}

		next(w, r.WithContext(auth.WithUserID(r.Context(), userID)))
	}
}

func currentUser(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, ok := auth.UserID(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return 0, false
	}
	return id, true
}
