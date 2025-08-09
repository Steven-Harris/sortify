# Multi-stage Dockerfile for Sortify
# Stage 1: Build the frontend
FROM node:alpine AS frontend-builder

WORKDIR /app

# Copy package files
COPY frontend/package.json frontend/pnpm-lock.yaml ./

# Install pnpm and dependencies
RUN npm install -g pnpm
RUN pnpm install --frozen-lockfile

# Copy frontend source code
COPY frontend/ .

# Build the frontend
RUN pnpm build

# Stage 2: Build the Go backend
FROM golang:alpine AS go-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata wget

# Copy go mod files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source code
COPY backend/ ./

# Create a non-root user and group
RUN addgroup -g 1000 appgroup && adduser -u 1000 -G appgroup -D appuser

# Build the binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o sortify ./cmd/server

# Copy frontend build into backend for embedding
COPY --from=frontend-builder /app/dist ./web

# Create volume mount point
VOLUME ["/media"]

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/health || exit 1

# Switch to non-root user
USER appuser

# Run the application
ENTRYPOINT ["./sortify"]
