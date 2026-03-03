package game

import (
	"testing"
	"time"
)

// makeYearGame builds a game played on a specific date with given participants and winners.
func makeYearGame(date time.Time, participants, winners []int64) Game {
	return Game{
		PlayedAt:       date,
		ParticipantIDs: participants,
		WinnerIDs:      winners,
		IsActive:       true,
	}
}

func day(year int, month time.Month, d int) time.Time {
	return time.Date(year, month, d, 12, 0, 0, 0, time.UTC)
}

func TestComputeYearStandings_NoGames(t *testing.T) {
	ys := ComputeYearStandings(nil, 2026, nil)

	if len(ys.Stats) != 0 {
		t.Errorf("Stats len = %d, want 0", len(ys.Stats))
	}
	if ys.WinnerID != nil {
		t.Error("WinnerID should be nil with no games")
	}
}

func TestComputeYearStandings_ClearWinner(t *testing.T) {
	// Player 1: 3 days, 3 wins. Player 2: 1 day, 1 win. Player 1 qualifies and wins.
	games := []Game{
		makeYearGame(day(2026, 1, 5), []int64{1, 2}, []int64{1}),
		makeYearGame(day(2026, 1, 6), []int64{1}, []int64{1}),
		makeYearGame(day(2026, 1, 7), []int64{1}, []int64{1}),
	}
	ys := ComputeYearStandings(games, 2026, noTB)

	if ys.WinnerID == nil || *ys.WinnerID != 1 {
		t.Errorf("WinnerID = %v, want 1", ys.WinnerID)
	}
	if ys.TieUnresolved {
		t.Error("TieUnresolved should be false")
	}
}

func TestComputeYearStandings_QualificationCutoff(t *testing.T) {
	// 4 players: attendance 4, 3, 2, 1. Top half = 2 qualifiers (players 1 and 2).
	games := []Game{
		makeYearGame(day(2026, 1, 5), []int64{1, 2, 3, 4}, []int64{3}),
		makeYearGame(day(2026, 1, 6), []int64{1, 2, 3}, []int64{3}),
		makeYearGame(day(2026, 1, 7), []int64{1, 2}, []int64{2}),
		makeYearGame(day(2026, 1, 8), []int64{1}, []int64{1}),
	}
	ys := ComputeYearStandings(games, 2026, noTB)

	qualSet := map[int64]bool{}
	for _, pid := range ys.Qualifiers {
		qualSet[pid] = true
	}
	if !qualSet[1] || !qualSet[2] {
		t.Errorf("Qualifiers should be players 1 and 2, got %v", ys.Qualifiers)
	}
	if qualSet[3] || qualSet[4] {
		t.Errorf("Players 3 and 4 should not qualify, got qualifiers %v", ys.Qualifiers)
	}
}

func TestComputeYearStandings_AttendanceTiePullsAllIn(t *testing.T) {
	// 3 players all with attendance=2. Cutoff hits all three; all qualify.
	games := []Game{
		makeYearGame(day(2026, 1, 5), []int64{1, 2, 3}, []int64{1}),
		makeYearGame(day(2026, 1, 6), []int64{1, 2, 3}, []int64{2}),
	}
	ys := ComputeYearStandings(games, 2026, noTB)

	if len(ys.Qualifiers) != 3 {
		t.Errorf("all 3 players should qualify when attendance is tied, got %v", ys.Qualifiers)
	}
}

func TestComputeYearStandings_WinRateBeatsRawWins(t *testing.T) {
	// Player 1: attends 2 days, plays 3 games, wins 1 = 33% win rate.
	// Player 2: attends 2 days, plays 1 game, wins 1 = 100% win rate.
	// Both qualify (top half); player 2 should win despite fewer raw wins.
	games := []Game{
		makeYearGame(day(2026, 1, 5), []int64{1, 2}, []int64{1}), // both attend Jan 5; p1 wins
		makeYearGame(day(2026, 1, 6), []int64{1}, nil),           // p1 plays Jan 6, no winner
		makeYearGame(day(2026, 1, 6), []int64{1}, nil),           // p1 plays Jan 6, no winner
		makeYearGame(day(2026, 1, 6), []int64{2}, []int64{2}),    // p2 plays Jan 6, p2 wins
	}
	ys := ComputeYearStandings(games, 2026, noTB)

	if ys.WinnerID == nil || *ys.WinnerID != 2 {
		t.Errorf("WinnerID = %v, want 2 (better win rate)", ys.WinnerID)
	}
}

