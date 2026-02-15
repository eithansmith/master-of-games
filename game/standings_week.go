package game

import (
	"fmt"
	"sort"
	"time"
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

// ComputeWeekStandings filters games to the requested ISO year+week, counts wins, and applies any stored tiebreaker.
func ComputeWeekStandings(
	games []Game,
	year, week int,
	getTB func(scope, scopeKey string) (Tiebreaker, bool),
) WeekStandings {
	ws := WeekStandings{
		Year:     year,
		Week:     week,
		ScopeKey: WeekScopeKey(year, week),
		Wins:     map[int64]int{},
	}

	for _, g := range games {
		y, w := g.PlayedAt.In(time.Local).ISOWeek()
		if y != year || w != week {
			continue
		}
		ws.TotalGames++
		for _, wid := range g.WinnerIDs {
			ws.Wins[wid]++
		}
	}

	// Determine leaders
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
			if tb, ok := getTB("weekly", ws.ScopeKey); ok {
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
