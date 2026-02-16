package game

import (
	"fmt"
	"math"
	"sort"
	"time"
)

type PlayerYearStats struct {
	PlayerID    int64
	Attendance  int
	GamesPlayed int
	Wins        int
	WinRate     float64
	Qualified   bool
}

type YearStandings struct {
	Year     int
	ScopeKey string // "2026"

	Stats []PlayerYearStats

	Qualifiers []int64 // player IDs
	TopIDs     []int64 // tied leaders
	WinnerID   *int64

	TieUnresolved bool
}

func YearScopeKey(year int) string { return fmt.Sprintf("%d", year) }

// ComputeYearStandings computes the standings for a given year.
func ComputeYearStandings(
	games []Game,
	year int,
	getTB func(scope, scopeKey string) (Tiebreaker, bool, error),
) YearStandings {
	ys := YearStandings{
		Year:     year,
		ScopeKey: YearScopeKey(year),
	}

	// Attendance: a player attended a day if they were a participant on that day.
	attendedDays := map[int64]map[string]bool{} // playerID -> dateKey -> true
	playedCount := map[int64]int{}
	winsCount := map[int64]int{}

	for _, g := range games {
		dateKey := g.PlayedAt.In(time.Local).Format("2006-01-02")
		for _, pid := range g.ParticipantIDs {
			if attendedDays[pid] == nil {
				attendedDays[pid] = map[string]bool{}
			}
			attendedDays[pid][dateKey] = true
			playedCount[pid]++
		}
		for _, wid := range g.WinnerIDs {
			winsCount[wid]++
		}
	}

	// Build stats
	keys := unionKeys(attendedDays, playedCount, winsCount)
	for pid := range keys {
		att := len(attendedDays[pid])
		gp := playedCount[pid]
		wins := winsCount[pid]
		wr := 0.0
		if gp > 0 {
			wr = float64(wins) / float64(gp)
		}
		ys.Stats = append(ys.Stats, PlayerYearStats{
			PlayerID:    pid,
			Attendance:  att,
			GamesPlayed: gp,
			Wins:        wins,
			WinRate:     math.Round(wr*1000) / 10, // one decimal percent
		})
	}

	// Sort by wins desc, then win rate desc, then id asc
	sort.Slice(ys.Stats, func(i, j int) bool {
		if ys.Stats[i].Wins != ys.Stats[j].Wins {
			return ys.Stats[i].Wins > ys.Stats[j].Wins
		}
		if ys.Stats[i].WinRate != ys.Stats[j].WinRate {
			return ys.Stats[i].WinRate > ys.Stats[j].WinRate
		}
		return ys.Stats[i].PlayerID < ys.Stats[j].PlayerID
	})

	// Qualify: attendance >= 10 (same rule as before)
	for i := range ys.Stats {
		if ys.Stats[i].Attendance >= 10 {
			ys.Stats[i].Qualified = true
			ys.Qualifiers = append(ys.Qualifiers, ys.Stats[i].PlayerID)
		}
	}

	// Determine top among qualified (or among all if no qualifiers)
	candidates := ys.Qualifiers
	if len(candidates) == 0 {
		for _, st := range ys.Stats {
			candidates = append(candidates, st.PlayerID)
		}
	}

	maxWins := -1
	for _, pid := range candidates {
		w := winsCount[pid]
		if w > maxWins {
			maxWins = w
		}
	}
	for _, pid := range candidates {
		if winsCount[pid] == maxWins && maxWins >= 0 {
			ys.TopIDs = append(ys.TopIDs, pid)
		}
	}
	sort.Slice(ys.TopIDs, func(i, j int) bool { return ys.TopIDs[i] < ys.TopIDs[j] })

	if len(ys.TopIDs) == 1 {
		ys.WinnerID = &ys.TopIDs[0]
		return ys
	}

	if len(ys.TopIDs) > 1 {
		if getTB != nil {
			tb, ok, err := getTB("yearly", ys.ScopeKey)
			if err != nil {
				panic(err)
			}
			if ok {
				if containsID(ys.TopIDs, tb.WinnerID) {
					wid := tb.WinnerID
					ys.WinnerID = &wid
					return ys
				}
			}
		}
		ys.TieUnresolved = true
	}

	return ys
}

// unionKeys returns a set of keys across three maps.
func unionKeys(m1 map[int64]map[string]bool, m2 map[int64]int, m3 map[int64]int) map[int64]struct{} {
	out := map[int64]struct{}{}
	for k := range m1 {
		out[k] = struct{}{}
	}
	for k := range m2 {
		out[k] = struct{}{}
	}
	for k := range m3 {
		out[k] = struct{}{}
	}
	return out
}
