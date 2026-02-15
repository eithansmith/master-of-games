package game

import (
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
		s.AddPlayer(name)
	}
	for _, name := range SeedTitles {
		s.AddTitle(name)
	}

	return s
}

func tbKey(scope, scopeKey string) string { return scope + "|" + scopeKey }

// ============================
// Games
// ============================

func (s *MemoryStore) AddGame(g Game) Game {
	s.mu.Lock()
	defer s.mu.Unlock()

	g.ID = s.nextGameID
	s.nextGameID++
	s.games = append(s.games, g)
	return g
}

func (s *MemoryStore) DeleteGame(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.games {
		if s.games[i].ID == id {
			s.games = append(s.games[:i], s.games[i+1:]...)
			return true
		}
	}
	return false
}

func (s *MemoryStore) RecentGames(limit int) []Game {
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
	return out
}

// ============================
// Players
// ============================

func (s *MemoryStore) ListPlayers() []Player {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Player, len(s.players))
	copy(out, s.players)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *MemoryStore) AddPlayer(name string) Player {
	s.mu.Lock()
	defer s.mu.Unlock()

	p := Player{ID: s.nextPlayerID, Name: name}
	s.nextPlayerID++
	s.players = append(s.players, p)
	return p
}

func (s *MemoryStore) UpdatePlayer(id int64, name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.players {
		if s.players[i].ID == id {
			s.players[i].Name = name
			return true
		}
	}
	return false
}

func (s *MemoryStore) DeletePlayer(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.players {
		if s.players[i].ID == id {
			s.players = append(s.players[:i], s.players[i+1:]...)
			return true
		}
	}
	return false
}

// ============================
// Titles
// ============================

func (s *MemoryStore) ListTitles() []Title {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Title, len(s.titles))
	copy(out, s.titles)
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (s *MemoryStore) AddTitle(name string) Title {
	s.mu.Lock()
	defer s.mu.Unlock()

	t := Title{ID: s.nextTitleID, Name: name}
	s.nextTitleID++
	s.titles = append(s.titles, t)
	return t
}

func (s *MemoryStore) UpdateTitle(id int64, name string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.titles {
		if s.titles[i].ID == id {
			s.titles[i].Name = name
			return true
		}
	}
	return false
}

func (s *MemoryStore) DeleteTitle(id int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.titles {
		if s.titles[i].ID == id {
			s.titles = append(s.titles[:i], s.titles[i+1:]...)
			return true
		}
	}
	return false
}

// ============================
// Misc
// ============================

func IsWeekdayLocal(t time.Time) bool {
	wd := t.Weekday()
	return wd >= time.Monday && wd <= time.Friday
}

func (s *MemoryStore) GetTiebreaker(scope, scopeKey string) (Tiebreaker, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tb, ok := s.tiebreakers[tbKey(scope, scopeKey)]
	return tb, ok
}

func (s *MemoryStore) SetTiebreaker(tb Tiebreaker) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tiebreakers[tbKey(tb.Scope, tb.ScopeKey)] = tb
}
