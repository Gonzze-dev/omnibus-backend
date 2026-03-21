ALTER TABLE bus_terminal
    DROP CONSTRAINT bus_terminal_postal_code_fkey,
    ADD CONSTRAINT bus_terminal_postal_code_fkey
        FOREIGN KEY (postal_code) REFERENCES city(postal_code) ON UPDATE CASCADE;
