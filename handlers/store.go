package handlers

import (
	"context"

	"github.com/eithansmith/master-of-games/game"
)

// Store is the dependency boundary for handlers.
// Anything (MemoryStore, PostgresStore, etc.) that implements this can back the app.
//
//goland:noinspection GoCommentStart
type Store interface {
	// games
	AddGame(g game.Game) (game.Game, error)
	DeleteGame(id int64) error
	SetGameActive(id int64, active bool) error
	RecentGames(limit int) ([]game.Game, error)

	GetWeek(year, week int) ([]game.Game, error)
	GetYear(year int) ([]game.Game, error)

	// players
	ListPlayers() ([]game.Player, error)
	AddPlayer(name string) (game.Player, error)
	UpdatePlayer(id int64, name string) error
	SetPlayerActive(id int64, active bool) error
	DeletePlayer(id int64) error

	// titles
	ListTitles() ([]game.Title, error)
	AddTitle(name string) (game.Title, error)
	UpdateTitle(id int64, name string) error
	SetTitleActive(id int64, active bool) error
	DeleteTitle(id int64) error

	// tiebreakers
	GetTiebreaker(scope, scopeKey string) (game.Tiebreaker, bool, error)
	SetTiebreaker(tb game.Tiebreaker) error
}

// Pinger is a simple interface for testing.
type Pinger interface {
	Ping(ctx context.Context) error
}
