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
	meta  Meta
}

// New constructs a Server with default template paths.
func New(store Store, meta Meta) *Server {
	r := NewRenderer(RendererConfig{
		Base: "web/templates/base.go.html",
		Home: "web/templates/home.go.html",
		Week: "web/templates/week.go.html",
		Year: "web/templates/year.go.html",
	})

	return &Server{
		r:     r,
		store: store,
		meta:  meta,
	}
}

// RegisterRoutes attaches all application routes to the provided mux.
func (s *Server) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", s.handleHome)
	mux.HandleFunc("/games", s.handleAddGame)           // POST
	mux.HandleFunc("/games/delete", s.handleDeleteGame) // POST
	mux.HandleFunc("/weeks/current", s.handleWeekCurrent)
	mux.HandleFunc("/weeks/", s.handleWeek) // GET + POST tiebreak
	mux.HandleFunc("/years/", s.handleYear) // GET + POST tiebreak
	mux.HandleFunc("/healthz", healthz)
}
