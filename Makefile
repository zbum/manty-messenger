.PHONY: run build test migrate frontend-dev frontend-build clean all

# Go commands
run:
	go run cmd/server/main.go

build:
	go build -o bin/server cmd/server/main.go

test:
	go test ./...

# Database
migrate:
	mysql -u $(DB_USER) -p$(DB_PASS) $(DB_NAME) < internal/database/migrations/001_init.sql

# Frontend commands
frontend-dev:
	cd frontend && npm run dev

frontend-build:
	cd frontend && npm run build

frontend-install:
	cd frontend && npm install

# Combined commands
dev:
	@echo "Starting backend server..."
	@go run cmd/server/main.go &
	@echo "Starting frontend dev server..."
	@cd frontend && npm run dev

all: frontend-build build

clean:
	rm -rf bin/
	rm -rf frontend/dist/
	rm -rf frontend/node_modules/
