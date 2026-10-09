include migrations.mk

help:
	@echo "Usage:"
	@echo "  make migrate-up       - Base vacía: crea el esquema; si no, aplica las pendientes"
	@echo "  make migrate-down     - Revierte la última migración aplicada"
	@echo "  make migrate-status   - Muestra qué migraciones están aplicadas"
	@echo "  make migrate-baseline - Marca todas como aplicadas (bases creadas a mano)"
	@echo "  make migrate-seed     - Inserta los datos de ejemplo (seed.sql)"