func TestComputeYearStandings_TieNoTiebreaker(t *testing.T) {
	// Players 1 and 2 both attend 2 days and each win 1 of 2 games.
	games := []Game{
		makeYearGame(day(2026, 1, 5), []int64{1, 2}, []int64{1}),
		makeYearGame(day(2026, 1, 6), []int64{1, 2}, []int64{2}),
	}
	ys := ComputeYearStandings(games, 2026, noTB)

	if ys.WinnerID != nil {
		t.Errorf("WinnerID should be nil for unresolved tie, got %v", ys.WinnerID)
	}
	if !ys.TieUnresolved {
		t.Error("TieUnresolved should be true")
	}
	if len(ys.TopIDs) != 2 {
		t.Errorf("TopIDs len = %d, want 2", len(ys.TopIDs))
	}
}

func TestComputeYearStandings_TieResolvedByTiebreaker(t *testing.T) {
	games := []Game{
		makeYearGame(day(2026, 1, 5), []int64{1, 2}, []int64{1}),
		makeYearGame(day(2026, 1, 6), []int64{1, 2}, []int64{2}),
	}
	scopeKey := YearScopeKey(2026)
	ys := ComputeYearStandings(games, 2026, tbFor(scopeKey, 1))

	if ys.WinnerID == nil || *ys.WinnerID != 1 {
		t.Errorf("WinnerID = %v, want 1", ys.WinnerID)
	}
	if ys.TieUnresolved {
		t.Error("TieUnresolved should be false when tiebreaker resolves it")
	}
}

func TestComputeYearStandings_CrossMultiplyTieDetection(t *testing.T) {
	// Player 1: 2/3 games. Player 2: 4/6 games. Same ratio — should tie.
	games := []Game{
		makeYearGame(day(2026, 1, 5), []int64{1}, []int64{1}),
		makeYearGame(day(2026, 1, 5), []int64{1}, []int64{1}),
		makeYearGame(day(2026, 1, 5), []int64{1}, []int64{}),
		makeYearGame(day(2026, 1, 6), []int64{2}, []int64{2}),
		makeYearGame(day(2026, 1, 6), []int64{2}, []int64{2}),
		makeYearGame(day(2026, 1, 6), []int64{2}, []int64{2}),
		makeYearGame(day(2026, 1, 6), []int64{2}, []int64{2}),
		makeYearGame(day(2026, 1, 6), []int64{2}, []int64{}),
		makeYearGame(day(2026, 1, 6), []int64{2}, []int64{}),
	}
	ys := ComputeYearStandings(games, 2026, noTB)

	if len(ys.TopIDs) != 2 {
		t.Errorf("TopIDs = %v, want both players tied (2/3 == 4/6)", ys.TopIDs)
	}
}

func TestComputeYearStandings_GamesFromOtherYearIgnored(t *testing.T) {
	games := []Game{
		makeYearGame(day(2025, 12, 31), []int64{1}, []int64{1}), // wrong year
		makeYearGame(day(2026, 1, 5), []int64{2}, []int64{2}),
	}
	ys := ComputeYearStandings(games, 2026, noTB)

	for _, s := range ys.Stats {
		if s.PlayerID == 1 {
			t.Error("player 1 (2025 game) should not appear in 2026 standings")
		}
	}
	if ys.WinnerID == nil || *ys.WinnerID != 2 {
		t.Errorf("WinnerID = %v, want 2", ys.WinnerID)
	}
}

func TestYearScopeKey(t *testing.T) {
	cases := []struct {
		year int
		want string
	}{
		{2026, "2026"},
		{2000, "2000"},
	}
	for _, tc := range cases {
		got := YearScopeKey(tc.year)
		if got != tc.want {
			t.Errorf("YearScopeKey(%d) = %q, want %q", tc.year, got, tc.want)
		}
	}
}
