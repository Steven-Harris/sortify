# 🎉 Sortify v0.0.1 - Initial Release

**First stable release of Sortify - an intelligent media file organizer with drag-and-drop web interface.**

## ✨ **Features**

### 🖥️ **Web Interface**
- Modern, responsive web UI built with Lit web components
- Drag-and-drop file upload with visual feedback
- Real-time image thumbnail generation
- Upload progress tracking with retry functionality
- Clean, intuitive design with Tailwind CSS

### 📁 **Media Organization**
- Automatic file organization by date (YYYY/Month structure)
- Smart date extraction from filenames and EXIF data
- Duplicate file detection and handling
- Support for images and various media formats
- Configurable media storage location

### 🔄 **Upload System**
- Chunked file uploads for large files
- Resume capability for interrupted uploads
- Configurable temporary directory
- Session-based upload management
- Comprehensive error handling and retry logic

### 🛡️ **Production Ready**
- Minimal Docker image based on scratch (security-focused)
- Multi-architecture support (AMD64, ARM64)
- Built-in health checks
- Comprehensive logging with structured JSON output
- Non-root container execution (user 1001)

### ⚙️ **Configuration**
- Environment variable configuration
- Configurable CORS settings
- Adjustable log levels
- Flexible port and path settings

## 🐳 **Docker Deployment**

**Quick Start:**
```bash
docker run -d \
  --name sortify \
  -p 8080:8080 \
  -e MEDIA_PATH=/media \
  -e TEMP_PATH=/tmp \
  -v /path/to/media:/media:rw \
  -v /path/to/temp:/tmp:rw \
  steven-harris/sortify:v0.0.1
```

**Docker Compose:**
Download the included `docker-compose.yml` and run:
```bash
docker compose up -d
```

## 🔧 **Technical Stack**
- **Backend**: Go with fiber web framework
- **Frontend**: TypeScript, Lit web components, Vite build system
- **Styling**: Tailwind CSS v4
- **Container**: Multi-stage Docker build with scratch base image
- **Architecture**: Multi-platform (linux/amd64, linux/arm64)

## 📋 **Requirements**
- Docker or compatible container runtime
- Persistent storage for media files
- Network access on port 8080 (configurable)

## 🚀 **Getting Started**
1. Download the release assets
2. Configure `docker-compose.yml` for your storage paths
3. Run `docker compose up -d`
4. Access the web interface at `http://127.0.0.1:8080`
5. Start uploading and organizing your media files!

## 🔐 **Security**
- Scratch-based container image for minimal attack surface
- Non-root execution with dedicated user (1001)
- Comprehensive security scanning in CI/CD pipeline
- No unnecessary dependencies or tools in production image

## 📦 **Release Assets**
- `docker-compose.yml` - Ready-to-use Docker Compose configuration
- `Dockerfile` - For custom builds and modifications
- `justfile` - Development and build commands
- `README.md` - Complete documentation

---

**Full Changelog**: https://github.com/Steven-harris/sortify/commits/v0.0.1

**Container Images**: 
- `steven-harris/sortify:v0.0.1`
- `steven-harris/sortify:0.0`  
- `steven-harris/sortify:latest`

Enjoy organizing your media files! 📸✨
