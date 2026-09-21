package handlers

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eithansmith/master-of-games/game"
)

func newTestStoreWithSession(t *testing.T) (*game.MemoryStore, string) {
	t.Helper()
	store := game.NewMemoryStore()

	if err := store.CreateUser(t.Context(), "alice", "hash"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	user, ok, err := store.GetUserByUsername(t.Context(), "alice")
	if err != nil || !ok {
		t.Fatalf("GetUserByUsername: ok=%v err=%v", ok, err)
	}

	token := "test-token"
	if _, err := store.CreateSession(t.Context(), game.Session{
		UserID:    user.ID,
		TokenHash: hashToken(token),
		ExpiresAt: time.Now().Add(time.Hour),
	}); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	return store, token
}

func spyHandler(called *bool) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*called = true
		w.WriteHeader(http.StatusOK)
	})
}

// ============================
// SessionAuth
// ============================

func TestSessionAuth_NoCookie_RedirectsToLogin(t *testing.T) {
	store := game.NewMemoryStore()
	var called bool
	mw := SessionAuth(store)(spyHandler(&called))

	req := httptest.NewRequest(http.MethodGet, "/players", nil)
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if called {
		t.Error("expected next handler not to be called")
	}
	if rec.Code != http.StatusSeeOther {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if loc := rec.Header().Get("Location"); loc == "" || loc[:6] != "/login" {
		t.Errorf("got Location %q, want it to start with /login", loc)
	}
}

func TestSessionAuth_InvalidCookie_RedirectsToLogin(t *testing.T) {
	store := game.NewMemoryStore()
	var called bool
	mw := SessionAuth(store)(spyHandler(&called))

	req := httptest.NewRequest(http.MethodGet, "/players", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: "not-a-real-token"})
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if called {
		t.Error("expected next handler not to be called")
	}
	if rec.Code != http.StatusSeeOther {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusSeeOther)
	}
}

func TestSessionAuth_ExpiredSession_RedirectsToLogin(t *testing.T) {
	store := game.NewMemoryStore()
	if err := store.CreateUser(t.Context(), "alice", "hash"); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	user, _, _ := store.GetUserByUsername(t.Context(), "alice")
	token := "expired-token"
	if _, err := store.CreateSession(t.Context(), game.Session{
		UserID:    user.ID,
		TokenHash: hashToken(token),
		ExpiresAt: time.Now().Add(-time.Hour),
	}); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}

	var called bool
	mw := SessionAuth(store)(spyHandler(&called))

	req := httptest.NewRequest(http.MethodGet, "/players", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if called {
		t.Error("expected next handler not to be called")
	}
	if rec.Code != http.StatusSeeOther {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusSeeOther)
	}
}

func TestSessionAuth_ValidSession_PassesThrough(t *testing.T) {
	store, token := newTestStoreWithSession(t)
	var called bool
	mw := SessionAuth(store)(spyHandler(&called))

	req := httptest.NewRequest(http.MethodGet, "/players", nil)
	req.AddCookie(&http.Cookie{Name: sessionCookieName, Value: token})
	rec := httptest.NewRecorder()
	mw.ServeHTTP(rec, req)

	if !called {
		t.Error("expected next handler to be called")
	}
	if rec.Code != http.StatusOK {
		t.Errorf("got status %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestSessionAuth_ExemptPaths_PassThroughWithNoCookie(t *testing.T) {
	store := game.NewMemoryStore()

	for _, path := range []string{"/healthz", "/readyz", "/login", "/static/app.css"} {
		var called bool
		mw := SessionAuth(store)(spyHandler(&called))

		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		mw.ServeHTTP(rec, req)

		if !called {
			t.Errorf("path %s: expected next handler to be called", path)
		}
	}
}
