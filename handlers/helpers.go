package handlers

import (
	"strconv"
	"time"

	"github.com/eithansmith/master-of-games/game"
)

// parseIntSlice converts string values (typically checkbox indices) into an int slice.
func parseIntSlice(vals []string) []int {
	out := make([]int, 0, len(vals))
	for _, v := range vals {
		i, err := strconv.Atoi(v)
		if err == nil {
			out = append(out, i)
		}
	}
	return out
}

// parseIntMap converts string values (typically checkbox indices) into a lookup map.
func parseIntMap(vals []string) map[int]bool {
	m := make(map[int]bool, len(vals))
	for _, v := range vals {
		i, err := strconv.Atoi(v)
		if err == nil {
			m[i] = true
		}
	}
	return m
}

// isSubset returns true iff sub is a subset of set.
func isSubset(sub, set []int) bool {
	m := map[int]bool{}
	for _, v := range set {
		m[v] = true
	}
	for _, v := range sub {
		if !m[v] {
			return false
		}
	}
	return true
}

// containsInt returns true iff v is in xs.
func containsInt(xs []int, v int) bool {
	for _, x := range xs {
		if x == v {
			return true
		}
	}
	return false
}

// isoWeeksInYear returns the number of ISO weeks in the given year.
func isoWeeksInYear(year int) int {
	// ISO week of Dec 28 is always the last ISO week of the year
	_, w := time.Date(year, 12, 28, 0, 0, 0, 0, time.UTC).ISOWeek()
	return w
}

// prevISOWeek returns the previous ISO week of the given year and week number.
func prevISOWeek(year, week int) (int, int) {
	if week > 1 {
		return year, week - 1
	}
	py := year - 1
	return py, isoWeeksInYear(py)
}

// nextISOWeek returns the next ISO week of the given year and week number.
func nextISOWeek(year, week int) (int, int) {
	last := isoWeeksInYear(year)
	if week < last {
		return year, week + 1
	}
	ny := year + 1
	return ny, 1
}

// yearsFromGames returns the min and max years covered by the given games.
func yearsFromGames(games []game.Game, fallbackYear int) (minY, maxY int) {
	minY, maxY = fallbackYear, fallbackYear
	if len(games) == 0 {
		return
	}
	minY, maxY = games[0].PlayedAt.Year(), games[0].PlayedAt.Year()
	for _, g := range games {
		y := g.PlayedAt.Year()
		if y < minY {
			minY = y
		}
		if y > maxY {
			maxY = y
		}
	}
	return
}
