package handlers

import "github.com/eithansmith/master-of-games/game"

// Store is the dependency boundary for handlers.
// Anything (MemoryStore, PostgresStore, etc.) that implements this can back the app.
type Store interface {
	AddGame(g game.Game) game.Game
	DeleteGame(id int64) bool
	RecentGames(limit int) []game.Game

	GetTiebreaker(scope, scopeKey string) (game.Tiebreaker, bool)
	SetTiebreaker(tb game.Tiebreaker)
}
