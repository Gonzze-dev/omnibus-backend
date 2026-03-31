ALTER TABLE bus_terminal
    ADD COLUMN external_terminal_id UUID NULL;

CREATE UNIQUE INDEX idx_bus_terminal_external_terminal_id
    ON bus_terminal (external_terminal_id)
    WHERE external_terminal_id IS NOT NULL;
