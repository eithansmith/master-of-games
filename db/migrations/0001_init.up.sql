-- Create application schema
CREATE SCHEMA IF NOT EXISTS app;

-- Auto-update updated_at
CREATE OR REPLACE FUNCTION app.set_updated_at()
    RETURNS TRIGGER AS
$$
BEGIN
    NEW.updated_at = now();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- players
CREATE TABLE IF NOT EXISTS app.players
(
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO app.players (name)
VALUES ('AFAILLA'),
       ('AMAAG'),
       ('BAIRD'),
       ('CNEUTZLING'),
       ('DSCHMITT'),
       ('ESMITH'),
       ('EZAMORA'),
       ('JWHITTEMORE'),
       ('LCOOK'),
       ('LGRAVOT'),
       ('LWOOTTEN'),
       ('RSTEUER'),
       ('RWALL'),
       ('SBLUE'),
       ('TRIEDER'),
       ('TSUMPTER'),
       ('TCOX');

-- titles
CREATE TABLE IF NOT EXISTS app.titles
(
    id         BIGSERIAL PRIMARY KEY,
    name       TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

INSERT INTO app.titles (name)
VALUES ('Bang'),
       ('Camel Up'),
       ('Cockroach Poker'),
       ('Coup'),
       ('Dice Forge'),
       ('Don''t LLAMA'),
       ('Flip 7'),
       ('King of New York'),
       ('King of Tokyo'),
       ('Martian Dice'),
       ('Steampunk Rally'),
       ('Strike Dice'),
       ('Take 5'),
       ('Zombie Dice');

-- games
CREATE TABLE IF NOT EXISTS app.games
(
    played_at       timestamp with time zone                        not null,
    created_at      timestamp with time zone default now()          not null,
    title_id        bigint                                          not null
        constraint fk_games_title_id
            references app.titles,
    participant_ids bigint[]                 default '{}'::bigint[] not null,
    winner_ids      bigint[]                 default '{}'::bigint[] not null,
    notes           text                     default ''::text       not null
);

CREATE INDEX IF NOT EXISTS idx_games_played_at
    ON app.games (played_at DESC);

-- tiebreakers
CREATE TABLE IF NOT EXISTS app.tiebreakers
(
    scope      text                                   not null,
    scope_key  text                                   not null,
    data       jsonb                                  not null,
    created_at timestamp with time zone default now() not null,
    updated_at timestamp with time zone default now() not null,
    primary key (scope, scope_key)
);

CREATE TRIGGER trg_tiebreakers_updated_at
    BEFORE UPDATE
    ON app.tiebreakers
    FOR EACH ROW
EXECUTE PROCEDURE app.set_updated_at();
