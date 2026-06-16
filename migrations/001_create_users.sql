CREATE EXTENSION IF NOT EXISTS pgcrypto;

CREATE TABLE IF NOT EXISTS users (
    id             UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    name           TEXT        NOT NULL,
    email          TEXT        NOT NULL UNIQUE,
    password       TEXT        NOT NULL,
    role           TEXT        NOT NULL DEFAULT 'user',
    inactivated_at TIMESTAMPTZ,
    deleted_at     TIMESTAMPTZ,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    preferences    JSONB
);