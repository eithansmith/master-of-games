package game

import (
	"sort"
)

type RaceMetric string

const (
	RaceMetricWins RaceMetric = "wins"
)

type RaceSeries struct {
	PlayerID int64
	Name     string
	Values   []float64
}

type YearRace struct {
	Year   int
	Weeks  []int
	Series []RaceSeries
}

// ComputeYearRace builds cumulative weekly data for a given metric.
func ComputeYearRace(
	games []Game,
	year int,
	metric RaceMetric,
	topN int,
	players []Player, // pass active players so we have names
) YearRace {

	// 1️⃣ Group games by week
	byWeek := map[int][]Game{}
	for _, g := range games {
		if g.PlayedAt.Year() != year {
			continue
		}
		_, week := g.PlayedAt.ISOWeek()
		byWeek[week] = append(byWeek[week], g)
	}

	// Collect sorted week list
	var weeks []int
	for w := range byWeek {
		weeks = append(weeks, w)
	}
	sort.Ints(weeks)

	// 2️⃣ Initialize cumulative stats
	type stat struct {
		Wins int
	}
	stats := map[int64]*stat{}

	// Initialize series for all players
	seriesMap := map[int64]*RaceSeries{}
	for _, p := range players {
		stats[p.ID] = &stat{}
		seriesMap[p.ID] = &RaceSeries{
			PlayerID: p.ID,
			Name:     p.Name,
			Values:   []float64{},
		}
	}

	// 3️⃣ Walk week by week
	for _, week := range weeks {
		gamesThisWeek := byWeek[week]

		for _, g := range gamesThisWeek {
			for _, winnerID := range g.WinnerIDs {
				stats[winnerID].Wins++
			}
		}

		// Snapshot cumulative values
		for _, p := range players {
			switch metric {
			case RaceMetricWins:
				seriesMap[p.ID].Values = append(
					seriesMap[p.ID].Values,
					float64(stats[p.ID].Wins),
				)
			}
		}
	}

	// 4️⃣ Determine topN players by final value
	type finalScore struct {
		PlayerID int64
		Value    float64
	}

	var finals []finalScore
	for _, s := range seriesMap {
		if len(s.Values) == 0 {
			continue
		}
		finals = append(finals, finalScore{
			PlayerID: s.PlayerID,
			Value:    s.Values[len(s.Values)-1],
		})
	}

	sort.Slice(finals, func(i, j int) bool {
		return finals[i].Value > finals[j].Value
	})

	if topN > len(finals) {
		topN = len(finals)
	}

	topSet := map[int64]bool{}
	for i := 0; i < topN; i++ {
		topSet[finals[i].PlayerID] = true
	}

	// 5️⃣ Build final series slice
	var finalSeries []RaceSeries
	for _, s := range seriesMap {
		if topSet[s.PlayerID] {
			finalSeries = append(finalSeries, *s)
		}
	}

	return YearRace{
		Year:   year,
		Weeks:  weeks,
		Series: finalSeries,
	}
}
