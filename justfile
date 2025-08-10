# Sortify Project Commands

# Show available commands
default:
    @just --list

# Development
dev-backend:
    cd backend && go run cmd/server/main.go

dev-frontend:
    cd frontend && pnpm run dev

dev:
    #!/usr/bin/env bash
    echo "� Starting development environment..."
    just dev-backend &
    sleep 2
    just dev-frontend &
    wait

# Building
build-frontend:
    cd frontend && pnpm run build

build-backend:
    cd backend && go build -o bin/sortify cmd/server/main.go

build: build-frontend build-backend

# Testing
test-backend:
    cd backend && go test -v ./...

test-frontend:
    cd frontend && pnpm run build

test: test-backend test-frontend

# Docker
docker-build:
    docker build -t sortify .

docker-run:
    docker compose up -d

docker-stop:
    docker compose down

docker-logs:
    docker compose logs -f

# Setup
install:
    cd frontend && pnpm install

setup: install
    @echo "✅ Setup complete! Run 'just dev' to start development"

# Utilities
clean:
    rm -rf backend/bin/ frontend/dist/ frontend/node_modules/.cache/

health:
    curl -s http://localhost:8080/api/health
