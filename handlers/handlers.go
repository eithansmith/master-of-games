package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/eithansmith/master-of-games/game"
)

func (s *Server) newHomeVM(showAllGames bool) HomeVM {
	allPlayers := s.store.ListPlayers()
	allTitles := s.store.ListTitles()

	players := make([]game.Player, 0, len(allPlayers))
	for _, p := range allPlayers {
		if p.IsActive {
			players = append(players, p)
		}
	}
	titles := make([]game.Title, 0, len(allTitles))
	for _, t := range allTitles {
		if t.IsActive {
			titles = append(titles, t)
		}
	}

	pMap := make(map[int64]string, len(allPlayers))
	for _, p := range allPlayers {
		pMap[p.ID] = p.Name
	}

	vm := HomeVM{
		Title:       "Master of Games",
		Version:     s.meta.Version,
		BuildTime:   s.meta.BuildTime,
		StartTime:   s.meta.StartTime,
		YearNow:     time.Now().Year(),
		Players:     players,
		PlayerNames: pMap,
		Titles:      titles,
		Games: func() []game.Game {
			gs := s.store.RecentGames(25)
			if showAllGames {
				return gs
			}
			out := make([]game.Game, 0, len(gs))
			for _, g := range gs {
				if g.IsActive {
					out = append(out, g)
				}
			}
			return out
		}(),
		ShowAllGames: showAllGames,
		Form:         s.defaultHomeForm(players, titles),
	}
	return vm
}

