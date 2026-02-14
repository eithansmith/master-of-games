package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/eithansmith/master-of-games/game"
)

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	vm := HomeVM{
		Title:     "Master of Games",
		Version:   s.meta.Version,
		BuildTime: s.meta.BuildTime,
		StartTime: s.meta.StartTime,
		YearNow:   time.Now().Year(),
		Players:   game.Players,
		Titles:    game.Titles,
		Games:     s.store.RecentGames(25),
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.homeTmpl.ExecuteTemplate(w, "home", vm); err != nil {
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

func (s *Server) handleWeekCurrent(w http.ResponseWriter, r *http.Request) {
	now := time.Now()
	year, week := now.ISOWeek()
	http.Redirect(w, r, "/weeks/"+strconv.Itoa(year)+"/"+strconv.Itoa(week), http.StatusSeeOther)
}

func (s *Server) handleWeek(w http.ResponseWriter, r *http.Request) {
	// GET  /weeks/{year}/{week}
	// POST /weeks/{year}/{week}/tiebreak
	path := strings.TrimPrefix(r.URL.Path, "/weeks/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) < 2 {
		http.NotFound(w, r)
		return
	}

	year, err1 := strconv.Atoi(parts[0])
	week, err2 := strconv.Atoi(parts[1])
	if err1 != nil || err2 != nil || week < 1 || week > 53 {
		http.NotFound(w, r)
		return
	}

	if len(parts) == 3 && parts[2] == "tiebreak" {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		s.handleWeekTiebreakPost(w, r, year, week)
		return
	}

	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	s.renderWeek(w, year, week, "")
}

func (s *Server) handleWeekTiebreakPost(w http.ResponseWriter, r *http.Request, year, week int) {
	if err := r.ParseForm(); err != nil {
		s.renderWeek(w, year, week, "Invalid form submission.")
		return
	}

	ws := game.ComputeWeekStandings(
		s.store.RecentGames(0),
		year,
		week,
		len(game.Players),
		s.store.GetTiebreaker,
	)

	if ws.TotalGames == 0 {
		s.renderWeek(w, year, week, "No games were played this week—no tiebreaker needed.")
		return
	}
	if len(ws.TopIDs) <= 1 {
		s.renderWeek(w, year, week, "This week is not tied—no tiebreaker needed.")
		return
	}

	winnerID, err := strconv.Atoi(r.FormValue("winner_id"))
	if err != nil {
		s.renderWeek(w, year, week, "Please select a winner.")
		return
	}
	if !containsInt(ws.TopIDs, winnerID) {
		s.renderWeek(w, year, week, "Selected winner is not part of the tied group.")
		return
	}

	scopeKey := fmt.Sprintf("%04d-W%02d", year, week)

	s.store.SetTiebreaker(game.Tiebreaker{
		Scope:         "weekly",
		ScopeKey:      scopeKey,
		TiedPlayerIDs: ws.TopIDs,
		WinnerID:      winnerID,
		Method:        "chance",
		DecidedAt:     time.Now().UTC(),
	})

	s.renderWeek(w, year, week, "")
}

func healthz(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) renderWeek(w http.ResponseWriter, year, week int, formErr string) {
	ws := game.ComputeWeekStandings(
		s.store.RecentGames(0),
		year,
		week,
		len(game.Players),
		s.store.GetTiebreaker,
	)

	vm := WeekVM{
		Title:     "Master of Games",
		Version:   s.meta.Version,
		BuildTime: s.meta.BuildTime,
		StartTime: s.meta.StartTime,
		YearNow:   time.Now().Year(),

		Year: year,
		Week: week,

		Players: game.Players,

		TotalGames: ws.TotalGames,
		Wins:       ws.Wins,

		TopIDs:        ws.TopIDs,
		WinnerID:      ws.WinnerID,
		TieUnresolved: ws.TotalGames > 0 && len(ws.TopIDs) > 1 && ws.WinnerID == nil,

		FormError: formErr,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.weekTmpl.ExecuteTemplate(w, "week", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleYear(w http.ResponseWriter, r *http.Request) {
	// GET  /years/{year}
	// POST /years/{year}/tiebreak
	path := strings.TrimPrefix(r.URL.Path, "/years/")
	parts := strings.Split(strings.Trim(path, "/"), "/")

	if len(parts) < 1 || parts[0] == "" {
		http.NotFound(w, r)
		return
	}

	year, err := strconv.Atoi(parts[0])
	if err != nil || year < 2000 || year > 3000 {
		http.NotFound(w, r)
		return
	}

	if len(parts) == 2 && parts[1] == "tiebreak" {
		if r.Method != http.MethodPost {
			http.NotFound(w, r)
			return
		}
		s.handleYearTiebreakPost(w, r, year)
		return
	}

	if r.Method != http.MethodGet {
		http.NotFound(w, r)
		return
	}

	s.renderYear(w, year, "")
}

func (s *Server) handleYearTiebreakPost(w http.ResponseWriter, r *http.Request, year int) {
	if err := r.ParseForm(); err != nil {
		s.renderYear(w, year, "Invalid form submission.")
		return
	}

	ys := game.ComputeYearStandings(
		s.store.RecentGames(0),
		year,
		len(game.Players),
		s.store.GetTiebreaker,
	)

	if len(ys.TopIDs) <= 1 {
		s.renderYear(w, year, "This year is not tied—no tiebreaker needed.")
		return
	}

	winnerID, err := strconv.Atoi(r.FormValue("winner_id"))
	if err != nil {
		s.renderYear(w, year, "Please select a winner.")
		return
	}
	if !containsInt(ys.TopIDs, winnerID) {
		s.renderYear(w, year, "Selected winner is not part of the tied group.")
		return
	}

	s.store.SetTiebreaker(game.Tiebreaker{
		Scope:         "yearly",
		ScopeKey:      strconv.Itoa(year),
		TiedPlayerIDs: ys.TopIDs,
		WinnerID:      winnerID,
		Method:        "chance",
		DecidedAt:     time.Now().UTC(),
	})

	s.renderYear(w, year, "")
}

func (s *Server) renderYear(w http.ResponseWriter, year int, formErr string) {
	ys := game.ComputeYearStandings(
		s.store.RecentGames(0),
		year,
		len(game.Players),
		s.store.GetTiebreaker,
	)

	vm := YearVM{
		Title:     "Master of Games",
		Version:   s.meta.Version,
		BuildTime: s.meta.BuildTime,
		StartTime: s.meta.StartTime,
		YearNow:   time.Now().Year(),

		Year:    year,
		Players: game.Players,

		Stats:         ys.Stats,
		Qualifiers:    ys.Qualifiers,
		TopIDs:        ys.TopIDs,
		WinnerID:      ys.WinnerID,
		TieUnresolved: len(ys.TopIDs) > 1 && ys.WinnerID == nil,

		FormError: formErr,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.yearTmpl.ExecuteTemplate(w, "year", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) renderHomeWithError(w http.ResponseWriter, msg string) {
	vm := HomeVM{
		Title:     "Master of Games",
		Version:   s.meta.Version,
		BuildTime: s.meta.BuildTime,
		StartTime: s.meta.StartTime,
		YearNow:   time.Now().Year(),
		Players:   game.Players,
		Titles:    game.Titles,
		Games:     s.store.RecentGames(25),
		FormError: msg,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_ = s.homeTmpl.ExecuteTemplate(w, "home", vm)
}
