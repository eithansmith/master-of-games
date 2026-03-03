package game

import (
	"testing"
	"time"
)

func TestIsWeekdayLocal(t *testing.T) {
	cases := []struct {
		name string
		t    time.Time
		want bool
	}{
		{"Monday", time.Date(2026, 1, 5, 12, 0, 0, 0, time.UTC), true},
		{"Tuesday", time.Date(2026, 1, 6, 12, 0, 0, 0, time.UTC), true},
		{"Wednesday", time.Date(2026, 1, 7, 12, 0, 0, 0, time.UTC), true},
		{"Thursday", time.Date(2026, 1, 8, 12, 0, 0, 0, time.UTC), true},
		{"Friday", time.Date(2026, 1, 9, 12, 0, 0, 0, time.UTC), true},
		{"Saturday", time.Date(2026, 1, 10, 12, 0, 0, 0, time.UTC), false},
		{"Sunday", time.Date(2026, 1, 11, 12, 0, 0, 0, time.UTC), false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := IsWeekdayLocal(tc.t)
			if got != tc.want {
				t.Errorf("IsWeekdayLocal(%s %s) = %v, want %v",
					tc.name, tc.t.Format("2006-01-02"), got, tc.want)
			}
		})
	}
}
