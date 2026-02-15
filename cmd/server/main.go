package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/eithansmith/master-of-games/db"
	"github.com/eithansmith/master-of-games/game"
	"github.com/eithansmith/master-of-games/handlers"
)

var (
	version   = "dev"
	buildTime = ""
)

func main() {
	addr := env("PORT", "8080")

	meta := handlers.Meta{
		Version:   version,
		BuildTime: buildTime,
		StartTime: time.Now().UTC().Format(time.RFC3339),
	}

	pool, err := db.NewPool(context.Background())
	if err != nil {
		log.Fatal(err)
	}
	defer pool.Close()

	store := game.NewPostgresStore(pool)

	s := handlers.New(store, pool, meta)

	mux := http.NewServeMux()
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))
	s.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:              ":" + addr,
		Handler:           logging(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("starting master-of-games version=%s buildTime=%s startTime=%s", meta.Version, meta.BuildTime, meta.StartTime)
	log.Fatal(srv.ListenAndServe())
}
