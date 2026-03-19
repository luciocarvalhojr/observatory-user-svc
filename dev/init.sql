-- user-svc schema — applied automatically by PostgreSQL on first start

CREATE TABLE IF NOT EXISTS users (
    id         UUID        PRIMARY KEY DEFAULT gen_random_uuid(),
    email      TEXT        UNIQUE NOT NULL,
    name       TEXT        NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
