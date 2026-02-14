package game

import (
	"fmt"
	"math"
	"sort"
)

type PlayerYearStats struct {
	PlayerID    int
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

	Qualifiers []int // player IDs
	TopIDs     []int // tied leaders by win_rate among qualifiers
	WinnerID   *int
}

func ComputeYearStandings(
	games []Game,
	year int,
	playersCount int,
	lookupTiebreaker func(scope, scopeKey string) (Tiebreaker, bool),
) YearStandings {
	//scopeKey := itoaYear(year)

	// attendanceDays[pid] = set of dates (yyyy-mm-dd) they participated
	attendanceDays := make([]map[string]bool, playersCount)
	for i := range attendanceDays {
		attendanceDays[i] = map[string]bool{}
	}

	stats := make([]PlayerYearStats, playersCount)
	for pid := 0; pid < playersCount; pid++ {
		stats[pid] = PlayerYearStats{PlayerID: pid}
	}

	for _, g := range games {
		if g.PlayedAt.Year() != year {
			continue
		}
		if !IsWeekdayLocal(g.PlayedAt) {
			continue
		}

		dayKey := g.PlayedAt.Format("2006-01-02") // local date; fine for now

		// Games played + attendance
		for _, pid := range g.ParticipantIDs {
			stats[pid].GamesPlayed++
			attendanceDays[pid][dayKey] = true
		}

		// Wins
		for _, wid := range g.WinnerIDs {
			stats[wid].Wins++
		}
	}

	// Attendance counts
	for pid := 0; pid < playersCount; pid++ {
		stats[pid].Attendance = len(attendanceDays[pid])
		if stats[pid].GamesPlayed > 0 {
			stats[pid].WinRate = float64(stats[pid].Wins) / float64(stats[pid].GamesPlayed)
		}
	}

	// Determine qualifiers: top half by attendance (ceil(N/2))
	type arow struct{ pid, att int }
	rows := make([]arow, 0, playersCount)
	for pid := 0; pid < playersCount; pid++ {
		rows = append(rows, arow{pid: pid, att: stats[pid].Attendance})
	}
	sort.Slice(rows, func(i, j int) bool {
		if rows[i].att != rows[j].att {
			return rows[i].att > rows[j].att
		}
		return rows[i].pid < rows[j].pid
	})

	qCount := int(math.Ceil(float64(playersCount) / 2.0))
	if qCount < 1 {
		qCount = 1
	}
	qualSet := map[int]bool{}
	for i := 0; i < qCount && i < len(rows); i++ {
		qualSet[rows[i].pid] = true
	}
	var qualifiers []int
	for pid := 0; pid < playersCount; pid++ {
		stats[pid].Qualified = qualSet[pid]
		if stats[pid].Qualified {
			qualifiers = append(qualifiers, pid)
		}
	}

	ys := YearStandings{
		Year:       year,
		ScopeKey:   itoaYear(year),
		Stats:      stats,
		Qualifiers: qualifiers,
	}

	// If nobody qualified with any games played, there is no winner.
	// (Rare, but safe)
	best := -1.0
	for _, pid := range qualifiers {
		// You might decide to exclude games_played == 0 from contention
		// For now, keep them; winRate will be 0.
		if stats[pid].WinRate > best {
			best = stats[pid].WinRate
		}
	}

	var top []int
	for _, pid := range qualifiers {
		if stats[pid].WinRate == best {
			top = append(top, pid)
		}
	}
	ys.TopIDs = top

	if len(top) == 1 {
		ys.WinnerID = &top[0]
		return ys
	}

	// Tie: check tiebreaker
	if tb, ok := lookupTiebreaker("yearly", itoaYear(year)); ok {
		ys.WinnerID = &tb.WinnerID
	}

	return ys
}

func itoaYear(year int) string {
	// use strconv.Itoa in real code
	return fmt.Sprintf("%d", year)
}
