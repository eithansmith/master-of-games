package handlers

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const sessionCookieName = "mog_session"

// SessionAuth protects the app with a server-side session cookie.
//
// Unlike Basic Auth, unauthenticated browser requests are redirected to
// /login rather than receiving a 401 - this is a real login flow, not a
// browser-native credential prompt.
//
// /healthz, /readyz, /login and /static/* are exempt: health checks and
// static assets aren't sensitive, and /login must stay reachable or the
// redirect would loop forever.
//
// This app has no CSRF tokens; the session cookie is set SameSite=Lax,
// which blocks it from being sent on cross-site POSTs (the CSRF vector)
// while still working for ordinary top-level navigation. That's sufficient
// because every mutating route here is a same-origin form/HTMX POST - there
// is no cross-origin API surface. This assumption breaks if a future route
// mutates state on a GET request.
func SessionAuth(store Store) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isPublicPath(r.URL.Path) {
				next.ServeHTTP(w, r)
				return
			}

			cookie, err := r.Cookie(sessionCookieName)
			if err != nil {
				redirectToLogin(w, r)
				return
			}

			sess, ok, err := store.GetSessionByTokenHash(r.Context(), hashToken(cookie.Value))
			if err != nil || !ok || sess.ExpiresAt.Before(time.Now()) {
				clearSessionCookie(w)
				redirectToLogin(w, r)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func isPublicPath(path string) bool {
	if path == "/healthz" || path == "/readyz" || path == "/login" {
		return true
	}
	return strings.HasPrefix(path, "/static/")
}

// hashToken returns the hex-encoded SHA-256 hash of a session token, which
// is what's stored in app.sessions - never the raw token itself, so a DB
// leak alone can't be used to replay a live session.
func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// newSessionToken returns a new cryptographically random, cookie-safe token.
func newSessionToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func setSessionCookie(w http.ResponseWriter, r *http.Request, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    token,
		Path:     "/",
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   isHTTPS(r),
		SameSite: http.SameSiteLaxMode,
	})
}

func clearSessionCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     sessionCookieName,
		Value:    "",
		Path:     "/",
		MaxAge:   -1,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
}

// isHTTPS reports whether the request arrived over HTTPS, either directly
// or (as in Fly's setup, which terminates TLS at the edge) via a forwarded
// proto header. Used to decide the cookie's Secure attribute so local
// `go run` over plain http still works.
func isHTTPS(r *http.Request) bool {
	return r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https"
}

func redirectToLogin(w http.ResponseWriter, r *http.Request) {
	next := r.URL.RequestURI()
	target := "/login"
	if next != "" && next != "/" {
		target += "?next=" + url.QueryEscape(next)
	}
	http.Redirect(w, r, target, http.StatusSeeOther)
}

// safeNext validates that next is a same-origin relative path, to avoid an
// open-redirect via a crafted ?next= value. Anything else falls back to "/".
func safeNext(next string) string {
	if next == "" || !strings.HasPrefix(next, "/") || strings.HasPrefix(next, "//") {
		return "/"
	}
	if u, err := url.Parse(next); err != nil || u.Host != "" || u.Scheme != "" {
		return "/"
	}
	return next
}
