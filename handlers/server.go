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
		Base:    "web/templates/base.go.html",
		Home:    "web/templates/home.go.html",
		Week:    "web/templates/week.go.html",
		Year:    "web/templates/year.go.html",
		Players: "web/templates/players.go.html",
		Titles:  "web/templates/titles.go.html",
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
	mux.HandleFunc("/", s.handleHome)
	mux.HandleFunc("/games", s.handleAddGame)           // POST
	mux.HandleFunc("/games/delete", s.handleDeleteGame) // POST (legacy alias)
	mux.HandleFunc("/games/toggle", s.handleGameToggle) // POST
	mux.HandleFunc("/weeks/current", s.handleWeekCurrent)
	mux.HandleFunc("/weeks/", s.handleWeek) // GET + POST tiebreak
	mux.HandleFunc("/years/", s.handleYear) // GET + POST tiebreak

	// Admin-ish lists (simple CRUD)
	mux.HandleFunc("/players", s.handlePlayers)             // GET + POST
	mux.HandleFunc("/players/update", s.handlePlayerUpdate) // POST
	mux.HandleFunc("/players/delete", s.handlePlayerDelete) // POST (legacy alias)
	mux.HandleFunc("/players/toggle", s.handlePlayerToggle) // POST

	mux.HandleFunc("/titles", s.handleTitles)             // GET + POST
	mux.HandleFunc("/titles/update", s.handleTitleUpdate) // POST
	mux.HandleFunc("/titles/delete", s.handleTitleDelete) // POST (legacy alias)
	mux.HandleFunc("/titles/toggle", s.handleTitleToggle) // POST

	mux.HandleFunc("/healthz", s.handleHealthz)
	mux.HandleFunc("/readyz", s.handleReadyz)
}
