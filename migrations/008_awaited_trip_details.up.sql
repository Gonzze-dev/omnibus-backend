-- Previous rows only have group_key: without the ticket the passenger's
-- wait cannot be rebuilt, so they are discarded.
DELETE FROM awaited_trip;

ALTER TABLE awaited_trip
    ADD COLUMN ticket          VARCHAR(100) NOT NULL,
    ADD COLUMN bus_terminal_id UUID         NOT NULL REFERENCES bus_terminal(uuid) ON DELETE CASCADE,
    ADD COLUMN created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    ADD COLUMN notified_at     TIMESTAMPTZ;

CREATE INDEX idx_awaited_trip_group_key ON awaited_trip(group_key);
