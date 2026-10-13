-- Full, current schema of backend-omnibus (equivalent to applying 001..NNN).
-- `go run ./cmd/migrate up` uses it to initialize an empty database.
-- When adding a new migration, reflect the change here too.

CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE IF NOT EXISTS city (
    postal_code VARCHAR(10)  PRIMARY KEY,
    name        VARCHAR(255) NOT NULL
);

CREATE TABLE IF NOT EXISTS bus_terminal (
    uuid                 UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    postal_code          VARCHAR(10)  NOT NULL REFERENCES city(postal_code) ON UPDATE CASCADE,
    name                 VARCHAR(255) NOT NULL,
    external_terminal_id UUID NULL
);

CREATE INDEX IF NOT EXISTS idx_bus_terminal_postal_code ON bus_terminal(postal_code);
CREATE UNIQUE INDEX IF NOT EXISTS idx_bus_terminal_external_terminal_id
    ON bus_terminal (external_terminal_id)
    WHERE external_terminal_id IS NOT NULL;

CREATE TABLE IF NOT EXISTS platform (
    code            SERIAL PRIMARY KEY,
    bay             VARCHAR(50) NOT NULL,
    coordinates     JSONB,
    bus_terminal_id UUID NOT NULL REFERENCES bus_terminal(uuid)
);

CREATE INDEX IF NOT EXISTS idx_platform_bus_terminal_id ON platform(bus_terminal_id);

CREATE TABLE IF NOT EXISTS role (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS users (
    uuid       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(255) NOT NULL,
    last_name  VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL,
    role_id    UUID         NOT NULL REFERENCES role(uuid),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_email   ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_role_id ON users(role_id);

CREATE TABLE IF NOT EXISTS user_refresh_tokens (
    uuid        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL UNIQUE REFERENCES users(uuid) ON DELETE CASCADE,
    token       VARCHAR(255) NOT NULL,
    expiry_date TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_user_refresh_tokens_user_id ON user_refresh_tokens(user_id);

CREATE TABLE IF NOT EXISTS user_terminal (
    user_id         UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    bus_terminal_id UUID NOT NULL REFERENCES bus_terminal(uuid) ON DELETE CASCADE,
    PRIMARY KEY (user_id, bus_terminal_id)
);

CREATE INDEX IF NOT EXISTS idx_user_terminal_user_id         ON user_terminal(user_id);
CREATE INDEX IF NOT EXISTS idx_user_terminal_bus_terminal_id ON user_terminal(bus_terminal_id);

CREATE TABLE IF NOT EXISTS awaited_trip (
    user_id         UUID PRIMARY KEY REFERENCES users(uuid) ON DELETE CASCADE,
    group_key       VARCHAR(255) NOT NULL,
    ticket          VARCHAR(100) NOT NULL,
    bus_terminal_id UUID         NOT NULL REFERENCES bus_terminal(uuid) ON DELETE CASCADE,
    created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    notified_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_awaited_trip_group_key ON awaited_trip(group_key);

CREATE TABLE IF NOT EXISTS notifications (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    group_key  TEXT,
    group_name TEXT        NOT NULL,
    expiration TIMESTAMPTZ NOT NULL,
    payload    JSONB       NOT NULL,
    date       TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Fixed roles: the code references them by name.
INSERT INTO role (uuid, name) VALUES
    ('b0000000-0000-0000-0000-000000000001', 'user'),
    ('b0000000-0000-0000-0000-000000000002', 'admin'),
    ('b0000000-0000-0000-0000-000000000003', 'super_admin')
ON CONFLICT DO NOTHING;
