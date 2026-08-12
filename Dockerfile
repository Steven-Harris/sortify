# Multi-stage Dockerfile for Sortify

# Stage 1: Build the frontend
FROM node:alpine AS frontend-builder

WORKDIR /app

# Copy package files
COPY frontend/package.json frontend/pnpm-lock.yaml ./

# Install pnpm and dependencies with optimizations
RUN npm install -g pnpm@9.15.9 && \
    # Reduce Sharp installation time by using pre-built binaries
    pnpm config set sharp-libvips-binary-host "https://github.com/lovell/sharp-libvips/releases/download" && \
    # Install with optimizations
    pnpm install --frozen-lockfile

# Copy frontend source code
COPY frontend/ .

# Build the frontend
RUN pnpm build

# Stage 2: Build the Go backend and healthcheck utility
FROM golang:alpine AS go-builder

WORKDIR /app

# Install build dependencies in a single layer
RUN apk add --no-cache git ca-certificates tzdata && \
    # Pre-warm the module cache
    go env -w GOPROXY=https://proxy.golang.org,direct && \
    go env -w GOSUMDB=sum.golang.org

# Copy go mod files and download dependencies (cached layer)
COPY backend/go.mod backend/go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download && go mod verify

# Copy only the source code needed for building (exclude tests and other files)
COPY backend/cmd/ ./cmd/
COPY backend/internal/ ./internal/

# Build the main binary with optimizations
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -trimpath \
    -buildvcs=false \
    -o sortify ./cmd/server

# Build the healthcheck binary
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build \
    -ldflags='-w -s -extldflags "-static"' \
    -trimpath \
    -buildvcs=false \
    -o healthcheck ./cmd/healthcheck

# Copy frontend build into backend for embedding
COPY --from=frontend-builder /app/dist ./web

# Stage 3: Prepare files for scratch
FROM alpine:latest AS file-builder

# Create the media directory structure with world-writable permissions
RUN mkdir -p /tmp/media && \
    chmod -R 755 /tmp/media

# Create tmp directory for Go multipart form processing  
RUN mkdir -p /tmp/app-tmp && \
    chmod 755 /tmp/app-tmp

# Stage 4: Final scratch runtime image
FROM scratch

# Copy necessary files for scratch compatibility
COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# Copy the built binaries
COPY --from=go-builder /app/sortify /sortify
COPY --from=go-builder /app/healthcheck /healthcheck
COPY --from=frontend-builder /app/dist /web

# Copy media and tmp directories
COPY --from=file-builder /tmp/media /media
COPY --from=file-builder /tmp/app-tmp /tmp

# Create volume mount point
VOLUME ["/media"]

# Expose port
EXPOSE 8080

# Health check using our custom binary
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/healthcheck"]

# Run the application
ENTRYPOINT ["/sortify"]
