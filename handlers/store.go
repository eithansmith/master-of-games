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
	AddGame(g game.Game) game.Game
	DeleteGame(id int64) bool
	SetGameActive(id int64, active bool) bool
	RecentGames(limit int) []game.Game
	GamesByWeek(year, week int) []game.Game
	GamesByYear(year int) []game.Game

	// players
	ListPlayers() []game.Player
	AddPlayer(name string) game.Player
	UpdatePlayer(id int64, name string) bool
	SetPlayerActive(id int64, active bool) bool
	DeletePlayer(id int64) bool

	// titles
	ListTitles() []game.Title
	AddTitle(name string) game.Title
	UpdateTitle(id int64, name string) bool
	SetTitleActive(id int64, active bool) bool
	DeleteTitle(id int64) bool

	// tiebreakers
	GetTiebreaker(scope, scopeKey string) (game.Tiebreaker, bool)
	SetTiebreaker(tb game.Tiebreaker)
}

// Pinger is a simple interface for testing.
type Pinger interface {
	Ping(ctx context.Context) error
}
