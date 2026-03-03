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
	AddGame(ctx context.Context, g game.Game) (game.Game, error)
	DeleteGame(ctx context.Context, id int64) error
	SetGameActive(ctx context.Context, id int64, active bool) error
	RecentGames(ctx context.Context, limit int) ([]game.Game, error)

	GetWeek(ctx context.Context, year, week int) ([]game.Game, error)
	GetYear(ctx context.Context, year int) ([]game.Game, error)

	// players
	ListPlayers(ctx context.Context) ([]game.Player, error)
	AddPlayer(ctx context.Context, name string) (game.Player, error)
	UpdatePlayer(ctx context.Context, id int64, name string) error
	SetPlayerActive(ctx context.Context, id int64, active bool) error
	DeletePlayer(ctx context.Context, id int64) error

	// titles
	ListTitles(ctx context.Context) ([]game.Title, error)
	AddTitle(ctx context.Context, name string) (game.Title, error)
	UpdateTitle(ctx context.Context, id int64, name string) error
	SetTitleActive(ctx context.Context, id int64, active bool) error
	DeleteTitle(ctx context.Context, id int64) error

	// tiebreakers
	GetTiebreaker(ctx context.Context, scope, scopeKey string) (game.Tiebreaker, bool, error)
	SetTiebreaker(ctx context.Context, tb game.Tiebreaker) error
}

// Pinger is a simple interface for testing.
type Pinger interface {
	Ping(ctx context.Context) error
}
