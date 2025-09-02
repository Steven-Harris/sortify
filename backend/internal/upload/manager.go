package upload

import (
	"crypto/sha256"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/Steven-Harris/sortify/backend/internal/models"
	"github.com/Steven-Harris/sortify/backend/internal/security"
	"github.com/Steven-Harris/sortify/backend/internal/storage"
)

type Manager struct {
	sessions       map[string]*models.UploadSession
	tempDir        string
	maxSessions    int
	mutex          sync.RWMutex
	storageManager *storage.Manager
}

func NewManager(tempDir string, maxSessions int) *Manager {
	if err := os.MkdirAll(tempDir, 0600); err != nil {
		// Log error but don't fail - we'll handle directory creation issues when needed
		slog.Warn("Failed to create temp directory during manager initialization", "error", err, "tempDir", tempDir)
	}

	// You may want to pass mediaPath here, adjust as needed
	storageMgr := storage.NewManager("")

	return &Manager{
		sessions:       make(map[string]*models.UploadSession),
		tempDir:        tempDir,
		maxSessions:    maxSessions,
		storageManager: storageMgr,
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

	// Validate temp path using security helper and ensure it's within temp directory
	cleanPath, err := security.ValidatePathWithinDirectory(tempPath, m.tempDir)
	if err != nil {
		return nil, fmt.Errorf("temp path validation failed: %w", err)
	}

	session := &models.UploadSession{
		ID:           sessionID,
		FileName:     req.FileName,
		FileSize:     req.FileSize,
		ChunkSize:    req.ChunkSize,
		TotalChunks:  totalChunks,
		UploadedSize: 0,
		Checksum:     req.Checksum,
		TempPath:     cleanPath,
		Metadata:     req.Metadata,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
		Status:       models.StatusInitialized,
	}

	file, err := os.Create(cleanPath) // #nosec G304 - path validated by security.ValidatePathWithinDirectory
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary file: %w", err)
	}

	if err := file.Truncate(req.FileSize); err != nil {
		if closeErr := file.Close(); closeErr != nil {
			slog.Warn("Failed to close file after truncate error", "error", closeErr)
		}
		if removeErr := os.Remove(tempPath); removeErr != nil {
			slog.Warn("Failed to remove temp file after truncate error", "error", removeErr, "path", tempPath)
		}
		return nil, fmt.Errorf("failed to allocate file space: %w", err)
	}
	if err := file.Close(); err != nil {
		slog.Warn("Failed to close file after truncate", "error", err)
	}

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

func (m *Manager) UploadChunk(sessionID string, chunkNumber int, chunkData []byte, expectedChecksum string, algorithm string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	var actualChecksum string
	if algorithm == "simple" {
		// Simple fallback hash
		var hash uint32 = 0
		for i := 0; i < len(chunkData); i++ {
			// Use the updated simple hash algorithm that's consistent with frontend
			hash = (hash + uint32(chunkData[i])) & 0xFFFFFFFF
		}
		actualChecksum = fmt.Sprintf("%x", hash)
	} else {
		hash := sha256.Sum256(chunkData)
		actualChecksum = fmt.Sprintf("%x", hash)
	}

	if expectedChecksum != "" && actualChecksum != expectedChecksum {
		return fmt.Errorf("chunk checksum mismatch")
	}

	return m.writeChunk(session, chunkNumber, chunkData)
}

// UploadChunkNoVerify uploads a chunk without verifying the checksum
// This is useful for recovering from checksum mismatch errors
func (m *Manager) UploadChunkNoVerify(sessionID string, chunkNumber int, chunkData []byte, serverChecksum string, algorithm string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	// Log that we're using the no-verify path for debugging
	slog.Info("Using UploadChunkNoVerify",
		"sessionId", sessionID,
		"chunk_number", chunkNumber,
		"algorithm", algorithm,
		"length", len(chunkData),
	)

	return m.writeChunk(session, chunkNumber, chunkData)
}

// Helper method to write chunk data to disk
func (m *Manager) writeChunk(session *models.UploadSession, chunkNumber int, chunkData []byte) error {
	offset := int64(chunkNumber) * session.ChunkSize

	file, err := os.OpenFile(session.TempPath, os.O_WRONLY, 0600)
	if err != nil {
		return fmt.Errorf("failed to open temporary file: %w", err)
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Warn("Failed to close temp file after chunk write", "error", err)
		}
	}()

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

func (m *Manager) CompleteUpload(sessionID string, expectedChecksum string, algorithm string) error {
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
		actualChecksum, err := m.calculateFileChecksum(session.TempPath, algorithm)
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

// CompleteUploadForce completes an upload without verifying the checksum.
// This is useful for recovering from checksum verification errors, especially for large files
// or files uploaded from mobile devices where checksums may not match exactly.
func (m *Manager) CompleteUploadForce(sessionID string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	session, exists := m.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found")
	}

	// Only verify that we received all bytes, but don't check the checksum
	if session.UploadedSize != session.FileSize {
		return fmt.Errorf("uploaded size mismatch: expected %d, got %d", session.FileSize, session.UploadedSize)
	}

	slog.Warn("Force completing upload without checksum verification",
		"sessionId", sessionID,
		"fileName", session.FileName,
		"fileSize", session.FileSize,
	)

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

	if err := os.Remove(session.TempPath); err != nil {
		slog.Warn("Failed to remove temp file during cancel", "error", err, "path", session.TempPath)
	}

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

	if err := os.Remove(session.TempPath); err != nil {
		slog.Warn("Failed to remove temp file during cleanup", "error", err, "path", session.TempPath)
	}

	delete(m.sessions, sessionID)

	return nil
}

func (m *Manager) calculateFileChecksum(filePath string, algorithm string) (string, error) {
	return m.storageManager.CalculateChecksum(filePath, algorithm)
}

func generateSessionID() string {
	return fmt.Sprintf("upload_%d_%d", time.Now().UnixNano(), time.Now().Unix())
}
