CREATE TABLE IF NOT EXISTS users (
    id         BIGSERIAL    PRIMARY KEY,
    name       VARCHAR(255) NOT NULL,
    email      VARCHAR(255) UNIQUE NOT NULL,
    password   VARCHAR(255) NOT NULL,
    created_at TIMESTAMPTZ  NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ  NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
