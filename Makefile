include migrations.mk

help:
	@echo "Usage:"
	@echo "  make migrate-up - Run all up migrations"
	@echo "  make migrate-down - Run all down migrations"