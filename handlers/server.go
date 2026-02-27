package handlers

import (
	"net/http"
)

// Meta holds build/runtime metadata you want available in templates.
type Meta struct {
	Version   string
	BuildTime string
	StartTime string
}

// Server owns HTTP handlers + template rendering for the app.
type Server struct {
	r     *Renderer
	store Store
	db    Pinger
	meta  Meta
}

// New constructs a Server with default template paths.
func New(store Store, db Pinger, meta Meta) *Server {
	r := NewRenderer(RendererConfig{
		Base:          "web/templates/base.go.html",
		Home:          "web/templates/home.go.html",
		Week:          "web/templates/week.go.html",
		Year:          "web/templates/year.go.html",
		YearRace:      "web/templates/year_race.go.html",
		YearRaceChart: "web/templates/year_race_chart.go.html",
		Players:       "web/templates/players.go.html",
		Titles:        "web/templates/titles.go.html",
	})

	return &Server{
		r:     r,
		store: store,
		db:    db,
		meta:  meta,
	}
}

// RegisterRoutes attaches all application routes to the provided mux.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	// Home
	mux.HandleFunc("GET /", s.handleHome)

	// Games
	mux.HandleFunc("POST /games", s.handleAddGame)
	// Toggle/retire a game (uses path params; HTMX posts here)
	mux.HandleFunc("POST /games/{id}/toggle", s.handleGameToggle)
	// Optional: hard-delete/retire endpoint if you want a distinct button later
	mux.HandleFunc("POST /games/{id}/delete", s.handleDeleteGame)

	// Weeks
	mux.HandleFunc("GET /weeks/current", s.handleWeekCurrent)
	mux.HandleFunc("GET /weeks/{year}/{week}", s.handleWeek)
	mux.HandleFunc("POST /weeks/{year}/{week}/tiebreak", s.handleWeekTiebreak)

	// Years
	mux.HandleFunc("GET /years/{year}", s.handleYear)
	mux.HandleFunc("POST /years/{year}/tiebreak", s.handleYearTiebreak)

	// Race charts
	mux.HandleFunc("GET /years/{year}/race", s.handleYearRace)
	mux.HandleFunc("GET /years/{year}/race/chart", s.handleYearRaceChart)

	// Admin-ish lists (simple CRUD)
	mux.HandleFunc("GET /players", s.handlePlayers)
	mux.HandleFunc("POST /players", s.handlePlayersPost)
	mux.HandleFunc("POST /players/{id}/update", s.handlePlayerUpdate)
	mux.HandleFunc("POST /players/{id}/toggle", s.handlePlayerToggle)
	mux.HandleFunc("POST /players/{id}/delete", s.handlePlayerDelete)

	mux.HandleFunc("GET /titles", s.handleTitles)
	mux.HandleFunc("POST /titles", s.handleTitlesPost)
	mux.HandleFunc("POST /titles/{id}/update", s.handleTitleUpdate)
	mux.HandleFunc("POST /titles/{id}/toggle", s.handleTitleToggle)
	mux.HandleFunc("POST /titles/{id}/delete", s.handleTitleDelete)

	// Health
	mux.HandleFunc("GET /healthz", s.handleHealthz)
	mux.HandleFunc("GET /readyz", s.handleReadyz)
}
