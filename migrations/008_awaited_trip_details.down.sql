DROP INDEX IF EXISTS idx_awaited_trip_group_key;

ALTER TABLE awaited_trip
    DROP COLUMN IF EXISTS notified_at,
    DROP COLUMN IF EXISTS created_at,
    DROP COLUMN IF EXISTS bus_terminal_id,
    DROP COLUMN IF EXISTS ticket;
