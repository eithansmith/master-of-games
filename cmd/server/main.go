package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/eithansmith/master-of-games/game"
)

var (
	version   = "dev"
	buildTime = ""
	startTime = ""
)

type Server struct {
	tmpl  *template.Template
	store *game.MemoryStore
}

type HomeVM struct {
	Title     string
	Version   string
	BuildTime string
	StartTime string

	Players []string
	Titles  []string
	Games   []game.Game

	// Form feedback
	FormError string
}

func main() {
	addr := env("PORT", "8080")

	s := &Server{
		tmpl: template.Must(template.ParseFiles(
			"web/templates/base.go.html",
			"web/templates/home.go.html",
		)),
		store: game.NewMemoryStore(),
	}

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	mux.HandleFunc("/", s.handleHome)
	mux.HandleFunc("/games", s.handleAddGame)           // POST
	mux.HandleFunc("/games/delete", s.handleDeleteGame) // POST (simple for now)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	srv := &http.Server{
		Addr:              ":" + addr,
		Handler:           logging(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	startTime = time.Now().UTC().Format(time.RFC3339)
	log.Printf("starting master-of-games version=%s buildTime=%s startTime=%s", version, buildTime, startTime)
	log.Fatal(srv.ListenAndServe())
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	vm := HomeVM{
		Title:     "Master of Games",
		Version:   version,
		BuildTime: buildTime,
		StartTime: startTime,
		Players:   game.Players,
		Titles:    game.Titles,
		Games:     s.store.RecentGames(25),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "base", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleAddGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err := r.ParseForm(); err != nil {
		s.renderHomeWithError(w, "Invalid form submission.")
		return
	}

	title := strings.TrimSpace(r.FormValue("title"))
	playedAtStr := strings.TrimSpace(r.FormValue("played_at"))

	participantIDs := parseIntSlice(r.Form["participants"])
	winnerIDs := parseIntSlice(r.Form["winners"])

	playedAt, err := time.Parse("2006-01-02T15:04", playedAtStr)
	if err != nil {
		s.renderHomeWithError(w, "Please provide a valid date/time.")
		return
	}

	// Validation
	if title == "" {
		s.renderHomeWithError(w, "Please choose a game title.")
		return
	}
	if len(participantIDs) < 2 {
		s.renderHomeWithError(w, "Please select at least 2 participants.")
		return
	}
	if len(winnerIDs) < 1 {
		s.renderHomeWithError(w, "Please select at least 1 winner.")
		return
	}
	if !game.IsWeekdayLocal(playedAt) {
		s.renderHomeWithError(w, "Games can only be logged Monday–Friday.")
		return
	}
	if !isSubset(winnerIDs, participantIDs) {
		s.renderHomeWithError(w, "Winners must be a subset of participants.")
		return
	}

	s.store.AddGame(game.Game{
		PlayedAt:       playedAt,
		Title:          title,
		ParticipantIDs: participantIDs,
		WinnerIDs:      winnerIDs,
	})

	// HTMX-friendly: if HTMX request, return the updated main content.
	if r.Header.Get("HX-Request") == "true" {
		s.handleHome(w, r)
		return
	}

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) handleDeleteGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if err := r.ParseForm(); err != nil {
		s.renderHomeWithError(w, "Invalid form submission.")
		return
	}

	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	s.store.DeleteGame(id)

	if r.Header.Get("HX-Request") == "true" {
		s.handleHome(w, r)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) renderHomeWithError(w http.ResponseWriter, msg string) {
	vm := HomeVM{
		Title:     "Master of Games",
		Version:   version,
		BuildTime: buildTime,
		StartTime: startTime,
		Players:   game.Players,
		Titles:    game.Titles,
		Games:     s.store.RecentGames(25),
		FormError: msg,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = s.tmpl.ExecuteTemplate(w, "base", vm)
}

func parseIntSlice(vals []string) []int {
	out := make([]int, 0, len(vals))
	for _, v := range vals {
		i, err := strconv.Atoi(v)
		if err == nil {
			out = append(out, i)
		}
	}
	return out
}

func isSubset(sub, set []int) bool {
	m := map[int]bool{}
	for _, v := range set {
		m[v] = true
	}
	for _, v := range sub {
		if !m[v] {
			return false
		}
	}
	return true
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}
