package handlers

import (
	"html/template"
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
	homeTmpl *template.Template
	weekTmpl *template.Template
	yearTmpl *template.Template
	store    Store
	meta     Meta
}

// New constructs a Server with default template paths.
func New(store Store, meta Meta) *Server {
	return &Server{
		homeTmpl: mustParse("web/templates/base.go.html", "web/templates/home.go.html"),
		weekTmpl: mustParse("web/templates/base.go.html", "web/templates/week.go.html"),
		yearTmpl: mustParse("web/templates/base.go.html", "web/templates/year.go.html"),
		store:    store,
		meta:     meta,
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
