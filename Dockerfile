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

# Stage 2: Build the Go backend and healthcheck utility
FROM golang:alpine AS go-builder

WORKDIR /app

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY backend/go.mod backend/go.sum ./
RUN go mod download

# Copy source code
COPY backend/ ./

# Build the main binary with optimizations
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o sortify ./cmd/server

# Build the healthcheck binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags='-w -s -extldflags "-static"' \
    -a -installsuffix cgo \
    -o healthcheck ./cmd/healthcheck

# Copy frontend build into backend for embedding
COPY --from=frontend-builder /app/dist ./web

# Stage 3: Prepare files for scratch
FROM alpine:latest AS file-builder

# Create the media directory structure (without temp subdirectory)
RUN mkdir -p /tmp/media && \
    chmod -R 755 /tmp/media && \
    chown -R 1001:1001 /tmp/media

# Create tmp directory for Go multipart form processing
RUN mkdir -p /tmp/app-tmp && \
    chmod 755 /tmp/app-tmp && \
    chown 1001:1001 /tmp/app-tmp

# Create passwd/group files for scratch compatibility with user 1001
RUN echo 'sortify:x:1001:1001:sortify:/media:/sbin/nologin' > /tmp/passwd && \
    echo 'sortify:x:1001:' > /tmp/group

# Stage 4: Final scratch runtime image
FROM scratch

# Copy necessary files for scratch compatibility
COPY --from=go-builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=file-builder /tmp/passwd /etc/passwd
COPY --from=file-builder /tmp/group /etc/group

# Copy the built binaries
COPY --from=go-builder /app/sortify /sortify
COPY --from=go-builder /app/healthcheck /healthcheck
COPY --from=frontend-builder /app/dist /web

# Copy media directory with proper ownership
COPY --from=file-builder --chown=1001:1001 /tmp/media /media

# Copy tmp directory for both Go multipart processing and uploads
COPY --from=file-builder --chown=1001:1001 /tmp/app-tmp /tmp

# Switch to non-root user 1001
USER 1001:1001

# Create volume mount point
VOLUME ["/media"]

# Expose port
EXPOSE 8080

# Health check using our custom binary
HEALTHCHECK --interval=30s --timeout=10s --start-period=5s --retries=3 \
    CMD ["/healthcheck"]

# Run the application
ENTRYPOINT ["/sortify"]
