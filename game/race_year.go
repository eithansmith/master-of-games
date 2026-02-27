package game

import "sort"

type RaceMetric string

const (
	RaceMetricWins RaceMetric = "wins"
)

type RaceSeries struct {
	PlayerID int64
	Name     string
	Values   []float64 // aligned to Weeks
}

type YearRace struct {
	Year   int
	Weeks  []int
	Series []RaceSeries
}

// ComputeYearRace builds cumulative weekly data for the given year.
// v1 supports metric=wins only, and filters to Top N by final value.
func ComputeYearRace(
	games []Game,
	year int,
	metric RaceMetric,
	topN int,
	players []Player,
) YearRace {
	// Group games by ISO week (only games in the target year)
	byWeek := map[int][]Game{}
	for _, g := range games {
		if g.PlayedAt.Year() != year {
			continue
		}
		_, w := g.PlayedAt.ISOWeek()
		byWeek[w] = append(byWeek[w], g)
	}

	// Sorted week list
	var weeks []int
	for w := range byWeek {
		weeks = append(weeks, w)
	}
	sort.Ints(weeks)

	// init stats + series for active players (or all players if you prefer)
	type stat struct{ wins int }
	stats := map[int64]*stat{}
	series := map[int64]*RaceSeries{}

	for _, p := range players {
		if !p.IsActive {
			continue
		}
		stats[p.ID] = &stat{}
		series[p.ID] = &RaceSeries{
			PlayerID: p.ID,
			Name:     p.Name,
			Values:   make([]float64, 0, len(weeks)),
		}
	}

	// Walk weeks, accumulate, then snapshot
	for _, w := range weeks {
		for _, g := range byWeek[w] {
			for _, winnerID := range g.WinnerIDs {
				if st, ok := stats[winnerID]; ok {
					st.wins++
				}
			}
		}

		for pid := range series {
			switch metric {
			case RaceMetricWins:
				series[pid].Values = append(series[pid].Values, float64(stats[pid].wins))
			default:
				series[pid].Values = append(series[pid].Values, 0)
			}
		}
	}

	// No data
	if len(weeks) == 0 {
		return YearRace{Year: year}
	}

	// Rank by final value and take top N
	type final struct {
		id    int64
		score float64
	}
	finals := make([]final, 0, len(series))
	for _, s := range series {
		if len(s.Values) == 0 {
			continue
		}
		finals = append(finals, final{id: s.PlayerID, score: s.Values[len(s.Values)-1]})
	}
	sort.Slice(finals, func(i, j int) bool { return finals[i].score > finals[j].score })

	if topN <= 0 {
		topN = 5
	}
	if topN > len(finals) {
		topN = len(finals)
	}

	out := make([]RaceSeries, 0, topN)
	for i := 0; i < topN; i++ {
		out = append(out, *series[finals[i].id])
	}

	return YearRace{
		Year:   year,
		Weeks:  weeks,
		Series: out,
	}
}
