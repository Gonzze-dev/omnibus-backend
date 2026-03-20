DATABASE_URL := postgres://postgres:1234@localhost:5432/omnibus-terminal

migrate-up:
	psql $(DATABASE_URL) -f ./migrations/001_create_tables.up.sql
	psql $(DATABASE_URL) -f ./migrations/002_auth_roles.up.sql
	psql $(DATABASE_URL) -f ./migrations/002_seed.sql