package handlers

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"
)

// BasicAuth protects the app with HTTP Basic Auth credentials read from env vars.
// Env vars:
//
//	BASIC_AUTH_USER
//	BASIC_AUTH_PASS
//
// Behavior:
// - /healthz is allowed through (so Render health checks won't break).
// - If creds are missing, it fails closed (500) to avoid accidentally exposing prod.
func BasicAuth(next http.Handler) http.Handler {
	user := strings.TrimSpace(os.Getenv("BASIC_AUTH_USER"))
	pass := os.Getenv("BASIC_AUTH_PASS")

	if user == "" || pass == "" {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			http.Error(w, "auth not configured", http.StatusInternalServerError)
		})
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Allow health checks without auth.
		if r.URL.Path == "/healthz" {
			next.ServeHTTP(w, r)
			return
		}

		u, p, ok := r.BasicAuth()
		if !ok ||
			subtle.ConstantTimeCompare([]byte(u), []byte(user)) != 1 ||
			subtle.ConstantTimeCompare([]byte(p), []byte(pass)) != 1 {
			w.Header().Set("WWW-Authenticate", `Basic realm="Master of Games"`)
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		next.ServeHTTP(w, r)
	})
}
