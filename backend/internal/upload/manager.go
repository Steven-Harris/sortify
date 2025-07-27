package upload

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Steven-harris/sortify/backend/internal/models"
)

// Manager handles upload sessions and operations
type Manager struct {
	sessions   map[string]*models.UploadSession
	tempDir    string
	maxSessions int
	mutex      sync.RWMutex
}

// NewManager creates a new upload manager
func NewManager(tempDir string, maxSessions int) *Manager {
	// Ensure temp directory exists
	os.MkdirAll(tempDir, 0755)
	
	return &Manager{
		sessions:    make(map[string]*models.UploadSession),
		tempDir:     tempDir,
		maxSessions: maxSessions,
	}
}

// CreateSession creates a new upload session
func (m *Manager) CreateSession(req *models.StartUploadRequest) (*models.UploadSession, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	// Check if we've reached max sessions
	if len(m.sessions) >= m.maxSessions {
		return nil, fmt.Errorf("maximum concurrent uploads reached")
	}

	// Generate session ID
	sessionID := generateSessionID()

	// Calculate total chunks
	totalChunks := int((req.FileSize + req.ChunkSize - 1) / req.ChunkSize)

	// Create temporary file path
	tempPath := filepath.Join(m.tempDir, sessionID+".tmp")

	session := &models.UploadSession{
		ID:           sessionID,
		FileName:     req.FileName,
		FileSize:     req.FileSize,
		ChunkSize:    req.ChunkSize,
		TotalChunks:  totalChunks,
		UploadedSize: 0,
		Checksum:     req.Checksum,
		TempPath:     tempPath,
		Metadata:     req.Metadata,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Status:       models.StatusInitialized,
	}

	// Create temporary file
	file, err := os.Create(tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}
	
	// Pre-allocate file space for better performance
	if err := file.Truncate(req.FileSize); err != nil {
		file.Close()
		os.Remove(tempPath)
		return nil, fmt.Errorf("failed to allocate file space: %w", err)
	}
	file.Close()

	m.sessions[sessionID] = session
	return session, nil
}

// GetSession retrieves an upload session
func (m *Manager) GetSession(sessionID string) (*models.UploadSession, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	return session, nil
}

// UploadChunk handles uploading a chunk of data
func (m *Manager) UploadChunk(sessionID string, chunkNumber int, chunkData []byte, expectedChecksum string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	// Verify chunk checksum
	hash := sha256.Sum256(chunkData)
	actualChecksum := fmt.Sprintf("%x", hash)
	if expectedChecksum != "" && actualChecksum != expectedChecksum {
		return fmt.Errorf("chunk checksum mismatch")
	}

	// Calculate chunk offset
	offset := int64(chunkNumber) * session.ChunkSize

	// Open temporary file for writing
	file, err := os.OpenFile(session.TempPath, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open temporary file: %w", err)
	}
	defer file.Close()

	// Seek to the correct position
	if _, err := file.Seek(offset, 0); err != nil {
		return fmt.Errorf("failed to seek to chunk position: %w", err)
	}

	// Write chunk data
	if _, err := file.Write(chunkData); err != nil {
		return fmt.Errorf("failed to write chunk data: %w", err)
	}

	// Update session
	session.UploadedSize += int64(len(chunkData))
	session.UpdatedAt = time.Now()
	session.Status = models.StatusUploading

	return nil
}

// CompleteUpload finalizes an upload session
func (m *Manager) CompleteUpload(sessionID string, expectedChecksum string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	// Verify file size
	if session.UploadedSize != session.FileSize {
		return fmt.Errorf("uploaded size mismatch: expected %d, got %d", 
			session.FileSize, session.UploadedSize)
	}

	// Verify file checksum if provided
	if expectedChecksum != "" || session.Checksum != "" {
		actualChecksum, err := m.calculateFileChecksum(session.TempPath)
		if err != nil {
			return fmt.Errorf("failed to calculate file checksum: %w", err)
		}

		checksumToVerify := expectedChecksum
		if checksumToVerify == "" {
			checksumToVerify = session.Checksum
		}

		if checksumToVerify != "" && actualChecksum != checksumToVerify {
			return fmt.Errorf("file checksum mismatch")
		}
	}

	// Update session status
	session.Status = models.StatusCompleted
	session.UpdatedAt = time.Now()

	return nil
}

// GetProgress returns the current progress of an upload
func (m *Manager) GetProgress(sessionID string) (*models.UploadProgress, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	percentComplete := float64(0)
	if session.FileSize > 0 {
		percentComplete = float64(session.UploadedSize) / float64(session.FileSize) * 100
	}

	uploadedChunks := int(session.UploadedSize / session.ChunkSize)
	if session.UploadedSize%session.ChunkSize > 0 {
		uploadedChunks++
	}

	return &models.UploadProgress{
		SessionID:       session.ID,
		FileName:        session.FileName,
		UploadedBytes:   session.UploadedSize,
		TotalBytes:      session.FileSize,
		UploadedChunks:  uploadedChunks,
		TotalChunks:     session.TotalChunks,
		PercentComplete: percentComplete,
		Status:          string(session.Status),
	}, nil
}

// PauseUpload pauses an upload session
func (m *Manager) PauseUpload(sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	session.Status = models.StatusPaused
	session.UpdatedAt = time.Now()

	return nil
}

// ResumeUpload resumes a paused upload session
func (m *Manager) ResumeUpload(sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	if session.Status != models.StatusPaused {
		return fmt.Errorf("session is not paused")
	}

	session.Status = models.StatusUploading
	session.UpdatedAt = time.Now()

	return nil
}

// CancelUpload cancels an upload session and cleans up temporary files
func (m *Manager) CancelUpload(sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	// Remove temporary file
	os.Remove(session.TempPath)

	// Update session status
	session.Status = models.StatusCancelled
	session.UpdatedAt = time.Now()

	// Remove session from memory
	delete(m.sessions, sessionID)

	return nil
}

// GetTempFilePath returns the temporary file path for a completed session
func (m *Manager) GetTempFilePath(sessionID string) (string, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return "", fmt.Errorf("session not found")
	}

	if session.Status != models.StatusCompleted {
		return "", fmt.Errorf("session not completed")
	}

	return session.TempPath, nil
}

// CleanupSession removes a session and its temporary file
func (m *Manager) CleanupSession(sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	// Remove temporary file
	os.Remove(session.TempPath)

	// Remove session from memory
	delete(m.sessions, sessionID)

	return nil
}

// calculateFileChecksum calculates SHA256 checksum of a file
func (m *Manager) calculateFileChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// generateSessionID generates a unique session ID
func generateSessionID() string {
	return fmt.Sprintf("upload_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}
