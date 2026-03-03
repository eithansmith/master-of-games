# Master of Games

A lightweight web app for tracking lunchtime board game results. Log games, track weekly and yearly standings, and settle ties with a tiebreaker.

## Features

- **Game log** — Record games with title, date/time, participants, winners, and notes. Weekday games only (Mon–Fri).
- **Weekly standings** — Win counts per player for any ISO week, with tiebreaker support.
- **Yearly standings** — Qualifiers (top half by attendance) ranked by win rate, with tiebreaker support.
- **Year race chart** — SVG line chart of cumulative wins across the year.
- **Players & Titles management** — Add, rename, and activate/deactivate players and game titles.
- **Soft deletes** — Deactivating a game, player, or title sets `is_active = false`; data is never lost.
- **Toast notifications** — Non-intrusive feedback on every successful mutation (Toastify.js + HTMX triggers).

## Tech stack

- **Go** stdlib HTTP server — no web framework
- **PostgreSQL** (`pgx/v5`) — all tables under the `app` schema
- **HTMX** — partial page swaps; no full reloads on mutations
- **Alpine.js** — light client-side reactivity
- **Toastify.js** — toast notifications via `HX-Trigger` response headers

## Getting started

### Prerequisites

- Go 1.22+
- PostgreSQL

### Environment variables

| Variable          | Default    | Notes                        |
|-------------------|------------|------------------------------|
| `DATABASE_URL`    | (required) | PostgreSQL connection string |
| `PORT`            | `8080`     | Listen port                  |
| `BASIC_AUTH_USER` | (required) | HTTP Basic Auth username     |
| `BASIC_AUTH_PASS` | (required) | HTTP Basic Auth password     |

If `BASIC_AUTH_USER` or `BASIC_AUTH_PASS` are missing the server fails closed (returns 500 on all requests except `/healthz`).

### Run

```bash
DATABASE_URL=postgres://... BASIC_AUTH_USER=admin BASIC_AUTH_PASS=secret go run ./cmd/server
```

### Build

```bash
go build ./...
```

### Test

```bash
go test ./...
```

### Vet

```bash
go vet ./...
```

## Project structure

```
cmd/server/      Entry point — reads env, wires dependencies, registers routes
game/            Domain layer — models, standings logic, year race, store implementations
handlers/        HTTP layer — handlers, view models, renderer, store interface
db/              DB pool setup (pgxpool)
web/templates/   Go HTML templates (parsed at startup, not embedded)
web/static/      CSS and static assets
```

## Standings rules

**Weekly:** Winner = player with the most wins in the week. Ties resolved by a stored tiebreaker.

**Yearly:** Qualifiers = top half of players by days present (not game count). Winner = highest win rate (wins ÷ games played) among qualifiers. Ties resolved by a stored tiebreaker.

Tiebreakers are stored in `app.tiebreakers` as JSON keyed by `(scope, scope_key)` where scope is `"weekly"` or `"yearly"` and scope_key is `"YYYY-Www"` or `"YYYY"`.

## Routes

| Method | Path                          | Description                        |
|--------|-------------------------------|------------------------------------|
| GET    | `/`                           | Home — log a game, recent games    |
| POST   | `/games`                      | Add a game                         |
| POST   | `/games/{id}/toggle`          | Activate / deactivate a game       |
| POST   | `/games/{id}/delete`          | Deactivate a game                  |
| GET    | `/weeks/current`              | Redirect to current ISO week       |
| GET    | `/weeks/{year}/{week}`        | Weekly standings                   |
| POST   | `/weeks/{year}/{week}/tiebreak` | Set weekly tiebreaker             |
| GET    | `/years/{year}`               | Yearly standings                   |
| POST   | `/years/{year}/tiebreak`      | Set yearly tiebreaker              |
| GET    | `/years/{year}/race`          | Year race page                     |
| GET    | `/years/{year}/race/chart`    | Year race SVG chart (HTMX partial) |
| GET    | `/players`                    | Players list                       |
| POST   | `/players`                    | Add a player                       |
| POST   | `/players/{id}/update`        | Rename a player                    |
| POST   | `/players/{id}/toggle`        | Activate / deactivate a player     |
| POST   | `/players/{id}/delete`        | Deactivate a player                |
| GET    | `/titles`                     | Titles list                        |
| POST   | `/titles`                     | Add a title                        |
| POST   | `/titles/{id}/update`         | Rename a title                     |
| POST   | `/titles/{id}/toggle`         | Activate / deactivate a title      |
| POST   | `/titles/{id}/delete`         | Deactivate a title                 |
| GET    | `/healthz`                    | Health check (no auth required)    |