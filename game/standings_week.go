package game

import (
	"fmt"
	"sort"
)

type WeekStandings struct {
	Year     int
	Week     int
	ScopeKey string // "2026-W07"

	TotalGames int
	Wins       map[int64]int // playerID -> wins

	TopIDs        []int64 // tied leaders by wins (could be 1)
	WinnerID      *int64
	TieUnresolved bool
}

func WeekScopeKey(year, week int) string {
	return fmt.Sprintf("%04d-W%02d", year, week)
}

// ComputeWeekStandings computes the standings for a given week.
func ComputeWeekStandings(
	games []Game,
	year, week int,
	getTB func(scope, scopeKey string) (Tiebreaker, bool, error),
) WeekStandings {
	ws := WeekStandings{
		Year:     year,
		Week:     week,
		ScopeKey: WeekScopeKey(year, week),

		TotalGames: len(games),
		Wins:       map[int64]int{},
	}

	for _, g := range games {
		for _, wid := range g.WinnerIDs {
			ws.Wins[wid]++
		}
	}

	maxWins := 0
	for _, w := range ws.Wins {
		if w > maxWins {
			maxWins = w
		}
	}
	for pid, w := range ws.Wins {
		if w == maxWins && maxWins > 0 {
			ws.TopIDs = append(ws.TopIDs, pid)
		}
	}
	sort.Slice(ws.TopIDs, func(i, j int) bool { return ws.TopIDs[i] < ws.TopIDs[j] })

	if len(ws.TopIDs) == 1 {
		ws.WinnerID = &ws.TopIDs[0]
		return ws
	}

	if len(ws.TopIDs) > 1 {
		// If we have a stored tiebreaker, apply it.
		if getTB != nil {
			tb, ok, err := getTB("weekly", ws.ScopeKey)
			if err != nil {
				panic(err)
			}
			if ok {
				if containsID(ws.TopIDs, tb.WinnerID) {
					wid := tb.WinnerID
					ws.WinnerID = &wid
					return ws
				}
			}
		}
		ws.TieUnresolved = true
	}

	return ws
}

func containsID(xs []int64, v int64) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}
