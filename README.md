# Master of Games

A lightweight web app for tracking lunchtime board game results. Log games, track weekly and yearly standings, and settle ties with a tiebreaker.

## Features

- **Game log** — Record games with title, date/time, participants, winners, and notes. Weekday games only (Mon – Fri).
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

| Variable         | Default    | Notes                                                          |
|------------------|------------|-----------------------------------------------------------------|
| `DATABASE_URL`   | (required) | PostgreSQL connection string                                    |
| `PORT`           | `8080`     | Listen port                                                      |
| `BOOTSTRAP_USER` | (optional) | Username to create on first run, if no users exist yet          |
| `BOOTSTRAP_PASS` | (optional) | Password for the bootstrap user (hashed with bcrypt before storage) |

Auth is a single-user login backed by `app.users`/`app.sessions` (see [Architecture](CLAUDE.md)). If no user exists yet, login always fails until one is created — either via `BOOTSTRAP_USER`/`BOOTSTRAP_PASS` on startup, or by inserting a row directly.

### Run

```bash
DATABASE_URL=postgres://... BOOTSTRAP_USER=admin BOOTSTRAP_PASS=secret go run ./cmd/server
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

## Deployment

The app runs on [Render](https://render.com). Render's free tier spins the service down after a period of
inactivity, so [`.github/workflows/keep-warm.yml`](.github/workflows/keep-warm.yml) pings `/healthz` every 10 minutes
during a window around lunch (weekdays only) to keep it warm for when scores actually get logged, without pinging
around the clock.

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

| Method | Path                            | Description                        |
|--------|---------------------------------|------------------------------------|
| GET    | `/login`                        | Login page                         |
| POST   | `/login`                        | Submit credentials, start a session |
| POST   | `/logout`                       | End the session                    |
| GET    | `/`                             | Home — log a game, recent games    |
| POST   | `/games`                        | Add a game                         |
| POST   | `/games/{id}/toggle`            | Activate / deactivate a game       |
| POST   | `/games/{id}/delete`            | Deactivate a game                  |
| GET    | `/weeks/current`                | Redirect to current ISO week       |
| GET    | `/weeks/{year}/{week}`          | Weekly standings                   |
| POST   | `/weeks/{year}/{week}/tiebreak` | Set weekly tiebreaker              |
| GET    | `/years/{year}`                 | Yearly standings                   |
| POST   | `/years/{year}/tiebreak`        | Set yearly tiebreaker              |
| GET    | `/years/{year}/race`            | Year race page                     |
| GET    | `/years/{year}/race/chart`      | Year race SVG chart (HTMX partial) |
| GET    | `/players`                      | Players list                       |
| POST   | `/players`                      | Add a player                       |
| POST   | `/players/{id}/update`          | Rename a player                    |
| POST   | `/players/{id}/toggle`          | Activate / deactivate a player     |
| POST   | `/players/{id}/delete`          | Deactivate a player                |
| GET    | `/titles`                       | Titles list                        |
| POST   | `/titles`                       | Add a title                        |
| POST   | `/titles/{id}/update`           | Rename a title                     |
| POST   | `/titles/{id}/toggle`           | Activate / deactivate a title      |
| POST   | `/titles/{id}/delete`           | Deactivate a title                 |
| GET    | `/healthz`                      | Health check (no auth required)    |