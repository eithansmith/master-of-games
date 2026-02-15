package game

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresStore struct {
	db  *pgxpool.Pool
	now func() time.Time
}

func NewPostgresStore(db *pgxpool.Pool) *PostgresStore {
	return &PostgresStore{
		db:  db,
		now: time.Now,
	}
}

// AddGame inserts a new game row and returns the saved game with ID populated.
func (s *PostgresStore) AddGame(g Game) Game {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Adjust these field names if your Game differs.
	// Assumes:
	//   g.Title string
	//   g.PlayedAt time.Time
	//   g.ID int64
	err := s.db.QueryRow(ctx,
		`INSERT INTO app.games (title, played_at) VALUES ($1, $2) RETURNING id`,
		g.Title,
		g.PlayedAt,
	).Scan(&g.ID)
	if err != nil {
		// Match MemoryStore behavior: it never returns an error.
		// For now, panic/log is reasonable; later we can plumb errors up.
		// If you already have a logger on Server, we can pass it in and log instead.
		panic(fmt.Errorf("AddGame insert: %w", err))
	}

	return g
}

// DeleteGame deletes by id; returns true if a row was deleted.
func (s *PostgresStore) DeleteGame(id int64) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ct, err := s.db.Exec(ctx, `DELETE FROM app.games WHERE id = $1`, id)
	if err != nil {
		panic(fmt.Errorf("DeleteGame: %w", err))
	}
	return ct.RowsAffected() > 0
}

// RecentGames returns most recent by played_at desc (tie-break by id desc).
func (s *PostgresStore) RecentGames(limit int) []Game {
	if limit <= 0 {
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.db.Query(ctx,
		`SELECT id, title, played_at
		   FROM app.games
		  ORDER BY played_at DESC, id DESC
		  LIMIT $1`,
		limit,
	)
	if err != nil {
		panic(fmt.Errorf("RecentGames query: %w", err))
	}
	defer rows.Close()

	out := make([]Game, 0, limit)
	for rows.Next() {
		var g Game
		if err := rows.Scan(&g.ID, &g.Title, &g.PlayedAt); err != nil {
			panic(fmt.Errorf("RecentGames scan: %w", err))
		}
		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		panic(fmt.Errorf("RecentGames rows: %w", err))
	}
	return out
}

func (s *PostgresStore) GetTiebreaker(scope, scopeKey string) (Tiebreaker, bool) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var raw []byte
	err := s.db.QueryRow(ctx,
		`SELECT data FROM app.tiebreakers WHERE scope = $1 AND scope_key = $2`,
		scope, scopeKey,
	).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Tiebreaker{}, false
		}
		panic(fmt.Errorf("GetTiebreaker query: %w", err))
	}

	var tb Tiebreaker
	if err := json.Unmarshal(raw, &tb); err != nil {
		panic(fmt.Errorf("GetTiebreaker unmarshal: %w", err))
	}
	return tb, true
}

func (s *PostgresStore) SetTiebreaker(tb Tiebreaker) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	raw, err := json.Marshal(tb)
	if err != nil {
		panic(fmt.Errorf("SetTiebreaker marshal: %w", err))
	}

	// Adjust field names if Tiebreaker differs:
	// assumes tb.Scope string, tb.ScopeKey string
	_, err = s.db.Exec(ctx, `
		INSERT INTO app.tiebreakers (scope, scope_key, data)
		VALUES ($1, $2, $3)
		ON CONFLICT (scope, scope_key)
		DO UPDATE SET data = EXCLUDED.data
	`, tb.Scope, tb.ScopeKey, raw)
	if err != nil {
		panic(fmt.Errorf("SetTiebreaker upsert: %w", err))
	}
}
