include migrations.mk

help:
	@echo "Usage:"
	@echo "  make migrate-up       - Empty database: creates the schema; otherwise applies pending ones"
	@echo "  make migrate-down     - Reverts the last applied migration"
	@echo "  make migrate-status   - Shows which migrations are applied"
	@echo "  make migrate-baseline - Marks all as applied (manually created databases)"
	@echo "  make migrate-seed     - Inserts the sample data (seed.sql)"
	@echo "  make migrate-seed-ar  - Loads cities and terminals of Argentina (migrations/data/*.json)"
	@echo "  make migrate-seed-platforms - Loads the bays of TERMINAL DE GUALEGUAYCHÚ (after seed-ar)"
