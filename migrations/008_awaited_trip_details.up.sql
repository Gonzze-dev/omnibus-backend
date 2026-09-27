-- Las filas previas solo tienen group_key: sin el ticket no se puede
-- reconstruir la espera del pasajero, así que se descartan.
DELETE FROM awaited_trip;

ALTER TABLE awaited_trip
    ADD COLUMN ticket          VARCHAR(100) NOT NULL,
    ADD COLUMN bus_terminal_id UUID         NOT NULL REFERENCES bus_terminal(uuid) ON DELETE CASCADE,
    ADD COLUMN created_at      TIMESTAMPTZ  NOT NULL DEFAULT now(),
    ADD COLUMN notified_at     TIMESTAMPTZ;

CREATE INDEX idx_awaited_trip_group_key ON awaited_trip(group_key);
