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

type Tiebreaker struct {
	Scope         string // "weekly" | "yearly"
	ScopeKey      string // "2026-W07" | "2026"
	TiedPlayerIDs []int
	WinnerID      int
	Method        string // "chance"
	DecidedAt     time.Time
}
