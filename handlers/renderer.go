package handlers

import (
	"html/template"
	"net/http"
)

type Renderer struct {
	home *template.Template
	week *template.Template
	year *template.Template
}

// RendererConfig centralizes template paths.
type RendererConfig struct {
	Base string
	Home string
	Week string
	Year string
}

func NewRenderer(cfg RendererConfig) *Renderer {
	funcs := template.FuncMap{
		"derefInt": func(p *int) int {
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
		home: parse(cfg.Base, cfg.Home),
		week: parse(cfg.Base, cfg.Week),
		year: parse(cfg.Base, cfg.Year),
	}
}

// HTML renders a named template from the selected set.
// set: "home" | "week" | "year"
// name: template name inside that set ("home", "week", "year")
func (r *Renderer) HTML(w http.ResponseWriter, set, name string, vm any) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	var t *template.Template
	switch set {
	case "home":
		t = r.home
	case "week":
		t = r.week
	case "year":
		t = r.year
	default:
		http.Error(w, "unknown template set", http.StatusInternalServerError)
		return nil
	}

	if err := t.ExecuteTemplate(w, name, vm); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return err
	}
	return nil
}
