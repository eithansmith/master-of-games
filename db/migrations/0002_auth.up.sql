-- users: single-user table
CREATE TABLE IF NOT EXISTS app.users
(
    id            BIGSERIAL PRIMARY KEY,
    username      TEXT        NOT NULL UNIQUE,
    password_hash TEXT        NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TRIGGER trg_users_updated_at
    BEFORE UPDATE
    ON app.users
    FOR EACH ROW
EXECUTE PROCEDURE app.set_updated_at();

-- sessions: server-side session store, keyed by a hash of the cookie token
CREATE TABLE IF NOT EXISTS app.sessions
(
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT      NOT NULL REFERENCES app.users (id) ON DELETE CASCADE,
    token_hash TEXT        NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_sessions_token_hash ON app.sessions (token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_user_id ON app.sessions (user_id);
