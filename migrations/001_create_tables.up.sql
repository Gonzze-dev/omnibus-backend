CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE city (
    postal_code VARCHAR(10) PRIMARY KEY,
    name        VARCHAR(255) NOT NULL
);

CREATE TABLE bus_terminal (
    uuid        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    postal_code VARCHAR(10) NOT NULL REFERENCES city(postal_code),
    name        VARCHAR(255) NOT NULL
);

CREATE INDEX idx_bus_terminal_postal_code ON bus_terminal(postal_code);

CREATE TABLE platform (
    code            SERIAL PRIMARY KEY,
    anden           VARCHAR(50) NOT NULL,
    coordinates     JSONB,
    bus_terminal_id UUID NOT NULL REFERENCES bus_terminal(uuid)
);

CREATE INDEX idx_platform_bus_terminal_id ON platform(bus_terminal_id);
