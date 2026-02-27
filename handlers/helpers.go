package handlers

import (
	"fmt"
	"math"
	"net/http"
	"sort"
	"strconv"
	"time"
)

// pathInt reads an integer path parameter from Go's ServeMux patterns.
func pathInt(r *http.Request, key string) (int, bool) {
	v := r.PathValue(key)
	if v == "" {
		return 0, false
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, false
	}
	return n, true
}

// pathInt64 reads an int64 path parameter from Go's ServeMux patterns.
func pathInt64(r *http.Request, key string) (int64, error) {
	v := r.PathValue(key)
	if v == "" {
		return 0, strconv.ErrSyntax
	}
	return strconv.ParseInt(v, 10, 64)
}

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
	maximum := isoWeeksInYear(year)
	if week < maximum {
		return year, week + 1
	}
	return year + 1, 1
}

// Round up to a nice axis max (1,2,5 * 10^k style)
func niceCeil(x float64) float64 {
	if x <= 0 {
		return 1
	}
	exp := math.Floor(math.Log10(x))
	base := math.Pow(10, exp)
	f := x / base

	var nice float64
	switch {
	case f <= 1:
		nice = 1
	case f <= 2:
		nice = 2
	case f <= 5:
		nice = 5
	default:
		nice = 10
	}
	return nice * base
}

func uniqueSortedInts(in []int) []int {
	m := map[int]bool{}
	for _, v := range in {
		if v < 0 {
			continue
		}
		m[v] = true
	}
	out := make([]int, 0, len(m))
	for v := range m {
		out = append(out, v)
	}
	sort.Ints(out)
	return out
}

func seriesColor(i int) string {
	// golden-angle-ish spacing to avoid clumping
	hue := (i * 137) % 360
	// looks good on a dark background
	return fmt.Sprintf("hsl(%d 70%% 55%%)", hue)
}

func yTickStep(axisMax int) int {
	switch {
	case axisMax <= 10:
		return 1
	case axisMax <= 20:
		return 2
	case axisMax <= 50:
		return 5
	case axisMax <= 100:
		return 10
	default:
		return 20
	}
}
