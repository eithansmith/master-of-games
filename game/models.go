package game

import "time"

type Game struct {
	ID             int64
	PlayedAt       time.Time
	Title          string
	ParticipantIDs []int
	WinnerIDs      []int
	Notes          string
}
