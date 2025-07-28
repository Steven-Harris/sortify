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

type Manager struct {
	sessions    map[string]*models.UploadSession
	tempDir     string
	maxSessions int
	mutex       sync.RWMutex
}

func NewManager(tempDir string, maxSessions int) *Manager {
	os.MkdirAll(tempDir, 0755)

	return &Manager{
		sessions:    make(map[string]*models.UploadSession),
		tempDir:     tempDir,
		maxSessions: maxSessions,
	}
}

func (m *Manager) CreateSession(req *models.StartUploadRequest) (*models.UploadSession, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if len(m.sessions) >= m.maxSessions {
		return nil, fmt.Errorf("maximum concurrent uploads reached")
	}

	sessionID := generateSessionID()

	totalChunks := int((req.FileSize + req.ChunkSize - 1) / req.ChunkSize)

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

	file, err := os.Create(tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	if err := file.Truncate(req.FileSize); err != nil {
		file.Close()
		os.Remove(tempPath)
		return nil, fmt.Errorf("failed to allocate file space: %w", err)
	}
	file.Close()

	m.sessions[sessionID] = session
	return session, nil
}

func (m *Manager) GetSession(sessionID string) (*models.UploadSession, error) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return nil, fmt.Errorf("session not found")
	}

	return session, nil
}

func (m *Manager) UploadChunk(sessionID string, chunkNumber int, chunkData []byte, expectedChecksum string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	hash := sha256.Sum256(chunkData)
	actualChecksum := fmt.Sprintf("%x", hash)
	if expectedChecksum != "" && actualChecksum != expectedChecksum {
		return fmt.Errorf("chunk checksum mismatch")
	}

	offset := int64(chunkNumber) * session.ChunkSize

	file, err := os.OpenFile(session.TempPath, os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open temporary file: %w", err)
	}
	defer file.Close()

	if _, err := file.Seek(offset, 0); err != nil {
		return fmt.Errorf("failed to seek to chunk position: %w", err)
	}

	if _, err := file.Write(chunkData); err != nil {
		return fmt.Errorf("failed to write chunk data: %w", err)
	}

	session.UploadedSize += int64(len(chunkData))
	session.UpdatedAt = time.Now()
	session.Status = models.StatusUploading

	return nil
}

func (m *Manager) CompleteUpload(sessionID string, expectedChecksum string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	if session.UploadedSize != session.FileSize {
		return fmt.Errorf("uploaded size mismatch: expected %d, got %d", session.FileSize, session.UploadedSize)
	}

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

	session.Status = models.StatusCompleted
	session.UpdatedAt = time.Now()

	return nil
}

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

func (m *Manager) CancelUpload(sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	os.Remove(session.TempPath)

	session.Status = models.StatusCancelled
	session.UpdatedAt = time.Now()

	delete(m.sessions, sessionID)

	return nil
}

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

func (m *Manager) CleanupSession(sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	os.Remove(session.TempPath)

	delete(m.sessions, sessionID)

	return nil
}

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

func generateSessionID() string {
	return fmt.Sprintf("upload_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}
