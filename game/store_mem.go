package game

import (
	"fmt"
	"sort"
	"sync"
	"time"
)

type MemoryStore struct {
	mu     sync.Mutex
	nextID int64
	games  []Game

	tiebreakers map[string]Tiebreaker // key = scope + "|" + scopeKey
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		nextID:      1,
		tiebreakers: map[string]Tiebreaker{},
	}
}

func tbKey(scope, scopeKey string) string { return scope + "|" + scopeKey }

func (s *MemoryStore) AddGame(g Game) Game {
	s.mu.Lock()
	defer s.mu.Unlock()

	g.ID = s.nextID
	s.nextID++
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
		return out[i].PlayedAt.After(out[j].PlayedAt)
	})

	if limit > 0 && len(out) > limit {
		out = out[:limit]
	}
	return out
}

func PlayerNameByID(id int) (string, error) {
	if id < 0 || id >= len(Players) {
		return "", fmt.Errorf("invalid player id: %d", id)
	}
	return Players[id], nil
}

func IsWeekdayLocal(t time.Time) bool {
	// Using local time; later we can set a fixed location if you want.
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
