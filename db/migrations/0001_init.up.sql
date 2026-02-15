-- Create application schema
CREATE SCHEMA IF NOT EXISTS app;

-- ============================
-- Games
-- ============================

CREATE TABLE IF NOT EXISTS app.games
(
    id         BIGSERIAL PRIMARY KEY,

    title      TEXT        NOT NULL,
    played_at  TIMESTAMPTZ NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_games_played_at
    ON app.games (played_at DESC);

-- ============================
-- Tiebreakers
-- ============================

CREATE TABLE IF NOT EXISTS app.tiebreakers
(
    scope      TEXT        NOT NULL,
    scope_key  TEXT        NOT NULL,

    -- store the entire struct as JSON
    data       JSONB       NOT NULL,

    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    PRIMARY KEY (scope, scope_key)
);

-- Auto-update updated_at
CREATE OR REPLACE FUNCTION app.set_updated_at()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_tiebreakers_updated_at ON app.tiebreakers;

CREATE TRIGGER trg_tiebreakers_updated_at
    BEFORE UPDATE
    ON app.tiebreakers
    FOR EACH ROW
EXECUTE FUNCTION app.set_updated_at();
