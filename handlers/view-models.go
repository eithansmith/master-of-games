package handlers

import "github.com/eithansmith/master-of-games/game"

type HomeForm struct {
	Title        string
	PlayedAt     string
	Participants map[int]bool
	Winners      map[int]bool
}

type HomeVM struct {
	Title     string
	Version   string
	BuildTime string
	StartTime string
	YearNow   int

	Players []string
	Titles  []string
	Games   []game.Game

	FormError string
	Form      HomeForm
}

type WeekVM struct {
	Title     string
	Version   string
	BuildTime string
	StartTime string
	YearNow   int

	Year int
	Week int

	Years []int
	Weeks []int

	PrevYear int
	PrevWeek int
	HasPrev  bool

	NextYear int
	NextWeek int
	HasNext  bool

	Players []string

	TotalGames int
	Wins       map[int]int

	TopIDs        []int
	WinnerID      *int
	TieUnresolved bool

	FormError string
}

type YearVM struct {
	Title     string
	Version   string
	BuildTime string
	StartTime string
	YearNow   int

	Year    int
	Players []string

	Stats []game.PlayerYearStats

	Qualifiers    []int
	TopIDs        []int
	WinnerID      *int
	TieUnresolved bool

	FormError string
}
