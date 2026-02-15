package handlers

import (
	"context"

	"github.com/eithansmith/master-of-games/game"
)

// Store is the dependency boundary for handlers.
// Anything (MemoryStore, PostgresStore, etc.) that implements this can back the app.
type Store interface {
	AddGame(g game.Game) game.Game
	DeleteGame(id int64) bool
	RecentGames(limit int) []game.Game

	ListPlayers() []game.Player
	AddPlayer(name string) game.Player
	UpdatePlayer(id int64, name string) bool
	DeletePlayer(id int64) bool

	ListTitles() []game.Title
	AddTitle(name string) game.Title
	UpdateTitle(id int64, name string) bool
	DeleteTitle(id int64) bool

	GetTiebreaker(scope, scopeKey string) (game.Tiebreaker, bool)
	SetTiebreaker(tb game.Tiebreaker)
}

// Pinger is a simple interface for testing.
type Pinger interface {
	Ping(ctx context.Context) error
}
