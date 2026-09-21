package game

import (
	"context"
	"errors"
	"sort"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type MemoryStore struct {
	mu sync.Mutex

	nextGameID    int64
	nextPlayerID  int64
	nextTitleID   int64
	nextUserID    int64
	nextSessionID int64

	games    []Game
	players  []Player
	titles   []Title
	users    []User
	sessions []Session

	tiebreakers map[string]Tiebreaker // key = scope + "|" + scopeKey
}

//goland:noinspection GoUnusedExportedFunction
func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{
		nextGameID:    1,
		nextPlayerID:  1,
		nextTitleID:   1,
		nextUserID:    1,
		nextSessionID: 1,
		tiebreakers:   map[string]Tiebreaker{},
	}

	// Seed with the historical hardcoded lists.
	for _, name := range SeedPlayers {
		_, _ = s.AddPlayer(context.Background(), name)
	}
	for _, name := range SeedTitles {
		_, _ = s.AddTitle(context.Background(), name)
	}

	// Seed a dev user (username "dev", password "dev") so local `go run`
	// without Postgres still has a working login.
	if hash, err := bcrypt.GenerateFromPassword([]byte("dev"), bcrypt.DefaultCost); err == nil {
		_ = s.CreateUser(context.Background(), "dev", string(hash))
	}

	return s
}

func tbKey(scope, scopeKey string) string { return scope + "|" + scopeKey }

// ============================
// Games
// ============================

func (s *MemoryStore) AddGame(_ context.Context, g Game) (Game, error) {
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

func (s *MemoryStore) DeleteGame(_ context.Context, id int64) error {
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

func (s *MemoryStore) SetGameActive(_ context.Context, id int64, active bool) error {
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

func (s *MemoryStore) RecentGames(_ context.Context, limit int) ([]Game, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Game, len(s.games))
	copy(out, s.games)

	sort.Slice(out, func(i, j int) bool {
		if out[i].IsActive != out[j].IsActive {
			return out[i].IsActive
		}
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

func (s *MemoryStore) GetWeek(_ context.Context, year, week int) ([]Game, error) {
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

func (s *MemoryStore) GetYear(_ context.Context, year int) ([]Game, error) {
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

//goland:noinspection DuplicatedCode
func (s *MemoryStore) ListPlayers(_ context.Context) ([]Player, error) {
	//goland:noinspection DuplicatedCode
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Player, len(s.players))
	copy(out, s.players)
	sort.Slice(out, func(i, j int) bool {
		if out[i].IsActive != out[j].IsActive {
			return out[i].IsActive
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func (s *MemoryStore) AddPlayer(_ context.Context, name string) (Player, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p := Player{ID: s.nextPlayerID, Name: name, IsActive: true}
	s.nextPlayerID++
	s.players = append(s.players, p)
	return p, nil
}

func (s *MemoryStore) UpdatePlayer(_ context.Context, id int64, name string) error {
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

func (s *MemoryStore) SetPlayerActive(_ context.Context, id int64, active bool) error {
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

func (s *MemoryStore) DeletePlayer(_ context.Context, id int64) error {
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

//goland:noinspection DuplicatedCode
func (s *MemoryStore) ListTitles(_ context.Context) ([]Title, error) {
	//goland:noinspection DuplicatedCode
	s.mu.Lock()
	defer s.mu.Unlock()

	out := make([]Title, len(s.titles))
	copy(out, s.titles)
	sort.Slice(out, func(i, j int) bool {
		if out[i].IsActive != out[j].IsActive {
			return out[i].IsActive
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

func (s *MemoryStore) AddTitle(_ context.Context, name string) (Title, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	t := Title{ID: s.nextTitleID, Name: name, IsActive: true}
	s.nextTitleID++
	s.titles = append(s.titles, t)
	return t, nil
}

func (s *MemoryStore) UpdateTitle(_ context.Context, id int64, name string) error {
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

func (s *MemoryStore) SetTitleActive(_ context.Context, id int64, active bool) error {
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

func (s *MemoryStore) DeleteTitle(_ context.Context, id int64) error {
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

func (s *MemoryStore) GetTiebreaker(_ context.Context, scope, scopeKey string) (Tiebreaker, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	tb, ok := s.tiebreakers[tbKey(scope, scopeKey)]
	if !ok {
		return Tiebreaker{}, ok, errors.New("tiebreaker not found")
	}
	return tb, ok, nil
}

func (s *MemoryStore) SetTiebreaker(_ context.Context, tb Tiebreaker) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.tiebreakers[tbKey(tb.Scope, tb.ScopeKey)] = tb
	return nil
}

// ============================
// Auth
// ============================

func (s *MemoryStore) GetUserByUsername(_ context.Context, username string) (User, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if u.Username == username {
			return u, true, nil
		}
	}
	return User{}, false, nil
}

func (s *MemoryStore) CreateUser(_ context.Context, username, passwordHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, u := range s.users {
		if u.Username == username {
			return nil
		}
	}
	s.users = append(s.users, User{ID: s.nextUserID, Username: username, PasswordHash: passwordHash})
	s.nextUserID++
	return nil
}

func (s *MemoryStore) CreateSession(_ context.Context, sess Session) (Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess.ID = s.nextSessionID
	s.nextSessionID++
	s.sessions = append(s.sessions, sess)
	return sess, nil
}

func (s *MemoryStore) GetSessionByTokenHash(_ context.Context, tokenHash string) (Session, bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, sess := range s.sessions {
		if sess.TokenHash == tokenHash {
			return sess, true, nil
		}
	}
	return Session{}, false, nil
}

func (s *MemoryStore) DeleteSession(_ context.Context, tokenHash string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.sessions {
		if s.sessions[i].TokenHash == tokenHash {
			s.sessions = append(s.sessions[:i], s.sessions[i+1:]...)
			return nil
		}
	}
	return nil
}

func (s *MemoryStore) DeleteExpiredSessions(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	out := s.sessions[:0]
	now := time.Now()
	for _, sess := range s.sessions {
		if sess.ExpiresAt.After(now) {
			out = append(out, sess)
		}
	}
	s.sessions = out
	return nil
}
