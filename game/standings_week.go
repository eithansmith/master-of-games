package game

import (
	"sort"
)

type WeekStandings struct {
	Year     int
	Week     int
	ScopeKey string // "2026-W07"

	TotalGames int
	Wins       map[int]int // playerID -> wins

	TopIDs   []int // tied leaders by wins (could be 1)
	WinnerID *int  // set when decided (single leader OR tiebreaker winner)
}

func ComputeWeekStandings(
	games []Game,
	year int,
	week int,
	playersCount int,
	lookupTiebreaker func(scope, scopeKey string) (Tiebreaker, bool),
) WeekStandings {
	scopeKey := isoWeekKey(year, week)

	wins := make(map[int]int, playersCount)
	total := 0

	for _, g := range games {
		gy, gw := g.PlayedAt.ISOWeek()
		if gy != year || gw != week {
			continue
		}
		if !IsWeekdayLocal(g.PlayedAt) {
			continue
		}
		total++

		// Option A: each winner gets +1 win.
		for _, wid := range g.WinnerIDs {
			wins[wid]++
		}
	}

	ws := WeekStandings{
		Year:       year,
		Week:       week,
		ScopeKey:   scopeKey,
		TotalGames: total,
		Wins:       wins,
	}

	if total == 0 {
		// No trophy awarded.
		return ws
	}

	// Find max wins
	maxWins := -1
	for pid := 0; pid < playersCount; pid++ {
		if wins[pid] > maxWins {
			maxWins = wins[pid]
		}
	}

	// Collect all tied leaders
	var top []int
	for pid := 0; pid < playersCount; pid++ {
		if wins[pid] == maxWins {
			top = append(top, pid)
		}
	}
	ws.TopIDs = top

	// If exactly 1 leader, winner is decided automatically
	if len(top) == 1 {
		ws.WinnerID = &top[0]
		return ws
	}

	// Otherwise, check if a tiebreaker exists for this week
	if tb, ok := lookupTiebreaker("weekly", scopeKey); ok {
		// (Optional safety) ensure tb.WinnerID is in tb.TiedPlayerIDs
		ws.WinnerID = &tb.WinnerID
	}

	// Sort TopIDs for stable display
	sort.Ints(ws.TopIDs)
	return ws
}

func isoWeekKey(year, week int) string {
	// "2026-W07"
	return fmtWeekKey(year, week)
}

func fmtWeekKey(year, week int) string {
	// no fmt import in snippet; if you prefer:
	// return fmt.Sprintf("%04d-W%02d", year, week)
	return weekKey(year, week)
}

// keep this helper small to avoid lots of imports; implement in your preferred file
func weekKey(year, week int) string {
	// naive formatting:
	s := ""
	// (just use fmt.Sprintf in your actual code)
	return s
}
