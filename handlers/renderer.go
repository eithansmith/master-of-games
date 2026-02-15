package handlers

import (
	"errors"
	"html/template"
	"net/http"
)

type Renderer struct {
	home    *template.Template
	week    *template.Template
	year    *template.Template
	players *template.Template
	titles  *template.Template
}

// RendererConfig centralizes template paths.
type RendererConfig struct {
	Base    string
	Home    string
	Week    string
	Year    string
	Players string
	Titles  string
}

func NewRenderer(cfg RendererConfig) *Renderer {
	funcs := template.FuncMap{
		"derefInt": func(p *int) int {
			if p == nil {
				return 0
			}
			return *p
		},
		"derefInt64": func(p *int64) int64 {
			if p == nil {
				return 0
			}
			return *p
		},
	}

	parse := func(files ...string) *template.Template {
		t := template.New("").Funcs(funcs)
		return template.Must(t.ParseFiles(files...))
	}

	return &Renderer{
		home:    parse(cfg.Base, cfg.Home),
		week:    parse(cfg.Base, cfg.Week),
		year:    parse(cfg.Base, cfg.Year),
		players: parse(cfg.Base, cfg.Players),
		titles:  parse(cfg.Base, cfg.Titles),
	}
}

func (r *Renderer) HTML(w http.ResponseWriter, layout, name string, data any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	switch name {
	case "home":
		return r.home.ExecuteTemplate(w, layout, data)
	case "week":
		return r.week.ExecuteTemplate(w, layout, data)
	case "year":
		return r.year.ExecuteTemplate(w, layout, data)
	case "players":
		return r.players.ExecuteTemplate(w, layout, data)
	case "titles":
		return r.titles.ExecuteTemplate(w, layout, data)
	default:
		return errors.New("unknown template: " + name)
	}
}
