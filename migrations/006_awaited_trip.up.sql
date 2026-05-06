CREATE TABLE awaited_trip (
    user_id   UUID PRIMARY KEY REFERENCES users(uuid) ON DELETE CASCADE,
    group_key VARCHAR(255) NOT NULL
);
