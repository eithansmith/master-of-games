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
//
// Spec implemented:
// - Attendance = unique days attended (participated).
// - Qualifiers = top 1/2 of attendees (by attendance).
// - Winner = highest win rate (wins/games played) among qualifiers.
// - Any tie for winner is resolved by chance (stored tiebreaker), else unresolved.
func ComputeYearStandings(
	games []Game,
	year int,
	getTB func(scope, scopeKey string) (Tiebreaker, bool, error),
) YearStandings {
	ys := YearStandings{
		Year:     year,
		ScopeKey: YearScopeKey(year),
	}

	attendedDays := map[int64]map[string]bool{} // playerID -> dateKey -> true
	playedCount := map[int64]int{}
	winsCount := map[int64]int{}

	// Only consider games in the requested year (in local time).
	for _, g := range games {
		local := g.PlayedAt.In(time.Local)
		if local.Year() != year {
			continue
		}

		dateKey := local.Format("2006-01-02")
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

	// Build stats for anyone who appeared in any map.
	keys := unionKeys(attendedDays, playedCount, winsCount)
	for pid := range keys {
		att := len(attendedDays[pid])
		gp := playedCount[pid]
		wins := winsCount[pid]

		wr := 0.0
		if gp > 0 {
			wr = float64(wins) / float64(gp) // 0..1
		}

		ys.Stats = append(ys.Stats, PlayerYearStats{
			PlayerID:    pid,
			Attendance:  att,
			GamesPlayed: gp,
			Wins:        wins,
			WinRate:     math.Round(wr*1000) / 10, // percent with 1 decimal (e.g. 66.7)
		})
	}

	// Sort standings for display: attendance desc, then win rate desc, then id asc.
	sort.Slice(ys.Stats, func(i, j int) bool {
		if ys.Stats[i].Attendance != ys.Stats[j].Attendance {
			return ys.Stats[i].Attendance > ys.Stats[j].Attendance
		}
		if ys.Stats[i].WinRate != ys.Stats[j].WinRate {
			return ys.Stats[i].WinRate > ys.Stats[j].WinRate
		}
		return ys.Stats[i].PlayerID < ys.Stats[j].PlayerID
	})

	// Qualifiers = top half by attendance.
	// If the cutoff is tied, include all tied at the cutoff attendance.
	if len(ys.Stats) == 0 {
		return ys
	}

	n := len(ys.Stats)
	half := (n + 1) / 2 // top half; for odd N, includes the larger half.

	cutAttendance := ys.Stats[half-1].Attendance
	for i := range ys.Stats {
		if ys.Stats[i].Attendance >= cutAttendance {
			ys.Stats[i].Qualified = true
			ys.Qualifiers = append(ys.Qualifiers, ys.Stats[i].PlayerID)
		} else {
			break // because sorted by attendance desc
		}
	}

	// Determine leader(s) by WIN RATE among qualifiers.
	bestRate := -1.0
	for _, pid := range ys.Qualifiers {
		gp := playedCount[pid]
		if gp == 0 {
			continue
		}
		rate := float64(winsCount[pid]) / float64(gp)
		if rate > bestRate {
			bestRate = rate
		}
	}

	// If nobody has any games played (weird but possible), there is no winner.
	if bestRate < 0 {
		return ys
	}

	for _, pid := range ys.Qualifiers {
		gp := playedCount[pid]
		if gp == 0 {
			continue
		}
		//rate := float64(winsCount[pid]) / float64(gp)

		// Use exact equality on the rational comparison by cross-multiplying to avoid float issues.
		// rate == bestRate  <=>  wins/gp == bestWins/bestGP
		// But we didn't store bestWins/bestGP. So we do a tolerant float compare OR recompute via cross-multiply by scanning again:
		// We'll do cross-multiply with bestRate represented as wins/gp by re-finding a "best representative".
	}

	// Re-find a representative best (wins, gp) pair, then use cross-multiply to find ties exactly.
	var bestWins, bestGP int
	for _, pid := range ys.Qualifiers {
		gp := playedCount[pid]
		if gp == 0 {
			continue
		}
		w := winsCount[pid]
		if float64(w)/float64(gp) == bestRate {
			bestWins, bestGP = w, gp
			break
		}
	}
	for _, pid := range ys.Qualifiers {
		gp := playedCount[pid]
		if gp == 0 {
			continue
		}
		w := winsCount[pid]
		if w*bestGP == bestWins*gp {
			ys.TopIDs = append(ys.TopIDs, pid)
		}
	}
	sort.Slice(ys.TopIDs, func(i, j int) bool { return ys.TopIDs[i] < ys.TopIDs[j] })

	if len(ys.TopIDs) == 1 {
		ys.WinnerID = &ys.TopIDs[0]
		return ys
	}

	// Tie: resolve via stored "game of chance" tiebreaker if present.
	if len(ys.TopIDs) > 1 {
		if getTB != nil {
			tb, ok, err := getTB("yearly", ys.ScopeKey)
			if err == nil && ok && containsID(ys.TopIDs, tb.WinnerID) {
				wid := tb.WinnerID
				ys.WinnerID = &wid
				return ys
			}
			// If err != nil, we *don't* panic; we just leave it unresolved.
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
