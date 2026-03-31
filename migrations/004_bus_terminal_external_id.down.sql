DROP INDEX IF EXISTS idx_bus_terminal_external_terminal_id;

ALTER TABLE bus_terminal
    DROP COLUMN IF EXISTS external_terminal_id;
