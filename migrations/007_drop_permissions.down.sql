CREATE TABLE permissions (
    uuid        UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name_action VARCHAR(100) NOT NULL UNIQUE
);

CREATE TABLE rol_permissions (
    rol_id         UUID NOT NULL REFERENCES rol(uuid) ON DELETE CASCADE,
    permissions_id UUID NOT NULL REFERENCES permissions(uuid) ON DELETE CASCADE,
    PRIMARY KEY (rol_id, permissions_id)
);
