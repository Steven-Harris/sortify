# Sortify Project Management
# Justfile for managing development tasks
# Default recipe
default:
    @just --list

# Backend commands
backend-dev:
    #!/usr/bin/env bash
    cd backend
    echo "🚀 Starting backend development server..."
    go run cmd/server/main.go

backend-build:
    #!/usr/bin/env bash
    cd backend
    echo "🔨 Building backend..."
    go build -o bin/sortify cmd/server/main.go

backend-test:
    #!/usr/bin/env bash
    cd backend
    echo "🧪 Running backend tests..."
    go test -v ./...

backend-lint:
    #!/usr/bin/env bash
    cd backend
    echo "🔍 Linting backend code..."
    golangci-lint run

backend-tidy:
    #!/usr/bin/env bash
    cd backend
    echo "📦 Tidying Go modules..."
    go mod tidy

backend-update:
    #!/usr/bin/env bash
    cd backend
    echo "🔄 Updating Go modules..."
    go get -u ./...
    echo "🧹 Cleaning up Go modules..." 
    go mod tidy

backend-download:
    #!/usr/bin/env bash
    go mod download

# Frontend commands
frontend-dev:
    #!/usr/bin/env bash
    cd frontend 
    echo "🎨 Starting frontend development server..."
    pnpm run dev

frontend-build:
    #!/usr/bin/env bash
    cd frontend 
    echo "🔨 Building frontend..."
    pnpm run build

frontend-preview:
    #!/usr/bin/env bash
    cd frontend 
    echo "👀 Starting frontend preview..."
    pnpm run preview

frontend-install:
    #!/usr/bin/env bash
    cd frontend 
    echo "📦 Installing frontend dependencies..."
    pnpm install

frontend-test:
    #!/usr/bin/env bash
    cd frontend 
    echo "🧪 Running frontend tests..."
    pnpm test

# Full stack commands
dev:
    #!/usr/bin/env bash
    echo "🚀 Starting full development environment..."
    echo "Backend will run on :8080, Frontend will run on :5173"
    # Run both backend and frontend concurrently
    just backend-dev &
    BACKEND_PID=$!
    sleep 2
    just frontend-dev &
    FRONTEND_PID=$!
    
    # Wait for Ctrl+C
    trap "kill $BACKEND_PID $FRONTEND_PID 2>/dev/null" EXIT
    wait

build: backend-build frontend-build

test: backend-test frontend-test

install: backend-download frontend-install

# Docker commands
docker-build:
    #!/usr/bin/env bash
    echo "🐳 Building Docker image..."
    docker build -t sortify:latest -f docker/Dockerfile .

docker-run:
    #!/usr/bin/env bash
    echo "🐳 Running Docker container..."
    docker run -p 8080:8080 -v $(pwd)/media:/media sortify:latest

docker-dev:
    #!/usr/bin/env bash
    echo "🐳 Building and running Docker container..."
    just docker-build
    just docker-run

# Database commands
db-create:
    #!/usr/bin/env bash
    cd backend
    echo "🗄️  Creating database directories..."
    mkdir -p data

db-reset:
    #!/usr/bin/env bash
    cd backend
    echo "🗄️  Resetting database..."
    rm -f data/sortify.db
    just db-create

# Media management
media-setup:
    #!/usr/bin/env bash
    echo "📁 Setting up media directories..."
    mkdir -p media
    mkdir -p media/temp
    chmod 755 media
    chmod 755 media/temp

media-clean:
    #!/usr/bin/env bash
    echo "🧹 Cleaning temporary media files..."
    rm -rf media/temp/*

media-reset:
    #!/usr/bin/env bash
    echo "🧹 Resetting all media (WARNING: This will delete all uploaded files)..."
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        rm -rf media/*
        just media-setup
        echo "✅ Media directories reset"
    else
        echo "❌ Operation cancelled"
    fi

# Development utilities
clean:
    #!/usr/bin/env bash
    echo "🧹 Cleaning build artifacts..."
    cd backend && rm -rf bin/
    rm -rf dist/
    rm -rf node_modules/.cache/
    just media-clean

setup:
    #!/usr/bin/env bash
    echo "⚙️  Setting up Sortify development environment..."
    echo "📦 Installing dependencies..."
    just install
    echo "📁 Setting up directories..."
    just media-setup
    just db-create
    echo "✅ Setup complete!"
    echo ""
    echo "🚀 To start development:"
    echo "   just dev        # Start both backend and frontend"
    echo "   just backend-dev # Start only backend"
    echo "   just frontend-dev # Start only frontend"

# API testing
api-health:
    #!/usr/bin/env bash
    echo "🏥 Testing API health endpoint..."
    curl -s http://localhost:8080/api/health | jq .

api-test-upload:
    #!/usr/bin/env bash
    echo "📤 Testing upload endpoint..."
    echo "Creating test upload session..."
    curl -X POST http://localhost:8080/api/upload/start \
        -H "Content-Type: application/json" \
        -d '{"filename":"test.jpg","file_size":1024,"chunk_size":256,"checksum":"abc123"}' \
        | jq .

# Git helpers
git-status:
    #!/usr/bin/env bash
    echo "📊 Git status..."
    git status

git-setup:
    #!/usr/bin/env bash
    echo "⚙️  Setting up Git hooks and configuration..."
    # Add any git setup commands here
    echo "✅ Git setup complete"

# Deployment
deploy-staging:
    #!/usr/bin/env bash
    echo "🚀 Deploying to staging..."
    just build
    just docker-build
    echo "✅ Ready for staging deployment"

deploy-production:
    #!/usr/bin/env bash
    echo "🚀 Deploying to production..."
    echo "⚠️  This should be done via GitHub Actions"
    echo "   Push to main branch to trigger deployment"

# Utilities
logs:
    #!/usr/bin/env bash
    echo "📋 Showing recent logs..."
    if [ -f "backend/logs/sortify.log" ]; then
        tail -f backend/logs/sortify.log
    else
        echo "No log file found. Start the backend to generate logs."
    fi

version:
    #!/usr/bin/env bash
    echo "📋 Sortify version information:"
    echo "Project: Sortify v1.0.0"
    echo "Go version: $(cd backend && go version)"
    echo "Node version: $(node --version)"
    echo "Docker version: $(docker --version)"

help:
    #!/usr/bin/env bash
    echo "🆘 Sortify Development Help"
    echo ""
    echo "🚀 Quick Start:"
    echo "   just setup      # First time setup"
    echo "   just dev        # Start development environment"
    echo ""
    echo "📋 Available commands:"
    just --list
