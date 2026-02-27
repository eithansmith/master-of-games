package game

import (
	"errors"
	"sort"
	"sync"
	"time"
)

type MemoryStore struct {
	mu sync.Mutex

	nextGameID   int64
	nextPlayerID int64
	nextTitleID  int64

	games   []Game
	players []Player
	titles  []Title

	tiebreakers map[string]Tiebreaker // key = scope + "|" + scopeKey
}

//goland:noinspection GoUnusedExportedFunction
func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{
		nextGameID:   1,
		nextPlayerID: 1,
		nextTitleID:  1,
		tiebreakers:  map[string]Tiebreaker{},
	}

	// Seed with the historical hardcoded lists.
	for _, name := range SeedPlayers {
		_, _ = s.AddPlayer(name)
	}
	for _, name := range SeedTitles {
		_, _ = s.AddTitle(name)
	}

	return s
}

func tbKey(scope, scopeKey string) string { return scope + "|" + scopeKey }

// ============================
// Games
// ============================

func (s *MemoryStore) AddGame(g Game) (Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	g.ID = s.nextGameID
	s.nextGameID++
	if !g.IsActive {
		g.IsActive = true
	}
	s.games = append(s.games, g)
	return g, nil
}

func (s *MemoryStore) DeleteGame(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.games {
		if s.games[i].ID == id {
			s.games = append(s.games[:i], s.games[i+1:]...)
			return nil
		}
	}
	return errors.New("game not found")
}

func (s *MemoryStore) SetGameActive(id int64, active bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.games {
		if s.games[i].ID == id {
			s.games[i].IsActive = active
			return nil
		}
	}
	return errors.New("game not found")
}

func (s *MemoryStore) RecentGames(limit int) ([]Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Game, len(s.games))
	copy(out, s.games)

	sort.Slice(out, func(i, j int) bool {
		if out[i].PlayedAt.Equal(out[j].PlayedAt) {
			return out[i].ID > out[j].ID
		}
		return out[i].PlayedAt.After(out[j].PlayedAt)
	})

	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out, nil
}

func (s *MemoryStore) GetWeek(year, week int) ([]Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Game, 0, len(s.games))
	for _, g := range s.games {
		if g.PlayedAt.Year() == year && g.PlayedAt.Weekday() == time.Weekday(week) {
			out = append(out, g)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PlayedAt.Before(out[j].PlayedAt) })

	return out, nil
}

func (s *MemoryStore) GetYear(year int) ([]Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Game, 0, len(s.games))
	for _, g := range s.games {
		if g.PlayedAt.Year() == year {
			out = append(out, g)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].PlayedAt.Before(out[j].PlayedAt) })

	return out, nil
}

// ============================
// Players
// ============================

func (s *MemoryStore) ListPlayers() ([]Player, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Player, len(s.players))
	copy(out, s.players)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (s *MemoryStore) AddPlayer(name string) (Player, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p := Player{ID: s.nextPlayerID, Name: name, IsActive: true}
	s.nextPlayerID++
	s.players = append(s.players, p)
	return p, nil
}

func (s *MemoryStore) UpdatePlayer(id int64, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.players {
		if s.players[i].ID == id {
			s.players[i].Name = name
			return nil
		}
	}
	return errors.New("player not found")
}

func (s *MemoryStore) SetPlayerActive(id int64, active bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.players {
		if s.players[i].ID == id {
			s.players[i].IsActive = active
			return nil
		}
	}
	return errors.New("player not found")
}

func (s *MemoryStore) DeletePlayer(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.players {
		if s.players[i].ID == id {
			s.players = append(s.players[:i], s.players[i+1:]...)
			return nil
		}
	}
	return errors.New("player not found")
}

// ============================
// Titles
// ============================

func (s *MemoryStore) ListTitles() ([]Title, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Title, len(s.titles))
	copy(out, s.titles)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, nil
}

func (s *MemoryStore) AddTitle(name string) (Title, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t := Title{ID: s.nextTitleID, Name: name, IsActive: true}
	s.nextTitleID++
	s.titles = append(s.titles, t)
	return t, nil
}

func (s *MemoryStore) UpdateTitle(id int64, name string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.titles {
		if s.titles[i].ID == id {
			s.titles[i].Name = name
			return nil
		}
	}
	return errors.New("title not found")
}

func (s *MemoryStore) SetTitleActive(id int64, active bool) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.titles {
		if s.titles[i].ID == id {
			s.titles[i].IsActive = active
			return nil
		}
	}
	return errors.New("title not found")
}

func (s *MemoryStore) DeleteTitle(id int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.titles {
		if s.titles[i].ID == id {
			s.titles = append(s.titles[:i], s.titles[i+1:]...)
			return nil
		}
	}
	return errors.New("title not found")
}

// ============================
// Misc
// ============================

func IsWeekdayLocal(t time.Time) bool {
	wd := t.Weekday()
	return wd >= time.Monday && wd <= time.Friday
}

func (s *MemoryStore) GetTiebreaker(scope, scopeKey string) (Tiebreaker, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tb, ok := s.tiebreakers[tbKey(scope, scopeKey)]
	if !ok {
		return Tiebreaker{}, ok, errors.New("tiebreaker not found")
	}
	return tb, ok, nil
}

func (s *MemoryStore) SetTiebreaker(tb Tiebreaker) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tiebreakers[tbKey(tb.Scope, tb.ScopeKey)] = tb
	return nil
}
