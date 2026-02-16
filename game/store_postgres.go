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

func (s *PostgresStore) AddGame(g Game) (Game, error) {
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
		return Game{}, fmt.Errorf("AddGame: %w", err)
	}
	return g, nil
}

func (s *PostgresStore) DeleteGame(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.Exec(ctx, `DELETE FROM app.games WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("DeleteGame: %w", err)
	}
	return nil
}

func (s *PostgresStore) SetGameActive(id int64, active bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.Exec(ctx, `UPDATE app.games SET is_active = $2 WHERE id = $1`, id, active)
	if err != nil {
		return fmt.Errorf("SetGameActive: %w", err)
	}
	return nil
}

func (s *PostgresStore) RecentGames(limit int) ([]Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	q := `SELECT g.id, g.played_at, g.title_id, t.name, g.participant_ids, g.winner_ids, g.notes, g.is_active
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
		return nil, fmt.Errorf("RecentGames query: %w", err)
	}
	defer rows.Close()

	out := make([]Game, 0, max(0, limit))
	for rows.Next() {
		var g Game
		if err := rows.Scan(&g.ID, &g.PlayedAt, &g.TitleID, &g.Title, &g.ParticipantIDs, &g.WinnerIDs, &g.Notes, &g.IsActive); err != nil {
			return nil, fmt.Errorf("RecentGames scan: %w", err)
		}
		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("RecentGames rows: %w", err)
	}
	return out, nil
}
func (s *PostgresStore) GamesByWeek(year, week int) ([]Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	q := `SELECT g.id, g.played_at, g.title_id, t.name, g.participant_ids, g.winner_ids, g.notes, g.is_active
		  FROM app.games g
		  JOIN app.titles t ON t.id = g.title_id
		 WHERE EXTRACT(YEAR FROM g.played_at) = $1 AND EXTRACT(WEEK FROM g.played_at) = $2
		 ORDER BY g.played_at, g.id`

	rows, err := s.db.Query(ctx, q, year, week)
	if err != nil {
		return nil, fmt.Errorf("GamesByWeek query: %w", err)
	}
	defer rows.Close()

	out := make([]Game, 0, 100)
	for rows.Next() {
		var g Game
		if err := rows.Scan(&g.ID, &g.PlayedAt, &g.TitleID, &g.Title, &g.ParticipantIDs, &g.WinnerIDs, &g.Notes, &g.IsActive); err != nil {
			return nil, fmt.Errorf("GamesByWeek scan: %w", err)
		}
		out = append(out, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GamesByWeek rows: %w", err)
	}

	return out, nil
}

func (s *PostgresStore) GamesByYear(year int) ([]Game, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	q := `SELECT g.id, g.played_at, g.title_id, t.name, g.participant_ids, g.winner_ids, g.notes, g.is_active
		  FROM app.games g
		  JOIN app.titles t ON t.id = g.title_id
		 WHERE EXTRACT(YEAR FROM g.played_at) = $1
		 ORDER BY g.played_at, g.id`

	rows, err := s.db.Query(ctx, q, year)
	if err != nil {
		return nil, fmt.Errorf("GamesByYear query: %w", err)
	}
	defer rows.Close()

	out := make([]Game, 0, 100)
	for rows.Next() {
		var g Game
		if err := rows.Scan(&g.ID, &g.PlayedAt, &g.TitleID, &g.Title, &g.ParticipantIDs, &g.WinnerIDs, &g.Notes, &g.IsActive); err != nil {
			return nil, fmt.Errorf("GamesByYear scan: %w", err)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GamesByYear rows: %w", err)
	}

	return out, nil
}

// ============================
// Players
// ============================

func (s *PostgresStore) ListPlayers() ([]Player, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.db.Query(ctx, `SELECT id, name, is_active FROM app.players ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("ListPlayers: %w", err)
	}
	defer rows.Close()

	var out []Player
	for rows.Next() {
		var p Player
		if err := rows.Scan(&p.ID, &p.Name, &p.IsActive); err != nil {
			return nil, fmt.Errorf("ListPlayers scan: %w", err)
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListPlayers rows: %w", err)
	}

	return out, nil
}

func (s *PostgresStore) AddPlayer(name string) (Player, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var p Player
	err := s.db.QueryRow(ctx, `INSERT INTO app.players (name) VALUES ($1) RETURNING id, name, is_active`, name).Scan(&p.ID, &p.Name, &p.IsActive)
	if err != nil {
		return Player{}, fmt.Errorf("AddPlayer: %w", err)
	}

	return p, nil
}

func (s *PostgresStore) UpdatePlayer(id int64, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.Exec(ctx, `UPDATE app.players SET name=$2 WHERE id=$1`, id, name)
	if err != nil {
		return fmt.Errorf("UpdatePlayer: %w", err)
	}

	return nil
}

func (s *PostgresStore) SetPlayerActive(id int64, active bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.Exec(ctx, `UPDATE app.players SET is_active = $2 WHERE id = $1`, id, active)
	if err != nil {
		return fmt.Errorf("SetPlayerActive: %w", err)
	}

	return nil
}

func (s *PostgresStore) DeletePlayer(id int64) error {
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
		return fmt.Errorf("DeletePlayer: %w", err)
	}
	if exists {
		return fmt.Errorf("player is referenced by a game")
	}

	_, err = s.db.Exec(ctx, `DELETE FROM app.players WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("DeletePlayer: %w", err)
	}

	return nil
}

// ============================
// Titles
// ============================

func (s *PostgresStore) ListTitles() ([]Title, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	rows, err := s.db.Query(ctx, `SELECT id, name, is_active FROM app.titles ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("ListTitles: %w", err)
	}
	defer rows.Close()

	var out []Title
	for rows.Next() {
		var t Title
		if err := rows.Scan(&t.ID, &t.Name, &t.IsActive); err != nil {
			return nil, fmt.Errorf("ListTitles scan: %w", err)
		}
		out = append(out, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ListTitles rows: %w", err)
	}

	return out, nil
}

func (s *PostgresStore) AddTitle(name string) (Title, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var t Title
	err := s.db.QueryRow(ctx, `INSERT INTO app.titles (name) VALUES ($1) RETURNING id, name, is_active`, name).Scan(&t.ID, &t.Name, &t.IsActive)
	if err != nil {
		return Title{}, fmt.Errorf("AddTitle: %w", err)
	}

	return t, nil
}

func (s *PostgresStore) UpdateTitle(id int64, name string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.Exec(ctx, `UPDATE app.titles SET name=$2 WHERE id=$1`, id, name)
	if err != nil {
		return fmt.Errorf("UpdateTitle: %w", err)
	}

	return nil
}

func (s *PostgresStore) SetTitleActive(id int64, active bool) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.Exec(ctx, `UPDATE app.titles SET is_active = $2 WHERE id = $1`, id, active)
	if err != nil {
		return fmt.Errorf("SetTitleActive: %w", err)
	}

	return nil
}

func (s *PostgresStore) DeleteTitle(id int64) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := s.db.Exec(ctx, `DELETE FROM app.titles WHERE id=$1`, id)
	if err != nil {
		return fmt.Errorf("DeleteTitle: %w", err)
	}

	return nil
}

// ============================
// Tiebreakers
// ============================

func (s *PostgresStore) GetTiebreaker(scope, scopeKey string) (Tiebreaker, bool, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var raw []byte
	err := s.db.QueryRow(ctx,
		`SELECT data FROM app.tiebreakers WHERE scope = $1 AND scope_key = $2`,
		scope, scopeKey,
	).Scan(&raw)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Tiebreaker{}, false, nil
		}
		return Tiebreaker{}, false, fmt.Errorf("GetTiebreaker: %w", err)
	}

	var tb Tiebreaker
	if err := json.Unmarshal(raw, &tb); err != nil {
		return Tiebreaker{}, false, fmt.Errorf("GetTiebreaker unmarshal: %w", err)
	}
	return tb, true, nil
}

func (s *PostgresStore) SetTiebreaker(tb Tiebreaker) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	b, err := json.Marshal(tb)
	if err != nil {
		return fmt.Errorf("SetTiebreaker marshal: %w", err)
	}

	_, err = s.db.Exec(ctx,
		`INSERT INTO app.tiebreakers (scope, scope_key, data)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (scope, scope_key)
		 DO UPDATE SET data = EXCLUDED.data`,
		tb.Scope, tb.ScopeKey, b,
	)

	if err != nil {
		return fmt.Errorf("SetTiebreaker: %w", err)
	}

	return nil
}
