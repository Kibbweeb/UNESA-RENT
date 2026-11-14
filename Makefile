.PHONY: install backend frontend dev db-up db-down db-reset

install:
	cd backend && go mod download
	cd frontend && npm install

backend:
	cd backend && go run ./cmd/server

frontend:
	cd frontend && npm run dev

dev:
	@trap 'kill 0' EXIT INT TERM; \
	(cd backend && go run ./cmd/server) & \
	(cd frontend && npm run dev) & \
	wait

db-up:
	cd postgres && docker compose up -d

db-down:
	cd postgres && docker compose down

db-reset:
	cd postgres && docker compose down -v && docker compose up -d
