package handlers

import (
	"strconv"
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
