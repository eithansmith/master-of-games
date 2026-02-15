package handlers

import (
	"strconv"
	"time"
)

// parseInt64Slice converts string values (typically checkbox IDs) into an int64 slice.
func parseInt64Slice(vals []string) []int64 {
	out := make([]int64, 0, len(vals))
	for _, v := range vals {
		i, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			out = append(out, i)
		}
	}
	return out
}

// parseInt64Map converts string values (typically checkbox IDs) into a lookup map.
func parseInt64Map(vals []string) map[int64]bool {
	m := make(map[int64]bool, len(vals))
	for _, v := range vals {
		i, err := strconv.ParseInt(v, 10, 64)
		if err == nil {
			m[i] = true
		}
	}
	return m
}

// isSubset returns true iff sub is a subset of set.
func isSubset(sub, set []int64) bool {
	m := map[int64]bool{}
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

// containsInt64 returns true iff v is in xs.
func containsInt64(xs []int64, v int64) bool {
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
	weeksInYear := isoWeeksInYear(year)
	if week < weeksInYear {
		return year, week + 1
	}
	return year + 1, 1
}
