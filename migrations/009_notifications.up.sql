-- La tabla existía en las bases creadas a mano pero no tenía migración.
CREATE TABLE IF NOT EXISTS notifications (
    id         UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    group_key  TEXT,
    group_name TEXT        NOT NULL,
    expiration TIMESTAMPTZ NOT NULL,
    payload    JSONB       NOT NULL,
    date       TIMESTAMPTZ NOT NULL DEFAULT now()
);
