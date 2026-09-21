package handlers

import (
	"net/http"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/eithansmith/master-of-games/game"
)

const sessionDuration = 30 * 24 * time.Hour

func (s *Server) handleLoginPage(w http.ResponseWriter, r *http.Request) {
	vm := LoginVM{
		Title:     "Sign in",
		Version:   s.meta.Version,
		BuildTime: s.meta.BuildTime,
		StartTime: s.meta.StartTime,
		YearNow:   time.Now().Year(),
		Next:      safeNext(r.URL.Query().Get("next")),
	}
	if err := s.r.HTML(w, "login", "login", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleLoginSubmit(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.renderLogin(w, "", "Invalid form submission.")
		return
	}

	username := strings.TrimSpace(r.FormValue("username"))
	password := r.FormValue("password")
	next := safeNext(r.FormValue("next"))

	user, ok, err := s.store.GetUserByUsername(r.Context(), username)
	if err != nil {
		s.renderLogin(w, next, "Something went wrong. Please try again.")
		return
	}
	if !ok || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		s.renderLogin(w, next, "Invalid username or password.")
		return
	}

	token, err := newSessionToken()
	if err != nil {
		s.renderLogin(w, next, "Something went wrong. Please try again.")
		return
	}

	expiresAt := time.Now().Add(sessionDuration)
	if _, err := s.store.CreateSession(r.Context(), game.Session{
		UserID:    user.ID,
		TokenHash: hashToken(token),
		ExpiresAt: expiresAt,
	}); err != nil {
		s.renderLogin(w, next, "Something went wrong. Please try again.")
		return
	}

	_ = s.store.DeleteExpiredSessions(r.Context())

	setSessionCookie(w, r, token, expiresAt)
	http.Redirect(w, r, next, http.StatusSeeOther)
}

func (s *Server) handleLogout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(sessionCookieName); err == nil {
		_ = s.store.DeleteSession(r.Context(), hashToken(cookie.Value))
	}
	clearSessionCookie(w)
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func (s *Server) renderLogin(w http.ResponseWriter, next, errMsg string) {
	vm := LoginVM{
		Title:     "Sign in",
		Version:   s.meta.Version,
		BuildTime: s.meta.BuildTime,
		StartTime: s.meta.StartTime,
		YearNow:   time.Now().Year(),
		Next:      next,
		FormError: errMsg,
	}
	if err := s.r.HTML(w, "login", "login", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
