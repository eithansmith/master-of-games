package game

import "time"

type Player struct {
	ID   int64
	Name string
}

type Title struct {
	ID   int64
	Name string
}

type Game struct {
	ID       int64
	PlayedAt time.Time

	TitleID int64
	Title   string // denormalized for reads (join)

	ParticipantIDs []int64
	WinnerIDs      []int64
	Notes          string
}

type Tiebreaker struct {
	Scope    string // "weekly" | "yearly"
	ScopeKey string // "2026-W07" | "2026"

	TiedPlayerIDs []int64
	WinnerID      int64
	Method        string // "chance"
	DecidedAt     time.Time
}
