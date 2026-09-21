package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/eithansmith/master-of-games/game"
)

// newTestServer builds a Server backed by a MemoryStore, using template
// paths relative to this package's directory (tests run with the package
// dir as the working directory, not the repo root).
func newTestServer(t *testing.T) (*Server, *game.MemoryStore) {
	t.Helper()

	store := game.NewMemoryStore()
	r := NewRenderer(RendererConfig{
		Base:          "../web/templates/base.go.html",
		Home:          "../web/templates/home.go.html",
		Week:          "../web/templates/week.go.html",
		Year:          "../web/templates/year.go.html",
		YearRace:      "../web/templates/year_race.go.html",
		YearRaceChart: "../web/templates/year_race_chart.go.html",
		Players:       "../web/templates/players.go.html",
		Titles:        "../web/templates/titles.go.html",
		Login:         "../web/templates/login.go.html",
	})

	return &Server{r: r, store: store, meta: Meta{Version: "test"}}, store
}

func seedTestUser(t *testing.T, store *game.MemoryStore, username, password string) {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("GenerateFromPassword: %v", err)
	}
	if err := store.CreateUser(t.Context(), username, string(hash)); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
}

func postLogin(s *Server, form url.Values) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	s.handleLoginSubmit(rec, req)
	return rec
}

// ============================
// handleLoginSubmit
// ============================

func TestHandleLoginSubmit_ValidCredentials_SetsCookieAndRedirects(t *testing.T) {
	s, store := newTestServer(t)
	seedTestUser(t, store, "alice", "secret123")

	rec := postLogin(s, url.Values{"username": {"alice"}, "password": {"secret123"}})

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if loc := rec.Header().Get("Location"); loc != "/" {
		t.Errorf("got redirect %q, want /", loc)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != sessionCookieName {
		t.Fatalf("got cookies %v, want one %q cookie", cookies, sessionCookieName)
	}
	if !cookies[0].HttpOnly {
		t.Error("expected session cookie to be HttpOnly")
	}
	if cookies[0].Value == "" {
		t.Error("expected session cookie to have a value")
	}
}

func TestHandleLoginSubmit_InvalidPassword_ReRendersWithError(t *testing.T) {
	s, store := newTestServer(t)
	seedTestUser(t, store, "alice", "secret123")

	rec := postLogin(s, url.Values{"username": {"alice"}, "password": {"wrong"}})

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Invalid username or password") {
		t.Error("expected body to contain the form error")
	}
	if cookies := rec.Result().Cookies(); len(cookies) != 0 {
		t.Errorf("expected no cookies to be set, got %v", cookies)
	}
}

func TestHandleLoginSubmit_UnknownUser_ReRendersWithError(t *testing.T) {
	s, _ := newTestServer(t)

	rec := postLogin(s, url.Values{"username": {"nobody"}, "password": {"whatever"}})

	if rec.Code != http.StatusOK {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusOK)
	}
	if !strings.Contains(rec.Body.String(), "Invalid username or password") {
		t.Error("expected body to contain the form error")
	}
}

func TestHandleLoginSubmit_NextParam_RedirectsToNext(t *testing.T) {
	s, store := newTestServer(t)
	seedTestUser(t, store, "alice", "secret123")

	rec := postLogin(s, url.Values{"username": {"alice"}, "password": {"secret123"}, "next": {"/players"}})

	if loc := rec.Header().Get("Location"); loc != "/players" {
		t.Errorf("got redirect %q, want /players", loc)
	}
}

func TestHandleLoginSubmit_UnsafeNext_FallsBackToRoot(t *testing.T) {
	s, store := newTestServer(t)
	seedTestUser(t, store, "alice", "secret123")

	rec := postLogin(s, url.Values{"username": {"alice"}, "password": {"secret123"}, "next": {"https://evil.example/steal"}})

	if loc := rec.Header().Get("Location"); loc != "/" {
		t.Errorf("got redirect %q, want / (open-redirect guard)", loc)
	}
}

// ============================
// handleLogout
// ============================

func TestHandleLogout_ClearsCookieAndDeletesSession(t *testing.T) {
	s, store := newTestServer(t)
	seedTestUser(t, store, "alice", "secret123")

	loginRec := postLogin(s, url.Values{"username": {"alice"}, "password": {"secret123"}})
	sessionCookie := loginRec.Result().Cookies()[0]

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	req.AddCookie(sessionCookie)
	rec := httptest.NewRecorder()
	s.handleLogout(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if loc := rec.Header().Get("Location"); loc != "/login" {
		t.Errorf("got redirect %q, want /login", loc)
	}

	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].MaxAge >= 0 {
		t.Fatalf("expected session cookie to be cleared, got %v", cookies)
	}

	_, ok, err := store.GetSessionByTokenHash(req.Context(), hashToken(sessionCookie.Value))
	if err != nil {
		t.Fatalf("GetSessionByTokenHash: %v", err)
	}
	if ok {
		t.Error("expected session to be deleted from the store")
	}
}

func TestHandleLogout_NoCookie_StillRedirects(t *testing.T) {
	s, _ := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	rec := httptest.NewRecorder()
	s.handleLogout(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("got status %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if loc := rec.Header().Get("Location"); loc != "/login" {
		t.Errorf("got redirect %q, want /login", loc)
	}
}
