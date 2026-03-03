package handlers

import (
	"context"
	"fmt"
	"html/template"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/eithansmith/master-of-games/game"
)

func (s *Server) newHomeVM(ctx context.Context) (HomeVM, error) {
	allPlayers, err := s.store.ListPlayers(ctx)
	if err != nil {
		return HomeVM{}, err
	}

	allTitles, err := s.store.ListTitles(ctx)
	if err != nil {
		return HomeVM{}, err
	}

	players := activePlayers(allPlayers)
	titles := activeTitles(allTitles)

	pMap := make(map[int64]string, len(allPlayers))
	for _, p := range allPlayers {
		pMap[p.ID] = p.Name
	}

	recentGames, err := s.store.RecentGames(ctx, 25)
	if err != nil {
		return HomeVM{}, err
	}

	vm := HomeVM{
		Title:        "Master of Games",
		Version:      s.meta.Version,
		BuildTime:    s.meta.BuildTime,
		StartTime:    s.meta.StartTime,
		YearNow:      time.Now().Year(),
		Players:      players,
		PlayerNames:  pMap,
		Titles:       titles,
		Games:        recentGames,
		ShowAllGames: true,
		Form:         s.defaultHomeForm(players, titles),
	}
	return vm, nil
}

func (s *Server) defaultHomeForm(_ []game.Player, _ []game.Title) HomeForm {
	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		loc = time.UTC
	}

	return HomeForm{
		TitleID:      0,
		PlayedAt:     time.Now().In(loc).Format("2006-01-02T15:04"),
		Participants: map[int64]bool{},
		Winners:      map[int64]bool{},
		Notes:        "",
	}
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	vm, err := s.newHomeVM(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if err := s.r.HTML(w, "home", "home", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleAddGame(w http.ResponseWriter, r *http.Request) {
	allPlayers, err := s.store.ListPlayers(r.Context())
	if err != nil {
		s.renderHomeWithError(r.Context(), w, "Unable to load player list.", s.defaultHomeForm(nil, nil))
		return
	}

	allTitles, err := s.store.ListTitles(r.Context())
	if err != nil {
		s.renderHomeWithError(r.Context(), w, "Unable to load title list.", s.defaultHomeForm(nil, nil))
		return
	}

	players := activePlayers(allPlayers)
	titles := activeTitles(allTitles)

	if err := r.ParseForm(); err != nil {
		s.renderHomeWithError(r.Context(), w, "Invalid form submission.", s.defaultHomeForm(players, titles))
		return
	}

	titleIDStr := strings.TrimSpace(r.FormValue("title_id"))
	playedAtStr := strings.TrimSpace(r.FormValue("played_at"))
	notes := strings.TrimSpace(r.FormValue("notes"))

	loc, err := time.LoadLocation("America/Chicago")
	if err != nil {
		loc = time.UTC
	}

	if playedAtStr == "" {
		playedAtStr = time.Now().In(loc).Format("2006-01-02T15:04")
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
		s.renderHomeWithError(r.Context(), w, "Please select a valid game title.", form)
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
		s.renderHomeWithError(r.Context(), w, "Please provide a valid date/time.", form)
		return
	}

	if !game.IsWeekdayLocal(playedAt) {
		s.renderHomeWithError(r.Context(), w, "Only weekday games are allowed (Mon–Fri).", form)
		return
	}

	if len(participantIDs) == 0 {
		s.renderHomeWithError(r.Context(), w, "Please select at least one participant.", form)
		return
	}

	if len(winnerIDs) == 0 {
		s.renderHomeWithError(r.Context(), w, "Please select at least one winner.", form)
		return
	}

	if !isSubset(winnerIDs, participantIDs) {
		s.renderHomeWithError(r.Context(), w, "Winners must also be selected as participants.", form)
		return
	}

	g := game.Game{
		TitleID:        titleID,
		PlayedAt:       playedAt,
		ParticipantIDs: participantIDs,
		WinnerIDs:      winnerIDs,
		Notes:          notes,
	}

	_, err = s.store.AddGame(r.Context(), g)
	if err != nil {
		s.renderHomeWithError(r.Context(), w, "Unable to save game.", form)
		return
	}

	// HTMX will swap #main, but a redirect works fine too.
	vm, err := s.newHomeVM(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vm.Form = s.defaultHomeForm(vm.Players, vm.Titles)
	setToast(w, "Game saved.")
	if err := s.r.HTML(w, "main", "home", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleDeleteGame(w http.ResponseWriter, r *http.Request) {
	id, err := pathInt64(r, "id")
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	err = s.store.SetGameActive(r.Context(), id, false)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vm, err := s.newHomeVM(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setToast(w, "Game removed.")
	if err := s.r.HTML(w, "main", "home", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleGameToggle(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form submission.", http.StatusBadRequest)
		return
	}
	id, err := pathInt64(r, "id")
	if err != nil || id <= 0 {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	active := r.FormValue("active") == "1"
	err = s.store.SetGameActive(r.Context(), id, active)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vm, err := s.newHomeVM(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	msg := "Game deactivated."
	if active {
		msg = "Game activated."
	}
	setToast(w, msg)
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
	year, ok1 := pathInt(r, "year")
	week, ok2 := pathInt(r, "week")
	if !ok1 || !ok2 || week < 1 || week > 53 {
		http.NotFound(w, r)
		return
	}

	s.renderWeek(r.Context(), w, year, week, "")
}

func (s *Server) handleWeekTiebreak(w http.ResponseWriter, r *http.Request) {
	year, ok1 := pathInt(r, "year")
	week, ok2 := pathInt(r, "week")
	if !ok1 || !ok2 || week < 1 || week > 53 {
		http.NotFound(w, r)
		return
	}
	s.handleWeekTiebreakPost(w, r, year, week)
}

func (s *Server) renderWeek(ctx context.Context, w http.ResponseWriter, year, week int, formErr string) {
	allPlayers, err := s.store.ListPlayers(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pMap := make(map[int64]game.Player, len(allPlayers))
	for _, p := range allPlayers {
		pMap[p.ID] = p
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

	gamesByWeek, err := s.store.GetWeek(ctx, year, week)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	getTB := func(scope, scopeKey string) (game.Tiebreaker, bool, error) {
		return s.store.GetTiebreaker(ctx, scope, scopeKey)
	}
	ws := game.ComputeWeekStandings(gamesByWeek, year, week, getTB)

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
		PlayerMap:     pMap,
		TotalGames:    ws.TotalGames,
		Wins:          ws.Wins,
		TotalWins:     ws.TotalWins,
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
		s.renderWeek(r.Context(), w, year, week, "Invalid form submission.")
		return
	}

	gamesByWeek, err := s.store.GetWeek(r.Context(), year, week)
	if err != nil {
		s.renderWeek(r.Context(), w, year, week, "Unable to load games for this week.")
		return
	}

	getTB := func(scope, scopeKey string) (game.Tiebreaker, bool, error) {
		return s.store.GetTiebreaker(r.Context(), scope, scopeKey)
	}
	ws := game.ComputeWeekStandings(gamesByWeek, year, week, getTB)

	if ws.TotalGames == 0 {
		s.renderWeek(r.Context(), w, year, week, "No games were played this week—no tiebreaker needed.")
		return
	}
	if len(ws.TopIDs) <= 1 {
		s.renderWeek(r.Context(), w, year, week, "This week is not tied—no tiebreaker needed.")
		return
	}

	winnerID, err := strconv.ParseInt(r.FormValue("winner_id"), 10, 64)
	if err != nil || !containsInt64(ws.TopIDs, winnerID) {
		s.renderWeek(r.Context(), w, year, week, "Please select a valid winner from the tied leaders.")
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
	err = s.store.SetTiebreaker(r.Context(), tb)
	if err != nil {
		s.renderWeek(r.Context(), w, year, week, err.Error())
		return
	}

	s.renderWeek(r.Context(), w, year, week, "Tiebreaker saved.")
}

func (s *Server) handleYear(w http.ResponseWriter, r *http.Request) {
	year, ok := pathInt(r, "year")
	if !ok {
		http.NotFound(w, r)
		return
	}

	s.renderYear(r.Context(), w, year, "")
}

func (s *Server) handleYearTiebreak(w http.ResponseWriter, r *http.Request) {
	year, ok := pathInt(r, "year")
	if !ok {
		http.NotFound(w, r)
		return
	}
	s.handleYearTiebreakPost(w, r, year)
}

func (s *Server) renderYear(ctx context.Context, w http.ResponseWriter, year int, formErr string) {
	allPlayers, err := s.store.ListPlayers(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	pMap := make(map[int64]game.Player, len(allPlayers))
	for _, p := range allPlayers {
		pMap[p.ID] = p
	}

	gamesByYear, err := s.store.GetYear(ctx, year)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	getTB := func(scope, scopeKey string) (game.Tiebreaker, bool, error) {
		return s.store.GetTiebreaker(ctx, scope, scopeKey)
	}
	ys := game.ComputeYearStandings(gamesByYear, year, getTB)

	vm := YearVM{
		Title:         "Year",
		Version:       s.meta.Version,
		BuildTime:     s.meta.BuildTime,
		StartTime:     s.meta.StartTime,
		YearNow:       time.Now().Year(),
		Year:          year,
		Players:       allPlayers,
		PlayerMap:     pMap,
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
		s.renderYear(r.Context(), w, year, "Invalid form submission.")
		return
	}

	gamesByYear, err := s.store.GetYear(r.Context(), year)
	if err != nil {
		s.renderYear(r.Context(), w, year, "Unable to load games for this year.")
		return
	}

	getTB := func(scope, scopeKey string) (game.Tiebreaker, bool, error) {
		return s.store.GetTiebreaker(r.Context(), scope, scopeKey)
	}
	ys := game.ComputeYearStandings(gamesByYear, year, getTB)

	if len(ys.TopIDs) <= 1 {
		s.renderYear(r.Context(), w, year, "This year is not tied—no tiebreaker needed.")
		return
	}

	winnerID, err := strconv.ParseInt(r.FormValue("winner_id"), 10, 64)
	if err != nil || !containsInt64(ys.TopIDs, winnerID) {
		s.renderYear(r.Context(), w, year, "Please select a valid winner from the tied leaders.")
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
	err = s.store.SetTiebreaker(r.Context(), tb)
	if err != nil {
		s.renderYear(r.Context(), w, year, err.Error())
		return
	}

	s.renderYear(r.Context(), w, year, "Tiebreaker saved.")
}

func (s *Server) handleYearRace(w http.ResponseWriter, r *http.Request) {
	year, ok := pathInt(r, "year")
	if !ok {
		http.NotFound(w, r)
		return
	}

	vm := YearRaceVM{
		Title:     "Year Race",
		Version:   s.meta.Version,
		BuildTime: s.meta.BuildTime,
		StartTime: s.meta.StartTime,
		YearNow:   time.Now().Year(),
		Year:      year,
	}

	if err := s.r.HTML(w, "year_race", "year_race", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleYearRaceChart(w http.ResponseWriter, r *http.Request) {
	year, ok := pathInt(r, "year")
	if !ok {
		http.NotFound(w, r)
		return
	}

	games, err := s.store.GetYear(r.Context(), year)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	players, err := s.store.ListPlayers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	race := game.ComputeYearRace(games, year, game.RaceMetricWins, 5, players)

	vm := buildYearRaceChartVM(race)

	if err := s.r.HTML(w, "year_race_chart", "year_race_chart", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func buildYearRaceChartVM(race game.YearRace) yearRaceChartVM {
	const (
		w   = 900.0
		h   = 420.0
		pad = 44.0
	)

	vm := yearRaceChartVM{
		SvgView: fmt.Sprintf("0 0 %.0f %.0f", w, h),
		Width:   w,
		Height:  h,
		Pad:     pad,
		Weeks:   race.Weeks,
	}

	// Max value across all series
	maximum := 0.0
	for _, s := range race.Series {
		for _, v := range s.Values {
			if v > maximum {
				maximum = v
			}
		}
	}

	// max is the actual maximum value across all series
	if maximum < 1 {
		maximum = 1
	}

	// Simplified axis max:
	// - keep it integer
	// - do NOT round 3 up to 5
	axisMax := math.Ceil(maximum)

	// Optional tiny headroom: if axisMax == max and max <= 10, add 1
	// so lines don’t sit exactly on the top.
	// If you don't want headroom, delete this block.
	if axisMax == maximum && axisMax <= 10 {
		axisMax += 1
	}

	vm.Max = axisMax

	n := len(race.Weeks)
	plotW := w - 2*pad
	plotH := h - 2*pad

	xAt := func(i int) float64 {
		if n <= 1 {
			return pad
		}
		return pad + (plotW * float64(i) / float64(n-1))
	}

	yAt := func(v float64) float64 {
		return (h - pad) - (plotH * (v / axisMax))
	}

	// ---- Y ticks (0..maximum) ----
	// choose 4 intervals (5 ticks). Round maximum to something nice.
	niceMax := maximum
	if maximum <= 10 {
		niceMax = maximum
	} else {
		niceMax = niceCeil(maximum)
	}
	if niceMax < 1 {
		niceMax = 1
	}
	vm.Max = niceMax

	// ---- Y ticks: integer ticks from 0..axisMax ----
	vm.YTicks = nil

	plotH = h - 2*pad
	yAtAxis := func(v float64) float64 {
		return (h - pad) - (plotH * (v / axisMax))
	}

	// If axisMax is big, don't draw 50 tick labels.
	// We'll choose a step size (1,2,5,10...) based on axisMax.
	step := yTickStep(int(axisMax))

	for v := 0; v <= int(axisMax); v += step {
		y := yAtAxis(float64(v))
		vm.YTicks = append(vm.YTicks, yearRaceTick{
			X:     pad - 8,
			Y:     y + 4,
			Label: fmt.Sprintf("%d", v),
		})
	}

	// ---- X ticks (start, quarter points, end) ----
	if n > 0 {
		xTickIdx := uniqueSortedInts([]int{
			0,
			n / 4,
			n / 2,
			(3 * n) / 4,
			n - 1,
		})

		for _, idx := range xTickIdx {
			week := race.Weeks[idx]
			x := xAt(idx)
			vm.XTicks = append(vm.XTicks, yearRaceTick{
				X:     x,
				Y:     h - pad + 18,
				Label: fmt.Sprintf("W%02d", week),
			})
		}
	}

	// ---- Series ----
	// Use distinct hues so the lines are clearly different.
	for i, s := range race.Series {
		color := template.CSS(seriesColor(i))

		ser := yearRaceSeriesVM{
			Name:  s.Name,
			Color: color,
		}

		// Path + points
		for j, v := range s.Values {
			x := xAt(j)
			y := yAt(v)

			ser.Points = append(ser.Points, yearRacePointVM{
				X:     x,
				Y:     y,
				Title: fmt.Sprintf("%s — Week %d: %.0f wins", s.Name, race.Weeks[j], v),
			})

			if j == 0 {
				ser.Path = fmt.Sprintf("M %.2f %.2f", x, y)
			} else {
				ser.Path += fmt.Sprintf(" L %.2f %.2f", x, y)
			}
		}

		vm.Series = append(vm.Series, ser)
	}

	return vm
}

// ============================
// Players / Titles CRUD
// ============================

func (s *Server) handlePlayers(w http.ResponseWriter, r *http.Request) {
	players, err := s.store.ListPlayers(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vm := PlayersVM{
		Title:     "Players",
		Version:   s.meta.Version,
		BuildTime: s.meta.BuildTime,
		StartTime: s.meta.StartTime,
		YearNow:   time.Now().Year(),
		Players:   players,
	}
	if err := s.r.HTML(w, "players", "players", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handlePlayersPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.renderPlayers(r.Context(), w, "Invalid form submission.")
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		s.renderPlayers(r.Context(), w, "Name is required.")
		return
	}
	if _, err := s.store.AddPlayer(r.Context(), name); err != nil {
		s.renderPlayers(r.Context(), w, err.Error())
		return
	}
	setToast(w, "Player added.")
	s.renderPlayers(r.Context(), w, "")
}

func (s *Server) renderPlayers(ctx context.Context, w http.ResponseWriter, errMsg string) {
	players, err := s.store.ListPlayers(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vm := PlayersVM{
		Title:     "Players",
		Version:   s.meta.Version,
		BuildTime: s.meta.BuildTime,
		StartTime: s.meta.StartTime,
		YearNow:   time.Now().Year(),
		Players:   players,
		FormError: errMsg,
	}
	if err := s.r.HTML(w, "main", "players", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handlePlayerUpdate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/players", http.StatusSeeOther)
		return
	}
	id, err := pathInt64(r, "id")
	if err != nil || id <= 0 {
		http.Redirect(w, r, "/players", http.StatusSeeOther)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Redirect(w, r, "/players", http.StatusSeeOther)
		return
	}
	err = s.store.UpdatePlayer(r.Context(), id, name)
	if err != nil {
		s.renderPlayers(r.Context(), w, err.Error())
		return
	}
	setToast(w, "Player updated.")
	s.renderPlayers(r.Context(), w, "")
}

func (s *Server) handlePlayerToggle(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/players", http.StatusSeeOther)
		return
	}
	id, err := pathInt64(r, "id")
	if err != nil || id <= 0 {
		http.Redirect(w, r, "/players", http.StatusSeeOther)
		return
	}
	active := r.FormValue("active") == "1"
	err = s.store.SetPlayerActive(r.Context(), id, active)
	if err != nil {
		s.renderPlayers(r.Context(), w, err.Error())
		return
	}

	msg := "Player deactivated."
	if active {
		msg = "Player activated."
	}
	setToast(w, msg)
	s.renderPlayers(r.Context(), w, "")
}

func (s *Server) handlePlayerDelete(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/players", http.StatusSeeOther)
		return
	}
	id, err := pathInt64(r, "id")
	if err != nil || id <= 0 {
		http.Redirect(w, r, "/players", http.StatusSeeOther)
		return
	}
	err = s.store.SetPlayerActive(r.Context(), id, false)
	if err != nil {
		s.renderPlayers(r.Context(), w, "Unable to delete player (they may be referenced by an existing game).")
		return
	}

	setToast(w, "Player deactivated.")
	s.renderPlayers(r.Context(), w, "")
}

func (s *Server) handleTitles(w http.ResponseWriter, r *http.Request) {
	titles, err := s.store.ListTitles(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vm := TitlesVM{
		Title:     "Titles",
		Version:   s.meta.Version,
		BuildTime: s.meta.BuildTime,
		StartTime: s.meta.StartTime,
		YearNow:   time.Now().Year(),
		Titles:    titles,
	}
	if err := s.r.HTML(w, "titles", "titles", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleTitlesPost(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.renderTitles(r.Context(), w, "Invalid form submission.")
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		s.renderTitles(r.Context(), w, "Name is required.")
		return
	}
	if _, err := s.store.AddTitle(r.Context(), name); err != nil {
		s.renderTitles(r.Context(), w, err.Error())
		return
	}
	setToast(w, "Title added.")
	s.renderTitles(r.Context(), w, "")
}

func (s *Server) renderTitles(ctx context.Context, w http.ResponseWriter, errMsg string) {
	titles, err := s.store.ListTitles(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	vm := TitlesVM{
		Title:     "Titles",
		Version:   s.meta.Version,
		BuildTime: s.meta.BuildTime,
		StartTime: s.meta.StartTime,
		YearNow:   time.Now().Year(),
		Titles:    titles,
		FormError: errMsg,
	}
	if err := s.r.HTML(w, "main", "titles", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (s *Server) handleTitleUpdate(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/titles", http.StatusSeeOther)
		return
	}
	id, err := pathInt64(r, "id")
	if err != nil || id <= 0 {
		http.Redirect(w, r, "/titles", http.StatusSeeOther)
		return
	}
	name := strings.TrimSpace(r.FormValue("name"))
	if name == "" {
		http.Redirect(w, r, "/titles", http.StatusSeeOther)
		return
	}
	err = s.store.UpdateTitle(r.Context(), id, name)
	if err != nil {
		s.renderTitles(r.Context(), w, err.Error())
		return
	}

	setToast(w, "Title updated.")
	s.renderTitles(r.Context(), w, "")
}

func (s *Server) handleTitleToggle(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/titles", http.StatusSeeOther)
		return
	}
	id, err := pathInt64(r, "id")
	if err != nil || id <= 0 {
		http.Redirect(w, r, "/titles", http.StatusSeeOther)
		return
	}
	active := r.FormValue("active") == "1"
	err = s.store.SetTitleActive(r.Context(), id, active)
	if err != nil {
		s.renderTitles(r.Context(), w, err.Error())
		return
	}

	msg := "Title deactivated."
	if active {
		msg = "Title activated."
	}
	setToast(w, msg)
	s.renderTitles(r.Context(), w, "")
}

func (s *Server) handleTitleDelete(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Redirect(w, r, "/titles", http.StatusSeeOther)
		return
	}
	id, err := pathInt64(r, "id")
	if err != nil || id <= 0 {
		http.Redirect(w, r, "/titles", http.StatusSeeOther)
		return
	}
	err = s.store.SetTitleActive(r.Context(), id, false)
	if err != nil {
		s.renderTitles(r.Context(), w, "Unable to delete title (it may be referenced by an existing game).")
		return
	}

	setToast(w, "Title deactivated.")
	s.renderTitles(r.Context(), w, "")
}

func (s *Server) renderHomeWithError(ctx context.Context, w http.ResponseWriter, msg string, form HomeForm) {
	vm, err := s.newHomeVM(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	vm.FormError = msg
	vm.Form = form
	if err := s.r.HTML(w, "main", "home", vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
