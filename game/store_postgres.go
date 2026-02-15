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

// ============================
// Games
// ============================

func (s *PostgresStore) AddGame(g Game) Game {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := s.db.QueryRow(ctx,
		`INSERT INTO app.games (title_id, played_at, participant_ids, winner_ids, notes)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id`,
		g.TitleID,
		g.PlayedAt,
		g.ParticipantIDs,
		g.WinnerIDs,
		g.Notes,
	).Scan(&g.ID)
	if err != nil {
		panic(fmt.Errorf("AddGame insert: %w", err))
	}
	return g
}

func (s *PostgresStore) DeleteGame(id int64) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ct, err := s.db.Exec(ctx, `DELETE FROM app.games WHERE id = $1`, id)
	if err != nil {
		panic(fmt.Errorf("DeleteGame: %w", err))
	}
	return ct.RowsAffected() > 0
}

func (s *PostgresStore) RecentGames(limit int) []Game {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	q := `SELECT g.id, g.played_at, g.title_id, t.name, g.participant_ids, g.winner_ids, g.notes
		  FROM app.games g
		  JOIN app.titles t ON t.id = g.title_id
		 ORDER BY g.played_at DESC, g.id DESC`
	var args []any
	if limit > 0 {
		q += ` LIMIT $1`
		args = append(args, limit)
	}

	rows, err := s.db.Query(ctx, q, args...)
	if err != nil {
		panic(fmt.Errorf("RecentGames query: %w", err))
	}
	defer rows.Close()

	out := make([]Game, 0, maximum(0, limit))
	for rows.Next() {
		var g Game
		if err := rows.Scan(&g.ID, &g.PlayedAt, &g.TitleID, &g.Title, &g.ParticipantIDs, &g.WinnerIDs, &g.Notes); err != nil {
			panic(fmt.Errorf("RecentGames scan: %w", err))
		}
		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		panic(fmt.Errorf("RecentGames rows: %w", err))
	}
	return out
}

// ============================
// Players
// ============================

func (s *PostgresStore) ListPlayers() []Player {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.db.Query(ctx, `SELECT id, name FROM app.players ORDER BY name`)
	if err != nil {
		panic(fmt.Errorf("ListPlayers: %w", err))
	}
	defer rows.Close()

	var out []Player
	for rows.Next() {
		var p Player
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			panic(fmt.Errorf("ListPlayers scan: %w", err))
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		panic(fmt.Errorf("ListPlayers rows: %w", err))
	}
	return out
}

func (s *PostgresStore) AddPlayer(name string) Player {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var p Player
	err := s.db.QueryRow(ctx, `INSERT INTO app.players (name) VALUES ($1) RETURNING id, name`, name).Scan(&p.ID, &p.Name)
	if err != nil {
		panic(fmt.Errorf("AddPlayer: %w", err))
	}
	return p
}

func (s *PostgresStore) UpdatePlayer(id int64, name string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ct, err := s.db.Exec(ctx, `UPDATE app.players SET name=$2 WHERE id=$1`, id, name)
	if err != nil {
		panic(fmt.Errorf("UpdatePlayer: %w", err))
	}
	return ct.RowsAffected() > 0
}

func (s *PostgresStore) DeletePlayer(id int64) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Prevent deleting a player referenced by any game.
	var exists bool
	err := s.db.QueryRow(ctx,
		`SELECT EXISTS (
			SELECT 1 FROM app.games
			WHERE $1 = ANY(participant_ids) OR $1 = ANY(winner_ids)
		)`, id,
	).Scan(&exists)
	if err != nil {
		panic(fmt.Errorf("DeletePlayer check: %w", err))
	}
	if exists {
		return false
	}

	ct, err := s.db.Exec(ctx, `DELETE FROM app.players WHERE id=$1`, id)
	if err != nil {
		panic(fmt.Errorf("DeletePlayer: %w", err))
	}
	return ct.RowsAffected() > 0
}

// ============================
// Titles
// ============================

func (s *PostgresStore) ListTitles() []Title {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.db.Query(ctx, `SELECT id, name FROM app.titles ORDER BY name`)
	if err != nil {
		panic(fmt.Errorf("ListTitles: %w", err))
	}
	defer rows.Close()

	var out []Title
	for rows.Next() {
		var t Title
		if err := rows.Scan(&t.ID, &t.Name); err != nil {
			panic(fmt.Errorf("ListTitles scan: %w", err))
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		panic(fmt.Errorf("ListTitles rows: %w", err))
	}
	return out
}

func (s *PostgresStore) AddTitle(name string) Title {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var t Title
	err := s.db.QueryRow(ctx, `INSERT INTO app.titles (name) VALUES ($1) RETURNING id, name`, name).Scan(&t.ID, &t.Name)
	if err != nil {
		panic(fmt.Errorf("AddTitle: %w", err))
	}
	return t
}

func (s *PostgresStore) UpdateTitle(id int64, name string) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ct, err := s.db.Exec(ctx, `UPDATE app.titles SET name=$2 WHERE id=$1`, id, name)
	if err != nil {
		panic(fmt.Errorf("UpdateTitle: %w", err))
	}
	return ct.RowsAffected() > 0
}

func (s *PostgresStore) DeleteTitle(id int64) bool {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	ct, err := s.db.Exec(ctx, `DELETE FROM app.titles WHERE id=$1`, id)
	if err != nil {
		panic(fmt.Errorf("DeleteTitle: %w", err))
	}
	return ct.RowsAffected() > 0
}

// ============================
// Tiebreakers
// ============================

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
		panic(fmt.Errorf("GetTiebreaker: %w", err))
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

	b, err := json.Marshal(tb)
	if err != nil {
		panic(fmt.Errorf("SetTiebreaker marshal: %w", err))
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO app.tiebreakers (scope, scope_key, data)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (scope, scope_key)
		 DO UPDATE SET data = EXCLUDED.data`,
		tb.Scope, tb.ScopeKey, b,
	)
	if err != nil {
		panic(fmt.Errorf("SetTiebreaker upsert: %w", err))
	}
}

func maximum(a, b int) int {
	if a > b {
		return a
	}
	return b
}
