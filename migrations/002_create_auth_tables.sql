CREATE TABLE IF NOT EXISTS refresh_tokens (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID        NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT        NOT NULL UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at TIMESTAMPTZ NOT NULL,
    revoked    BOOLEAN     NOT NULL DEFAULT false,
    user_agent TEXT,
    ip_address TEXT
);