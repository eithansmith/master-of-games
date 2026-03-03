package game

import (
	"context"
	"testing"
	"time"
)

var ctx = context.Background()

func newStore() *MemoryStore {
	return &MemoryStore{
		nextGameID:   1,
		nextPlayerID: 1,
		nextTitleID:  1,
		tiebreakers:  map[string]Tiebreaker{},
	}
}

// ============================
// Players
// ============================

func TestMemoryStore_AddAndListPlayers(t *testing.T) {
	s := newStore()
	_, _ = s.AddPlayer(ctx, "Alice")
	_, _ = s.AddPlayer(ctx, "Bob")

	players, err := s.ListPlayers(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(players) != 2 {
		t.Fatalf("len = %d, want 2", len(players))
	}
}

func TestMemoryStore_ListPlayers_SortActiveFirst(t *testing.T) {
	s := newStore()
	_, _ = s.AddPlayer(ctx, "Zelda")
	_, _ = s.AddPlayer(ctx, "Alice")
	p3, _ := s.AddPlayer(ctx, "Bob")
	_ = s.SetPlayerActive(ctx, p3.ID, false)

	players, _ := s.ListPlayers(ctx)

	// Active players should come before inactive.
	for i, p := range players {
		if !p.IsActive {
			for _, prev := range players[:i] {
				if prev.IsActive {
					continue
				}
				t.Errorf("inactive player appeared before active player: %v", players)
			}
		}
	}

	// Among active players, alphabetical order.
	var activePlayers []Player
	for _, p := range players {
		if p.IsActive {
			activePlayers = append(activePlayers, p)
		}
	}
	if len(activePlayers) >= 2 {
		for i := 1; i < len(activePlayers); i++ {
			if activePlayers[i].Name < activePlayers[i-1].Name {
				t.Errorf("active players not sorted alphabetically: %v", activePlayers)
			}
		}
	}
}

func TestMemoryStore_UpdatePlayer(t *testing.T) {
	s := newStore()
	p, _ := s.AddPlayer(ctx, "Alice")
	_ = s.UpdatePlayer(ctx, p.ID, "Alicia")

	players, _ := s.ListPlayers(ctx)
	if players[0].Name != "Alicia" {
		t.Errorf("Name = %q, want %q", players[0].Name, "Alicia")
	}
}

func TestMemoryStore_SetPlayerActive(t *testing.T) {
	s := newStore()
	p, _ := s.AddPlayer(ctx, "Alice")

	_ = s.SetPlayerActive(ctx, p.ID, false)
	players, _ := s.ListPlayers(ctx)
	if players[0].IsActive {
		t.Error("player should be inactive")
	}

	_ = s.SetPlayerActive(ctx, p.ID, true)
	players, _ = s.ListPlayers(ctx)
	if !players[0].IsActive {
		t.Error("player should be active again")
	}
}

func TestMemoryStore_DeletePlayer(t *testing.T) {
	s := newStore()
	p, _ := s.AddPlayer(ctx, "Alice")
	_ = s.DeletePlayer(ctx, p.ID)

	players, _ := s.ListPlayers(ctx)
	if len(players) != 0 {
		t.Errorf("len = %d, want 0 after delete", len(players))
	}
}

// ============================
// Titles
// ============================

func TestMemoryStore_AddAndListTitles(t *testing.T) {
	s := newStore()
	_, _ = s.AddTitle(ctx, "Chess")
	_, _ = s.AddTitle(ctx, "Dominion")

	titles, err := s.ListTitles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(titles) != 2 {
		t.Fatalf("len = %d, want 2", len(titles))
	}
}

func TestMemoryStore_ListTitles_SortActiveFirst(t *testing.T) {
	s := newStore()
	_, _ = s.AddTitle(ctx, "Dominion")
	_, _ = s.AddTitle(ctx, "Chess")
	t3, _ := s.AddTitle(ctx, "Azul")
	_ = s.SetTitleActive(ctx, t3.ID, false)

	titles, _ := s.ListTitles(ctx)

	for i, title := range titles {
		if !title.IsActive {
			for _, prev := range titles[:i] {
				if !prev.IsActive {
					t.Errorf("inactive title appeared before another inactive: %v", titles)
				}
			}
		}
	}
}

func TestMemoryStore_UpdateTitle(t *testing.T) {
	s := newStore()
	tt, _ := s.AddTitle(ctx, "Chess")
	_ = s.UpdateTitle(ctx, tt.ID, "Speed Chess")

	titles, _ := s.ListTitles(ctx)
	if titles[0].Name != "Speed Chess" {
		t.Errorf("Name = %q, want %q", titles[0].Name, "Speed Chess")
	}
}

func TestMemoryStore_SetTitleActive(t *testing.T) {
	s := newStore()
	tt, _ := s.AddTitle(ctx, "Chess")

	_ = s.SetTitleActive(ctx, tt.ID, false)
	titles, _ := s.ListTitles(ctx)
	if titles[0].IsActive {
		t.Error("title should be inactive")
	}
}

// ============================
// Games
// ============================

func TestMemoryStore_AddAndRecentGames(t *testing.T) {
	s := newStore()
	g := Game{
		PlayedAt:       time.Date(2026, 1, 5, 12, 0, 0, 0, time.UTC),
		TitleID:        1,
		ParticipantIDs: []int64{1},
		WinnerIDs:      []int64{1},
	}
	_, _ = s.AddGame(ctx, g)

	games, err := s.RecentGames(ctx, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(games) != 1 {
		t.Fatalf("len = %d, want 1", len(games))
	}
	if !games[0].IsActive {
		t.Error("new game should be active")
	}
}

func TestMemoryStore_RecentGames_ActiveFirst(t *testing.T) {
	s := newStore()
	earlier := time.Date(2026, 1, 5, 12, 0, 0, 0, time.UTC)
	later := time.Date(2026, 1, 6, 12, 0, 0, 0, time.UTC)

	g1, _ := s.AddGame(ctx, Game{PlayedAt: earlier, WinnerIDs: []int64{1}})
	g2, _ := s.AddGame(ctx, Game{PlayedAt: later, WinnerIDs: []int64{2}})
	_ = s.SetGameActive(ctx, g1.ID, false)

	games, _ := s.RecentGames(ctx, 10)

	// g2 (active, later) should come first; g1 (inactive) should come last.
	if games[0].ID != g2.ID {
		t.Errorf("first game should be active g2, got id=%d", games[0].ID)
	}
	if games[1].ID != g1.ID {
		t.Errorf("second game should be inactive g1, got id=%d", games[1].ID)
	}
}

func TestMemoryStore_RecentGames_Limit(t *testing.T) {
	s := newStore()
	for i := 0; i < 5; i++ {
		_, _ = s.AddGame(ctx, Game{PlayedAt: time.Date(2026, 1, i+1, 12, 0, 0, 0, time.UTC)})
	}

	games, _ := s.RecentGames(ctx, 3)
	if len(games) != 3 {
		t.Errorf("len = %d, want 3", len(games))
	}
}

func TestMemoryStore_SetGameActive(t *testing.T) {
	s := newStore()
	g, _ := s.AddGame(ctx, Game{PlayedAt: time.Now()})

	_ = s.SetGameActive(ctx, g.ID, false)
	games, _ := s.RecentGames(ctx, 10)
	if games[0].IsActive {
		t.Error("game should be inactive")
	}
}

func TestMemoryStore_DeleteGame(t *testing.T) {
	s := newStore()
	g, _ := s.AddGame(ctx, Game{PlayedAt: time.Now()})
	_ = s.DeleteGame(ctx, g.ID)

	games, _ := s.RecentGames(ctx, 10)
	if len(games) != 0 {
		t.Errorf("len = %d, want 0 after DeleteGame", len(games))
	}
}

func TestMemoryStore_DeleteTitle(t *testing.T) {
	s := newStore()
	tt, _ := s.AddTitle(ctx, "Chess")
	_ = s.DeleteTitle(ctx, tt.ID)

	titles, _ := s.ListTitles(ctx)
	if len(titles) != 0 {
		t.Errorf("len = %d, want 0 after DeleteTitle", len(titles))
	}
}

// ============================
// Tiebreakers
// ============================

func TestMemoryStore_SetAndGetTiebreaker(t *testing.T) {
	s := newStore()
	tb := Tiebreaker{
		Scope:    "weekly",
		ScopeKey: "2026-W01",
		WinnerID: 42,
	}
	_ = s.SetTiebreaker(ctx, tb)

	got, ok, err := s.GetTiebreaker(ctx, "weekly", "2026-W01")
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected tiebreaker to be found")
	}
	if got.WinnerID != 42 {
		t.Errorf("WinnerID = %d, want 42", got.WinnerID)
	}
}

func TestMemoryStore_GetTiebreaker_Missing(t *testing.T) {
	s := newStore()
	_, ok, _ := s.GetTiebreaker(ctx, "weekly", "2026-W99")
	if ok {
		t.Error("expected not found for missing tiebreaker")
	}
}

func TestMemoryStore_SetTiebreaker_Overwrites(t *testing.T) {
	s := newStore()
	_ = s.SetTiebreaker(ctx, Tiebreaker{Scope: "weekly", ScopeKey: "2026-W01", WinnerID: 1})
	_ = s.SetTiebreaker(ctx, Tiebreaker{Scope: "weekly", ScopeKey: "2026-W01", WinnerID: 2})

	got, _, _ := s.GetTiebreaker(ctx, "weekly", "2026-W01")
	if got.WinnerID != 2 {
		t.Errorf("WinnerID = %d, want 2 (overwritten)", got.WinnerID)
	}
}
