package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

// ============================
// parseInt64Slice
// ============================

func TestParseInt64Slice_Valid(t *testing.T) {
	got := parseInt64Slice([]string{"1", "2", "3"})
	if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Errorf("got %v, want [1 2 3]", got)
	}
}

func TestParseInt64Slice_SkipInvalid(t *testing.T) {
	got := parseInt64Slice([]string{"1", "abc", "3"})
	if len(got) != 2 || got[0] != 1 || got[1] != 3 {
		t.Errorf("got %v, want [1 3]", got)
	}
}

func TestParseInt64Slice_Empty(t *testing.T) {
	got := parseInt64Slice(nil)
	if len(got) != 0 {
		t.Errorf("got %v, want []", got)
	}
}

// ============================
// parseInt64Map
// ============================

func TestParseInt64Map_Valid(t *testing.T) {
	got := parseInt64Map([]string{"10", "20"})
	if !got[10] || !got[20] || len(got) != 2 {
		t.Errorf("got %v, want map with 10 and 20", got)
	}
}

func TestParseInt64Map_SkipInvalid(t *testing.T) {
	got := parseInt64Map([]string{"5", "nope"})
	if len(got) != 1 || !got[5] {
		t.Errorf("got %v, want map with only 5", got)
	}
}

// ============================
// isSubset
// ============================

func TestIsSubset_True(t *testing.T) {
	if !isSubset([]int64{1, 2}, []int64{1, 2, 3}) {
		t.Error("expected true")
	}
}

func TestIsSubset_False(t *testing.T) {
	if isSubset([]int64{1, 4}, []int64{1, 2, 3}) {
		t.Error("expected false")
	}
}

func TestIsSubset_EmptySub(t *testing.T) {
	if !isSubset(nil, []int64{1, 2}) {
		t.Error("empty sub is always a subset")
	}
}

func TestIsSubset_EmptySet(t *testing.T) {
	if isSubset([]int64{1}, nil) {
		t.Error("non-empty sub cannot be a subset of empty set")
	}
}

// ============================
// containsInt64
// ============================

func TestContainsInt64_Found(t *testing.T) {
	if !containsInt64([]int64{1, 2, 3}, 2) {
		t.Error("expected true")
	}
}

func TestContainsInt64_NotFound(t *testing.T) {
	if containsInt64([]int64{1, 2, 3}, 99) {
		t.Error("expected false")
	}
}

// ============================
// isoWeeksInYear
// ============================

func TestIsoWeeksInYear(t *testing.T) {
	cases := []struct {
		year int
		want int
	}{
		{2015, 53}, // known 53-week year
		{2026, 53}, // known 53-week year
		{2024, 52},
		{2023, 52},
	}
	for _, tc := range cases {
		got := isoWeeksInYear(tc.year)
		if got != tc.want {
			t.Errorf("isoWeeksInYear(%d) = %d, want %d", tc.year, got, tc.want)
		}
	}
}

// ============================
// prevISOWeek / nextISOWeek
// ============================

func TestPrevISOWeek_MidYear(t *testing.T) {
	y, w := prevISOWeek(2026, 5)
	if y != 2026 || w != 4 {
		t.Errorf("got (%d, %d), want (2026, 4)", y, w)
	}
}

func TestPrevISOWeek_YearBoundary(t *testing.T) {
	y, w := prevISOWeek(2026, 1)
	if y != 2025 {
		t.Errorf("year = %d, want 2025", y)
	}
	if w != isoWeeksInYear(2025) {
		t.Errorf("week = %d, want last week of 2025", w)
	}
}

func TestNextISOWeek_MidYear(t *testing.T) {
	y, w := nextISOWeek(2026, 5)
	if y != 2026 || w != 6 {
		t.Errorf("got (%d, %d), want (2026, 6)", y, w)
	}
}

func TestNextISOWeek_YearBoundary(t *testing.T) {
	lastWeek := isoWeeksInYear(2026)
	y, w := nextISOWeek(2026, lastWeek)
	if y != 2027 || w != 1 {
		t.Errorf("got (%d, %d), want (2027, 1)", y, w)
	}
}

// ============================
// niceCeil
// ============================

func TestNiceCeil(t *testing.T) {
	cases := []struct {
		in   float64
		want float64
	}{
		{0, 1},
		{1, 1},
		{1.1, 2},
		{2.5, 5},
		{7, 10},
		{15, 20},
		{55, 100},
	}
	for _, tc := range cases {
		got := niceCeil(tc.in)
		if got != tc.want {
			t.Errorf("niceCeil(%.1f) = %.0f, want %.0f", tc.in, got, tc.want)
		}
	}
}

// ============================
// yTickStep
// ============================

func TestYTickStep(t *testing.T) {
	cases := []struct{ max, want int }{
		{5, 1},
		{10, 1},
		{15, 2},
		{20, 2},
		{30, 5},
		{75, 10},
		{200, 20},
	}
	for _, tc := range cases {
		got := yTickStep(tc.max)
		if got != tc.want {
			t.Errorf("yTickStep(%d) = %d, want %d", tc.max, got, tc.want)
		}
	}
}

// ============================
// uniqueSortedInts
// ============================

func TestUniqueSortedInts(t *testing.T) {
	got := uniqueSortedInts([]int{3, 1, 2, 1, 3})
	if len(got) != 3 || got[0] != 1 || got[1] != 2 || got[2] != 3 {
		t.Errorf("got %v, want [1 2 3]", got)
	}
}

func TestUniqueSortedInts_NegativesSkipped(t *testing.T) {
	got := uniqueSortedInts([]int{-1, 0, 2})
	for _, v := range got {
		if v < 0 {
			t.Errorf("negative value %d should be skipped", v)
		}
	}
}

// ============================
// setToast
// ============================

func TestSetToast_Header(t *testing.T) {
	w := httptest.NewRecorder()
	setToast(w, "Game saved.")

	header := w.Header().Get("HX-Trigger")
	if header == "" {
		t.Fatal("HX-Trigger header not set")
	}

	var payload map[string]map[string]string
	if err := json.Unmarshal([]byte(header), &payload); err != nil {
		t.Fatalf("HX-Trigger is not valid JSON: %v", err)
	}

	msg, ok := payload["showToast"]["message"]
	if !ok {
		t.Fatal("showToast.message not found in HX-Trigger")
	}
	if msg != "Game saved." {
		t.Errorf("message = %q, want %q", msg, "Game saved.")
	}
}

func TestSetToast_EscapesSpecialChars(t *testing.T) {
	w := httptest.NewRecorder()
	setToast(w, `Alice's "game" saved.`)

	header := w.Header().Get("HX-Trigger")
	var payload map[string]map[string]string
	if err := json.Unmarshal([]byte(header), &payload); err != nil {
		t.Fatalf("HX-Trigger with special chars is not valid JSON: %v", err)
	}
}
