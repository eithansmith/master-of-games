package game

import (
	"testing"
	"time"
)

func raceGame(year int, isoWeek int, winnerID int64) Game {
	// Build a Monday of the given ISO week in the given year.
	t := time.Date(year, 1, 4, 12, 0, 0, 0, time.UTC) // Jan 4 is always in week 1
	y, w := t.ISOWeek()
	for y < year || w < isoWeek {
		t = t.Add(7 * 24 * time.Hour)
		y, w = t.ISOWeek()
	}
	return Game{
		PlayedAt:       t,
		ParticipantIDs: []int64{winnerID},
		WinnerIDs:      []int64{winnerID},
		IsActive:       true,
	}
}

func playerList(ids ...int64) []Player {
	out := make([]Player, len(ids))
	for i, id := range ids {
		out[i] = Player{ID: id, Name: nameFor(id), IsActive: true}
	}
	return out
}

func nameFor(id int64) string {
	names := map[int64]string{1: "Alice", 2: "Bob", 3: "Carol", 4: "Dave", 5: "Eve", 6: "Frank"}
	if n, ok := names[id]; ok {
		return n
	}
	return "Unknown"
}

func TestComputeYearRace_NoGames(t *testing.T) {
	race := ComputeYearRace(nil, 2026, RaceMetricWins, 5, playerList(1, 2))
	if len(race.Weeks) != 0 {
		t.Errorf("Weeks = %v, want empty", race.Weeks)
	}
	if len(race.Series) != 0 {
		t.Errorf("Series = %v, want empty", race.Series)
	}
}

func TestComputeYearRace_CumulativeWins(t *testing.T) {
	games := []Game{
		raceGame(2026, 1, 1), // Alice wins week 1
		raceGame(2026, 1, 1), // Alice wins week 1 again
		raceGame(2026, 2, 1), // Alice wins week 2
		raceGame(2026, 2, 2), // Bob wins week 2
	}
	players := playerList(1, 2)
	race := ComputeYearRace(games, 2026, RaceMetricWins, 5, players)

	if len(race.Weeks) != 2 {
		t.Fatalf("Weeks = %v, want [1 2]", race.Weeks)
	}

	seriesByName := map[string]RaceSeries{}
	for _, s := range race.Series {
		seriesByName[s.Name] = s
	}

	alice := seriesByName["Alice"]
	if alice.Values[0] != 2 {
		t.Errorf("Alice week 1 cumulative = %.0f, want 2", alice.Values[0])
	}
	if alice.Values[1] != 3 {
		t.Errorf("Alice week 2 cumulative = %.0f, want 3", alice.Values[1])
	}

	bob := seriesByName["Bob"]
	if bob.Values[0] != 0 {
		t.Errorf("Bob week 1 cumulative = %.0f, want 0", bob.Values[0])
	}
	if bob.Values[1] != 1 {
		t.Errorf("Bob week 2 cumulative = %.0f, want 1", bob.Values[1])
	}
}

func TestComputeYearRace_TopNFiltering(t *testing.T) {
	// 6 players, top 3 requested. Players ranked by final wins: 1>2>3>4=5=6.
	games := []Game{
		raceGame(2026, 1, 1), raceGame(2026, 1, 1), raceGame(2026, 1, 1), // 3 wins
		raceGame(2026, 1, 2), raceGame(2026, 1, 2), // 2 wins
		raceGame(2026, 1, 3), // 1 win
		// players 4, 5, 6 have 0 wins
	}
	players := playerList(1, 2, 3, 4, 5, 6)
	race := ComputeYearRace(games, 2026, RaceMetricWins, 3, players)

	if len(race.Series) != 3 {
		t.Fatalf("Series len = %d, want 3", len(race.Series))
	}
	names := map[string]bool{}
	for _, s := range race.Series {
		names[s.Name] = true
	}
	for _, expected := range []string{"Alice", "Bob", "Carol"} {
		if !names[expected] {
			t.Errorf("expected %s in top 3 series, got %v", expected, race.Series)
		}
	}
}

func TestComputeYearRace_GamesFromOtherYearIgnored(t *testing.T) {
	games := []Game{
		raceGame(2025, 1, 1), // wrong year
		raceGame(2026, 1, 2), // correct year
	}
	players := playerList(1, 2)
	race := ComputeYearRace(games, 2026, RaceMetricWins, 5, players)

	seriesByName := map[string]RaceSeries{}
	for _, s := range race.Series {
		seriesByName[s.Name] = s
	}
	if v := seriesByName["Alice"].Values[0]; v != 0 {
		t.Errorf("Alice (2025 winner) should have 0 wins in 2026, got %.0f", v)
	}
	if v := seriesByName["Bob"].Values[0]; v != 1 {
		t.Errorf("Bob should have 1 win in 2026, got %.0f", v)
	}
}

func TestComputeYearRace_InactivePlayersExcluded(t *testing.T) {
	games := []Game{raceGame(2026, 1, 1)}
	players := []Player{
		{ID: 1, Name: "Alice", IsActive: true},
		{ID: 2, Name: "Bob", IsActive: false}, // inactive — should be excluded
	}
	race := ComputeYearRace(games, 2026, RaceMetricWins, 5, players)

	for _, s := range race.Series {
		if s.Name == "Bob" {
			t.Error("inactive player Bob should not appear in the race")
		}
	}
}

func TestComputeYearRace_WeeksAreSorted(t *testing.T) {
	games := []Game{
		raceGame(2026, 3, 1),
		raceGame(2026, 1, 2),
		raceGame(2026, 2, 1),
	}
	race := ComputeYearRace(games, 2026, RaceMetricWins, 5, playerList(1, 2))

	for i := 1; i < len(race.Weeks); i++ {
		if race.Weeks[i] <= race.Weeks[i-1] {
			t.Errorf("Weeks not sorted: %v", race.Weeks)
		}
	}
}