func (s *Server) defaultHomeForm(_ []game.Player, titles []game.Title) HomeForm {
	// Default title: first title alphabetically (ListTitles already returns ordered in both stores)
	var titleID int64
	if len(titles) > 0 {
		titleID = titles[0].ID
	}

	return HomeForm{
		TitleID:      titleID,
		PlayedAt:     time.Now().Format("2006-01-02T15:04"),
		Participants: map[int64]bool{},
		Winners:      map[int64]bool{},
		Notes:        "",
	}
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	showAll := r.URL.Query().Get("all") == "1"
	vm := s.newHomeVM(showAll)
	if err := s.r.HTML(w, "home", "home", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleAddGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	allPlayers := s.store.ListPlayers()
	allTitles := s.store.ListTitles()

	players := make([]game.Player, 0, len(allPlayers))
	for _, p := range allPlayers {
		if p.IsActive {
			players = append(players, p)
		}
	}
	titles := make([]game.Title, 0, len(allTitles))
	for _, t := range allTitles {
		if t.IsActive {
			titles = append(titles, t)
		}
	}

	if err := r.ParseForm(); err != nil {
		s.renderHomeWithError(w, "Invalid form submission.", s.defaultHomeForm(players, titles))
		return
	}

	titleIDStr := strings.TrimSpace(r.FormValue("title_id"))
	playedAtStr := strings.TrimSpace(r.FormValue("played_at"))
	notes := strings.TrimSpace(r.FormValue("notes"))

	if playedAtStr == "" {
		playedAtStr = time.Now().Format("2006-01-02T15:04")
	}

	titleID, err := strconv.ParseInt(titleIDStr, 10, 64)
	if err != nil || titleID <= 0 {
		form := HomeForm{
			TitleID:      0,
			PlayedAt:     playedAtStr,
			Participants: parseInt64Map(r.Form["participants"]),
			Winners:      parseInt64Map(r.Form["winners"]),
			Notes:        notes,
		}
		s.renderHomeWithError(w, "Please select a valid game title.", form)
		return
	}

	participantIDs := parseInt64Slice(r.Form["participants"])
	winnerIDs := parseInt64Slice(r.Form["winners"])

	form := HomeForm{
		TitleID:      titleID,
		PlayedAt:     playedAtStr,
		Participants: parseInt64Map(r.Form["participants"]),
		Winners:      parseInt64Map(r.Form["winners"]),
		Notes:        notes,
	}

	playedAt, err := time.Parse("2006-01-02T15:04", playedAtStr)
	if err != nil {
		s.renderHomeWithError(w, "Please provide a valid date/time.", form)
		return
	}

	if !game.IsWeekdayLocal(playedAt) {
		s.renderHomeWithError(w, "Only weekday games are allowed (Mon–Fri).", form)
		return
	}

	if len(participantIDs) == 0 {
		s.renderHomeWithError(w, "Please select at least one participant.", form)
		return
	}

	if len(winnerIDs) == 0 {
		s.renderHomeWithError(w, "Please select at least one winner.", form)
		return
	}

	if !isSubset(winnerIDs, participantIDs) {
		s.renderHomeWithError(w, "Winners must also be selected as participants.", form)
		return
	}

	g := game.Game{
		TitleID:        titleID,
		PlayedAt:       playedAt,
		ParticipantIDs: participantIDs,
		WinnerIDs:      winnerIDs,
		Notes:          notes,
	}

	s.store.AddGame(g)

	// HTMX will swap #main, but a redirect works fine too.
	vm := s.newHomeVM(false)
	vm.Form = s.defaultHomeForm(vm.Players, vm.Titles)
	if err := s.r.HTML(w, "main", "home", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleDeleteGame(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if id > 0 {
		s.store.SetGameActive(id, false)
	}

	vm := s.newHomeVM(false)
	if err := s.r.HTML(w, "main", "home", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleGameToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	_ = r.ParseForm()
	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	active := r.FormValue("active") == "1"

	if id > 0 {
		s.store.SetGameActive(id, active)
	}

	vm := s.newHomeVM(false)
	if err := s.r.HTML(w, "main", "home", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ============================
// Week / Year pages
// ============================

func (s *Server) handleWeekCurrent(w http.ResponseWriter, r *http.Request) {
	year, week := time.Now().ISOWeek()
	http.Redirect(w, r, fmt.Sprintf("/weeks/%d/%d", year, week), http.StatusSeeOther)
}

func (s *Server) handleWeek(w http.ResponseWriter, r *http.Request) {
	// /weeks/{year}/{week}
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 {
		http.NotFound(w, r)
		return
	}
	year, err1 := strconv.Atoi(parts[1])
	week, err2 := strconv.Atoi(parts[2])
	if err1 != nil || err2 != nil || week < 1 || week > 53 {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodPost {
		s.handleWeekTiebreakPost(w, r, year, week)
		return
	}
	s.renderWeek(w, year, week, "")
}

func (s *Server) renderWeek(w http.ResponseWriter, year, week int, formErr string) {
	allPlayers := s.store.ListPlayers()
	pMap := make(map[int64]string, len(allPlayers))
	for _, p := range allPlayers {
		pMap[p.ID] = p.Name
	}

	var years []int
	yNow := time.Now().Year()
	for y := yNow - 2; y <= yNow+1; y++ {
		years = append(years, y)
	}

	weeks := make([]int, 0, isoWeeksInYear(year))
	for wNum := 1; wNum <= isoWeeksInYear(year); wNum++ {
		weeks = append(weeks, wNum)
	}

	py, pw := prevISOWeek(year, week)
	ny, nw := nextISOWeek(year, week)

	ws := game.ComputeWeekStandings(s.store.RecentGames(0), year, week, s.store.GetTiebreaker)

	vm := WeekVM{
		Title:         "Week",
		Version:       s.meta.Version,
		BuildTime:     s.meta.BuildTime,
		StartTime:     s.meta.StartTime,
		YearNow:       yNow,
		Year:          year,
		Week:          week,
		Years:         years,
		Weeks:         weeks,
		PrevYear:      py,
		PrevWeek:      pw,
		HasPrev:       true,
		NextYear:      ny,
		NextWeek:      nw,
		HasNext:       true,
		Players:       allPlayers,
		PlayerNames:   pMap,
		TotalGames:    ws.TotalGames,
		Wins:          ws.Wins,
		TopIDs:        ws.TopIDs,
		WinnerID:      ws.WinnerID,
		TieUnresolved: ws.TieUnresolved,
		FormError:     formErr,
	}

	if err := s.r.HTML(w, "week", "week", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleWeekTiebreakPost(w http.ResponseWriter, r *http.Request, year, week int) {
	if err := r.ParseForm(); err != nil {
		s.renderWeek(w, year, week, "Invalid form submission.")
		return
	}

	ws := game.ComputeWeekStandings(s.store.RecentGames(0), year, week, s.store.GetTiebreaker)

	if ws.TotalGames == 0 {
		s.renderWeek(w, year, week, "No games were played this week—no tiebreaker needed.")
		return
	}
	if len(ws.TopIDs) <= 1 {
		s.renderWeek(w, year, week, "This week is not tied—no tiebreaker needed.")
		return
	}

	winnerID, err := strconv.ParseInt(r.FormValue("winner_id"), 10, 64)
	if err != nil || !containsInt64(ws.TopIDs, winnerID) {
		s.renderWeek(w, year, week, "Please select a valid winner from the tied leaders.")
		return
	}

	tb := game.Tiebreaker{
		Scope:         "weekly",
		ScopeKey:      ws.ScopeKey,
		TiedPlayerIDs: ws.TopIDs,
		WinnerID:      winnerID,
		Method:        "chance",
		DecidedAt:     time.Now(),
	}
	s.store.SetTiebreaker(tb)

	s.renderWeek(w, year, week, "Tiebreaker saved.")
}

func (s *Server) handleYear(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 2 {
		http.NotFound(w, r)
		return
	}
	year, err := strconv.Atoi(parts[1])
	if err != nil {
		http.NotFound(w, r)
		return
	}

	if r.Method == http.MethodPost {
		s.handleYearTiebreakPost(w, r, year)
		return
	}

	s.renderYear(w, year, "")
}

func (s *Server) renderYear(w http.ResponseWriter, year int, formErr string) {
	allPlayers := s.store.ListPlayers()
	pMap := make(map[int64]string, len(allPlayers))
	for _, p := range allPlayers {
		pMap[p.ID] = p.Name
	}

	ys := game.ComputeYearStandings(s.store.RecentGames(0), year, s.store.GetTiebreaker)

	vm := YearVM{
		Title:         "Year",
		Version:       s.meta.Version,
		BuildTime:     s.meta.BuildTime,
		StartTime:     s.meta.StartTime,
		YearNow:       time.Now().Year(),
		Year:          year,
		Players:       allPlayers,
		PlayerNames:   pMap,
		Stats:         ys.Stats,
		Qualifiers:    ys.Qualifiers,
		TopIDs:        ys.TopIDs,
		WinnerID:      ys.WinnerID,
		TieUnresolved: ys.TieUnresolved,
		FormError:     formErr,
	}

	if err := s.r.HTML(w, "year", "year", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleYearTiebreakPost(w http.ResponseWriter, r *http.Request, year int) {
	if err := r.ParseForm(); err != nil {
		s.renderYear(w, year, "Invalid form submission.")
		return
	}

	ys := game.ComputeYearStandings(s.store.RecentGames(0), year, s.store.GetTiebreaker)

	if len(ys.TopIDs) <= 1 {
		s.renderYear(w, year, "This year is not tied—no tiebreaker needed.")
		return
	}

	winnerID, err := strconv.ParseInt(r.FormValue("winner_id"), 10, 64)
	if err != nil || !containsInt64(ys.TopIDs, winnerID) {
		s.renderYear(w, year, "Please select a valid winner from the tied leaders.")
		return
	}

	tb := game.Tiebreaker{
		Scope:         "yearly",
		ScopeKey:      ys.ScopeKey,
		TiedPlayerIDs: ys.TopIDs,
		WinnerID:      winnerID,
		Method:        "chance",
		DecidedAt:     time.Now(),
	}
	s.store.SetTiebreaker(tb)

	s.renderYear(w, year, "Tiebreaker saved.")
}

// ============================
// Players / Titles CRUD
// ============================

func (s *Server) handlePlayers(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			s.renderPlayers(w, "Invalid form submission.")
			return
		}
		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			s.renderPlayers(w, "Name is required.")
			return
		}
		s.store.AddPlayer(name)
		http.Redirect(w, r, "/players", http.StatusSeeOther)
		return
	}

	s.renderPlayers(w, "")
}

func (s *Server) renderPlayers(w http.ResponseWriter, errMsg string) {
	vm := PlayersVM{
		Title:     "Players",
		Version:   s.meta.Version,
		BuildTime: s.meta.BuildTime,
		StartTime: s.meta.StartTime,
		YearNow:   time.Now().Year(),
		Players:   s.store.ListPlayers(),
		FormError: errMsg,
	}
	if err := s.r.HTML(w, "players", "players", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handlePlayerUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/players", http.StatusSeeOther)
		return
	}
	_ = r.ParseForm()
	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	name := strings.TrimSpace(r.FormValue("name"))
	if id <= 0 || name == "" {
		http.Redirect(w, r, "/players", http.StatusSeeOther)
		return
	}
	s.store.UpdatePlayer(id, name)
	http.Redirect(w, r, "/players", http.StatusSeeOther)
}

func (s *Server) handlePlayerToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/players", http.StatusSeeOther)
		return
	}
	_ = r.ParseForm()
	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	active := r.FormValue("active") == "1"
	if id > 0 {
		s.store.SetPlayerActive(id, active)
	}
	http.Redirect(w, r, "/players", http.StatusSeeOther)
}

func (s *Server) handlePlayerDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/players", http.StatusSeeOther)
		return
	}
	_ = r.ParseForm()
	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if id > 0 {
		ok := s.store.SetPlayerActive(id, false)
		if !ok {
			s.renderPlayers(w, "Unable to delete player (they may be referenced by an existing game).")
			return
		}
	}
	http.Redirect(w, r, "/players", http.StatusSeeOther)
}

func (s *Server) handleTitles(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		if err := r.ParseForm(); err != nil {
			s.renderTitles(w, "Invalid form submission.")
			return
		}
		name := strings.TrimSpace(r.FormValue("name"))
		if name == "" {
			s.renderTitles(w, "Name is required.")
			return
		}
		s.store.AddTitle(name)
		http.Redirect(w, r, "/titles", http.StatusSeeOther)
		return
	}
	s.renderTitles(w, "")
}

func (s *Server) renderTitles(w http.ResponseWriter, errMsg string) {
	vm := TitlesVM{
		Title:     "Titles",
		Version:   s.meta.Version,
		BuildTime: s.meta.BuildTime,
		StartTime: s.meta.StartTime,
		YearNow:   time.Now().Year(),
		Titles:    s.store.ListTitles(),
		FormError: errMsg,
	}
	if err := s.r.HTML(w, "titles", "titles", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleTitleUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/titles", http.StatusSeeOther)
		return
	}
	_ = r.ParseForm()
	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	name := strings.TrimSpace(r.FormValue("name"))
	if id <= 0 || name == "" {
		http.Redirect(w, r, "/titles", http.StatusSeeOther)
		return
	}
	s.store.UpdateTitle(id, name)
	http.Redirect(w, r, "/titles", http.StatusSeeOther)
}

func (s *Server) handleTitleToggle(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/titles", http.StatusSeeOther)
		return
	}
	_ = r.ParseForm()
	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	active := r.FormValue("active") == "1"
	if id > 0 {
		s.store.SetTitleActive(id, active)
	}
	http.Redirect(w, r, "/titles", http.StatusSeeOther)
}

func (s *Server) handleTitleDelete(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/titles", http.StatusSeeOther)
		return
	}
	_ = r.ParseForm()
	id, _ := strconv.ParseInt(r.FormValue("id"), 10, 64)
	if id > 0 {
		ok := s.store.SetTitleActive(id, false)
		if !ok {
			s.renderTitles(w, "Unable to delete title (it may be referenced by an existing game).")
			return
		}
	}
	http.Redirect(w, r, "/titles", http.StatusSeeOther)
}

func (s *Server) renderHomeWithError(w http.ResponseWriter, msg string, form HomeForm) {
	vm := s.newHomeVM(false)
	vm.FormError = msg
	vm.Form = form
	if err := s.r.HTML(w, "main", "home", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
