package game

import (
	"testing"
	"time"
)

func noTB(_, _ string) (Tiebreaker, bool, error) { return Tiebreaker{}, false, nil }

func tbFor(scopeKey string, winnerID int64) func(string, string) (Tiebreaker, bool, error) {
	return func(scope, key string) (Tiebreaker, bool, error) {
		if key == scopeKey {
			return Tiebreaker{WinnerID: winnerID}, true, nil
		}
		return Tiebreaker{}, false, nil
	}
}

func makeGame(winners ...int64) Game {
	return Game{
		PlayedAt:       time.Date(2026, 1, 5, 12, 0, 0, 0, time.UTC), // Monday W01
		ParticipantIDs: winners,
		WinnerIDs:      winners,
		IsActive:       true,
	}
}

func TestComputeWeekStandings_NoGames(t *testing.T) {
	ws := ComputeWeekStandings(nil, 2026, 1, noTB)

	if ws.TotalGames != 0 {
		t.Errorf("TotalGames = %d, want 0", ws.TotalGames)
	}
	if ws.WinnerID != nil {
		t.Error("WinnerID should be nil with no games")
	}
	if ws.TieUnresolved {
		t.Error("TieUnresolved should be false with no games")
	}
}

func TestComputeWeekStandings_ClearWinner(t *testing.T) {
	games := []Game{
		makeGame(1),
		makeGame(1),
		makeGame(2),
	}
	ws := ComputeWeekStandings(games, 2026, 1, noTB)

	if ws.TotalGames != 3 {
		t.Errorf("TotalGames = %d, want 3", ws.TotalGames)
	}
	if ws.WinnerID == nil || *ws.WinnerID != 1 {
		t.Errorf("WinnerID = %v, want 1", ws.WinnerID)
	}
	if ws.TieUnresolved {
		t.Error("TieUnresolved should be false")
	}
	if ws.TotalWins != 2 {
		t.Errorf("TotalWins = %d, want 2", ws.TotalWins)
	}
}

func TestComputeWeekStandings_TieNoTiebreaker(t *testing.T) {
	games := []Game{makeGame(1), makeGame(2)}
	ws := ComputeWeekStandings(games, 2026, 1, noTB)

	if ws.WinnerID != nil {
		t.Errorf("WinnerID should be nil for unresolved tie, got %v", ws.WinnerID)
	}
	if !ws.TieUnresolved {
		t.Error("TieUnresolved should be true")
	}
	if len(ws.TopIDs) != 2 {
		t.Errorf("TopIDs len = %d, want 2", len(ws.TopIDs))
	}
}

func TestComputeWeekStandings_TieResolvedByTiebreaker(t *testing.T) {
	games := []Game{makeGame(1), makeGame(2)}
	scopeKey := WeekScopeKey(2026, 1)
	ws := ComputeWeekStandings(games, 2026, 1, tbFor(scopeKey, 2))

	if ws.WinnerID == nil || *ws.WinnerID != 2 {
		t.Errorf("WinnerID = %v, want 2", ws.WinnerID)
	}
	if ws.TieUnresolved {
		t.Error("TieUnresolved should be false when tiebreaker resolves it")
	}
}

func TestComputeWeekStandings_TiebreakerWrongPlayer(t *testing.T) {
	// Tiebreaker names player 99 who is not in TopIDs — should be ignored.
	games := []Game{makeGame(1), makeGame(2)}
	scopeKey := WeekScopeKey(2026, 1)
	ws := ComputeWeekStandings(games, 2026, 1, tbFor(scopeKey, 99))

	if ws.WinnerID != nil {
		t.Errorf("WinnerID should be nil, got %v", ws.WinnerID)
	}
	if !ws.TieUnresolved {
		t.Error("TieUnresolved should be true when tiebreaker player is invalid")
	}
}

func TestComputeWeekStandings_MultipleWinnersPerGame(t *testing.T) {
	// A co-op game where both players 1 and 2 win.
	games := []Game{
		{PlayedAt: time.Date(2026, 1, 5, 12, 0, 0, 0, time.UTC), WinnerIDs: []int64{1, 2}, IsActive: true},
		{PlayedAt: time.Date(2026, 1, 5, 12, 0, 0, 0, time.UTC), WinnerIDs: []int64{1}, IsActive: true},
	}
	ws := ComputeWeekStandings(games, 2026, 1, noTB)

	if ws.Wins[1] != 2 {
		t.Errorf("player 1 wins = %d, want 2", ws.Wins[1])
	}
	if ws.Wins[2] != 1 {
		t.Errorf("player 2 wins = %d, want 1", ws.Wins[2])
	}
	if ws.WinnerID == nil || *ws.WinnerID != 1 {
		t.Errorf("WinnerID = %v, want 1", ws.WinnerID)
	}
}

func TestComputeWeekStandings_ScopeKey(t *testing.T) {
	ws := ComputeWeekStandings(nil, 2026, 7, noTB)
	if ws.ScopeKey != "2026-W07" {
		t.Errorf("ScopeKey = %q, want %q", ws.ScopeKey, "2026-W07")
	}
}

func TestWeekScopeKey(t *testing.T) {
	cases := []struct {
		year, week int
		want       string
	}{
		{2026, 1, "2026-W01"},
		{2026, 52, "2026-W52"},
		{2000, 9, "2000-W09"},
	}
	for _, tc := range cases {
		got := WeekScopeKey(tc.year, tc.week)
		if got != tc.want {
			t.Errorf("WeekScopeKey(%d, %d) = %q, want %q", tc.year, tc.week, got, tc.want)
		}
	}
}
