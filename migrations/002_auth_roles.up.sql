CREATE TABLE rol (
    uuid UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name VARCHAR(50) NOT NULL UNIQUE
);

CREATE TABLE permissions (
    uuid        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name_action VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE rol_permissions (
    rol_id         UUID NOT NULL REFERENCES rol(uuid) ON DELETE CASCADE,
    permissions_id UUID NOT NULL REFERENCES permissions(uuid) ON DELETE CASCADE,
    PRIMARY KEY (rol_id, permissions_id)
);

CREATE TABLE users (
    uuid       UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    first_name VARCHAR(255) NOT NULL,
    last_name  VARCHAR(255) NOT NULL,
    email      VARCHAR(255) NOT NULL UNIQUE,
    password   VARCHAR(255) NOT NULL,
    dni        VARCHAR(20)  NOT NULL UNIQUE,
    rol_id     UUID         NOT NULL REFERENCES rol(uuid),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_email  ON users(email);
CREATE INDEX idx_users_dni    ON users(dni);
CREATE INDEX idx_users_rol_id ON users(rol_id);

CREATE TABLE user_refresh_tokens (
    uuid        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id     UUID NOT NULL UNIQUE REFERENCES users(uuid) ON DELETE CASCADE,
    token       VARCHAR(255) NOT NULL,
    expiry_date TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at  TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_user_refresh_tokens_user_id ON user_refresh_tokens(user_id);

CREATE TABLE user_terminal (
    user_id         UUID NOT NULL REFERENCES users(uuid) ON DELETE CASCADE,
    bus_terminal_id UUID NOT NULL REFERENCES bus_terminal(uuid) ON DELETE CASCADE,
    PRIMARY KEY (user_id, bus_terminal_id)
);

CREATE INDEX idx_user_terminal_user_id         ON user_terminal(user_id);
CREATE INDEX idx_user_terminal_bus_terminal_id ON user_terminal(bus_terminal_id);
