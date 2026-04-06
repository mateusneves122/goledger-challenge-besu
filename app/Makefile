.PHONY: run build tidy db-up db-down down

run:
	docker compose up -d

build:
	docker compose build

db-up:
	docker compose up -d postgres

db-down:
	docker compose stop postgres

down:
	docker compose down
