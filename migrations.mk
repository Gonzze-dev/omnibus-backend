# Usa las mismas variables DB_* (.env) que la API.
MIGRATE := go run ./cmd/migrate

migrate-up:
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down

migrate-status:
	$(MIGRATE) status

migrate-baseline:
	$(MIGRATE) baseline

migrate-seed:
	$(MIGRATE) seed
