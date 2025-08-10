# 📸 Sortify

> **Intelligent media file organizer with drag-and-drop web interface**

Sortify automatically organizes your photos and media files by date with a clean, modern web interface. Simply drag and drop your files, and Sortify will intelligently extract dates from filenames and EXIF data to organize them into a structured folder hierarchy.

![License](https://img.shields.io/github/license/Steven-harris/sortify)
![Docker Pulls](https://img.shields.io/docker/pulls/steven-harris/sortify)
![GitHub release](https://img.shields.io/github/v/release/Steven-harris/sortify)
![Build Status](https://img.shields.io/github/actions/workflow/status/Steven-harris/sortify/test.yml)

## ✨ Features

- 🖥️ **Modern Web Interface** - Responsive UI built with Lit web components
- 📁 **Smart Organization** - Automatic date-based folder structure (YYYY/Month)
- 🔄 **Chunked Uploads** - Handle large files with resume capability
- 🖼️ **Real-time Thumbnails** - Image previews during upload
- 🚨 **Duplicate Detection** - Prevents duplicate files with checksum verification
- 🛡️ **Security First** - Minimal scratch-based Docker image, non-root execution
- ⚙️ **Configurable** - Environment-based configuration for all settings
- 🌐 **Multi-platform** - Supports AMD64 and ARM64 architectures

## 🚀 Quick Start

### Docker Compose (Recommended)

1. **Download docker-compose.yml**:
```bash
curl -O https://raw.githubusercontent.com/Steven-harris/sortify/main/docker-compose.yml
```

2. **Start Sortify**:
```bash
docker compose up -d
```

3. **Open your browser**: http://localhost:8080

### Docker Run

```bash
docker run -d \
  --name sortify \
  -p 8080:8080 \
  -e MEDIA_PATH=/media \
  -e TEMP_PATH=/tmp \
  -v ./media:/media:rw \
  -v ./temp:/tmp:rw \
  steven-harris/sortify:latest
```

## 📋 Configuration

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MEDIA_PATH` | `/media` | Path for organized media files |
| `TEMP_PATH` | `/tmp` | Path for temporary upload files |
| `PORT` | `8080` | Server port |
| `LOG_LEVEL` | `info` | Log level (debug, info, warn, error) |
| `CORS_ORIGINS` | `*` | CORS origins (comma-separated) |

### Volume Mounts

| Host Path | Container Path | Purpose |
|-----------|----------------|---------|
| `./media` | `/media` | Persistent storage for organized files |
| `./temp` | `/tmp` | Temporary storage for upload processing |

## 🔧 Development

### Prerequisites

- [Go 1.22+](https://golang.org/dl/)
- [Node.js 18+](https://nodejs.org/)
- [pnpm](https://pnpm.io/)
- [just](https://github.com/casey/just) (optional, for task runner)

### Setup

```bash
# Clone the repository
git clone https://github.com/Steven-harris/sortify.git
cd sortify

# Install dependencies (if using just)
just setup

# Or manually
cd frontend && pnpm install
```

### Development Server

```bash
# Start both backend and frontend (if using just)
just dev

# Or manually
# Terminal 1: Backend
cd backend && go run cmd/server/main.go

# Terminal 2: Frontend
cd frontend && pnpm run dev
```

- **Backend**: http://localhost:8080
- **Frontend**: http://localhost:5173 (Vite dev server)

### Available Commands

```bash
just setup          # Install dependencies
just dev            # Start development environment
just build          # Build both frontend and backend
just test           # Run tests
just docker-build   # Build Docker image
just docker-run     # Run with Docker Compose
```

## 📁 How It Works

### File Organization

Sortify organizes files using this structure:
```
media/
├── 2024/
│   ├── January/
│   │   ├── IMG_20240115_143022.jpg
│   │   └── vacation_photo.png
│   └── December/
│       └── holiday_pics.jpg
└── 2025/
    └── March/
        └── spring_cleanup.jpg
```

### Date Extraction Priority

1. **EXIF DateTimeOriginal** - Most accurate for photos
2. **EXIF DateTime** - Camera timestamp
3. **Filename patterns** - Recognizes common formats:
   - `IMG_20240115_143022.jpg`
   - `2024-01-15_photo.png`
   - `photo_2024_01_15.jpg`
4. **File modification time** - Fallback option

### Upload Process

1. **File Selection** - Drag & drop or click to select
2. **Thumbnail Generation** - Real-time image previews
3. **Chunked Upload** - Large files uploaded in chunks
4. **Metadata Extraction** - Date and EXIF analysis
5. **Duplicate Check** - SHA256 checksum verification
6. **Organization** - File moved to date-based folder

## 🛡️ Security

- **Scratch-based container** - Minimal attack surface
- **Non-root execution** - Runs as user 1001
- **No shell access** - Container has no shell or utilities
- **Dependency scanning** - Automated security scans in CI/CD
- **CORS protection** - Configurable origin restrictions

## 🐳 Production Deployment

### Docker Compose Production

```yaml
services:
  sortify:
    image: steven-harris/sortify:latest
    restart: unless-stopped
    ports:
      - "8080:8080"
    environment:
      - MEDIA_PATH=/media
      - TEMP_PATH=/tmp
      - LOG_LEVEL=warn
      - CORS_ORIGINS=https://yourdomain.com
    volumes:
      - /path/to/media:/media:rw
      - /path/to/temp:/tmp:rw
    healthcheck:
      test: ["/healthcheck"]
      interval: 30s
      timeout: 10s
      retries: 3
```

### Reverse Proxy (Nginx)

```nginx
server {
    listen 80;
    server_name sortify.yourdomain.com;
    
    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        # For large file uploads
        client_max_body_size 500M;
        proxy_read_timeout 600s;
        proxy_send_timeout 600s;
    }
}
```

## 📊 Monitoring

### Health Check

```bash
curl http://localhost:8080/api/health
```

### Logs

```bash
# Docker logs
docker logs sortify --tail 100 -f

# Or with Docker Compose
docker compose logs -f
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Commit your changes: `git commit -m 'Add amazing feature'`
4. Push to the branch: `git push origin feature/amazing-feature`
5. Open a Pull Request

### Code Style

- **Go**: Follow standard Go conventions, use `gofmt`
- **TypeScript**: Prettier formatting, ESLint rules
- **Commits**: Conventional commits format

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Lit](https://lit.dev/) - Web components framework
- [Tailwind CSS](https://tailwindcss.com/) - Utility-first CSS framework
- [Fiber](https://gofiber.io/) - Express-inspired web framework for Go
- [Vite](https://vitejs.dev/) - Next generation frontend tooling

---

**⭐ Star this repo if you find it helpful!**
