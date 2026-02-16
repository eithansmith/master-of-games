package handlers

import "github.com/eithansmith/master-of-games/game"

type HomeForm struct {
	TitleID      int64
	PlayedAt     string
	Participants map[int64]bool
	Winners      map[int64]bool
	Notes        string
}

type HomeVM struct {
	Title     string
	Version   string
	BuildTime string
	StartTime string
	YearNow   int

	Players      []game.Player
	PlayerNames  map[int64]string
	Titles       []game.Title
	Games        []game.Game
	ShowAllGames bool

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

	Players     []game.Player
	PlayerNames map[int64]string

	TotalGames int
	Wins       map[int64]int

	TopIDs        []int64
	WinnerID      *int64
	TieUnresolved bool

	FormError string
}

type YearVM struct {
	Title     string
	Version   string
	BuildTime string
	StartTime string
	YearNow   int

	Year int

	Players     []game.Player
	PlayerNames map[int64]string

	Stats []game.PlayerYearStats

	Qualifiers    []int64
	TopIDs        []int64
	WinnerID      *int64
	TieUnresolved bool

	FormError string
}

type PlayersVM struct {
	Title     string
	Version   string
	BuildTime string
	StartTime string
	YearNow   int

	Players   []game.Player
	FormError string
}

type TitlesVM struct {
	Title     string
	Version   string
	BuildTime string
	StartTime string
	YearNow   int

	Titles    []game.Title
	FormError string
}
