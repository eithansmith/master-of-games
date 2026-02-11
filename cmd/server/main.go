package main

import (
	"html/template"
	"log"
	"net/http"
	"os"
	"time"
)

var (
	version   = "dev"
	buildTime = "unknown"
)

type Server struct {
	tmpl *template.Template
}

func main() {
	addr := env("PORT", "8080")
	s := &Server{
		tmpl: template.Must(template.ParseFiles(
			"web/templates/base.go.html",
			"web/templates/home.go.html",
		)),
	}

	mux := http.NewServeMux()

	// Static files
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// Pages
	mux.HandleFunc("/", s.handleHome)

	// Basic health endpoint (Render-friendly)
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

	log.Printf("starting master-of-games version=%s buildTime=%s", version, buildTime)
	log.Fatal(srv.ListenAndServe())
}

func (s *Server) handleHome(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{
		"Title":     "Master of Games",
		"Version":   version,
		"BuildTime": buildTime,
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := s.tmpl.ExecuteTemplate(w, "base", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
