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

backend-air:
    #!/usr/bin/env bash
    cd backend
    echo "🔥 Starting backend with Air live reload..."
    ~/go/bin/air

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

install: frontend-install

# Docker commands
docker-build:
    #!/usr/bin/env bash
    echo "🐳 Building Docker image..."
    # Use docker build (which uses buildx under the hood in modern Docker)
    docker build -t sortify:latest .

docker-run:
    #!/usr/bin/env bash
    echo "🐳 Running Docker container..."
    docker run -p 8080:8080 -v $(pwd)/media:/media sortify:latest

docker-dev:
    #!/usr/bin/env bash
    echo "🐳 Building and running Docker container..."
    just docker-build
    just docker-run

docker-push REGISTRY IMAGE_NAME:
    #!/usr/bin/env bash
    echo "🐳 Building and pushing Docker image..."
    # Try buildx for multi-platform, fallback to regular build + push
    if docker buildx version >/dev/null 2>&1; then
        echo "Using buildx for multi-platform build and push..."
        docker buildx build --platform linux/amd64,linux/arm64 \
            -t {{REGISTRY}}/{{IMAGE_NAME}}:latest \
            --push .
    else
        echo "Using regular build and push..."
        docker build -t {{REGISTRY}}/{{IMAGE_NAME}}:latest .
        docker push {{REGISTRY}}/{{IMAGE_NAME}}:latest
    fi

# Development utilities
clean:
    #!/usr/bin/env bash
    echo "🧹 Cleaning build artifacts..."
    cd backend && rm -rf bin/
    rm -rf dist/
    rm -rf node_modules/.cache/

setup:
    #!/usr/bin/env bash
    echo "⚙️  Setting up Sortify development environment..."
    echo "📦 Installing dependencies..."
    just install
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
        -d '{"filename":"test.jpg","fileSize":1024,"chunk_size":256,"checksum":"abc123"}' \
        | jq .

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

docker-compose-up:
    #!/usr/bin/env bash
    echo "🐳 Starting services with docker-compose..."
    docker-compose up -d

docker-compose-down:
    #!/usr/bin/env bash
    echo "🐳 Stopping services with docker-compose..."
    docker-compose down

docker-compose-logs:
    #!/usr/bin/env bash
    echo "📋 Showing docker-compose logs..."
    docker-compose logs -f

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
