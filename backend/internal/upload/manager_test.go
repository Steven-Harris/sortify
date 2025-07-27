package upload

import (
	"crypto/sha256"
	"fmt"
	"os"
	"testing"

	"github.com/Steven-harris/sortify/backend/internal/models"
)

func TestNewManager(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, 5)

	if manager.tempDir != tempDir {
		t.Errorf("Expected tempDir %s, got %s", tempDir, manager.tempDir)
	}

	if manager.maxSessions != 5 {
		t.Errorf("Expected maxSessions 5, got %d", manager.maxSessions)
	}

	if len(manager.sessions) != 0 {
		t.Errorf("Expected empty sessions map, got %d sessions", len(manager.sessions))
	}
}

func TestCreateSession(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, 5)

	req := &models.StartUploadRequest{
		FileName:  "test.jpg",
		FileSize:  1024,
		ChunkSize: 256,
		Checksum:  "abcd1234",
		Metadata:  map[string]string{"source": "test"},
	}

	session, err := manager.CreateSession(req)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Verify session properties
	if session.FileName != req.FileName {
		t.Errorf("Expected filename %s, got %s", req.FileName, session.FileName)
	}

	if session.FileSize != req.FileSize {
		t.Errorf("Expected file size %d, got %d", req.FileSize, session.FileSize)
	}

	if session.ChunkSize != req.ChunkSize {
		t.Errorf("Expected chunk size %d, got %d", req.ChunkSize, session.ChunkSize)
	}

	expectedChunks := int((req.FileSize + req.ChunkSize - 1) / req.ChunkSize)
	if session.TotalChunks != expectedChunks {
		t.Errorf("Expected total chunks %d, got %d", expectedChunks, session.TotalChunks)
	}

	if session.UploadedSize != 0 {
		t.Errorf("Expected uploaded size 0, got %d", session.UploadedSize)
	}

	if session.Status != models.StatusInitialized {
		t.Errorf("Expected status %s, got %s", models.StatusInitialized, session.Status)
	}

	// Verify temporary file exists
	if _, err := os.Stat(session.TempPath); os.IsNotExist(err) {
		t.Errorf("Temporary file should exist at %s", session.TempPath)
	}

	// Verify file size is pre-allocated
	fileInfo, err := os.Stat(session.TempPath)
	if err != nil {
		t.Fatalf("Failed to stat temp file: %v", err)
	}

	if fileInfo.Size() != req.FileSize {
		t.Errorf("Expected temp file size %d, got %d", req.FileSize, fileInfo.Size())
	}
}

func TestCreateSessionMaxLimit(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, 2) // Max 2 sessions

	req := &models.StartUploadRequest{
		FileName:  "test.jpg",
		FileSize:  1024,
		ChunkSize: 256,
	}

	// Create first session - should succeed
	_, err := manager.CreateSession(req)
	if err != nil {
		t.Fatalf("First session creation failed: %v", err)
	}

	// Create second session - should succeed
	req.FileName = "test2.jpg"
	_, err = manager.CreateSession(req)
	if err != nil {
		t.Fatalf("Second session creation failed: %v", err)
	}

	// Create third session - should fail
	req.FileName = "test3.jpg"
	_, err = manager.CreateSession(req)
	if err == nil {
		t.Error("Expected error when exceeding max sessions, got nil")
	}
}

func TestGetSession(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, 5)

	req := &models.StartUploadRequest{
		FileName:  "test.jpg",
		FileSize:  1024,
		ChunkSize: 256,
	}

	// Create session
	createdSession, err := manager.CreateSession(req)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Get session
	retrievedSession, err := manager.GetSession(createdSession.ID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if retrievedSession.ID != createdSession.ID {
		t.Errorf("Expected session ID %s, got %s", createdSession.ID, retrievedSession.ID)
	}

	// Test non-existent session
	_, err = manager.GetSession("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent session, got nil")
	}
}

func TestUploadChunk(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, 5)

	req := &models.StartUploadRequest{
		FileName:  "test.jpg",
		FileSize:  1024,
		ChunkSize: 256,
	}

	session, err := manager.CreateSession(req)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Create test chunk data
	chunkData := []byte("test chunk data for chunk 0")
	expectedChecksum := fmt.Sprintf("%x", sha256.Sum256(chunkData))

	// Upload chunk
	err = manager.UploadChunk(session.ID, 0, chunkData, expectedChecksum)
	if err != nil {
		t.Fatalf("UploadChunk failed: %v", err)
	}

	// Verify session was updated
	updatedSession, err := manager.GetSession(session.ID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if updatedSession.UploadedSize != int64(len(chunkData)) {
		t.Errorf("Expected uploaded size %d, got %d", len(chunkData), updatedSession.UploadedSize)
	}

	if updatedSession.Status != models.StatusUploading {
		t.Errorf("Expected status %s, got %s", models.StatusUploading, updatedSession.Status)
	}

	// Verify chunk data was written to file
	file, err := os.Open(session.TempPath)
	if err != nil {
		t.Fatalf("Failed to open temp file: %v", err)
	}
	defer file.Close()

	readData := make([]byte, len(chunkData))
	n, err := file.Read(readData)
	if err != nil {
		t.Fatalf("Failed to read temp file: %v", err)
	}

	if n != len(chunkData) {
		t.Errorf("Expected to read %d bytes, got %d", len(chunkData), n)
	}

	if string(readData) != string(chunkData) {
		t.Errorf("Chunk data mismatch. Expected %s, got %s", string(chunkData), string(readData))
	}
}

func TestUploadChunkChecksumValidation(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, 5)

	req := &models.StartUploadRequest{
		FileName:  "test.jpg",
		FileSize:  1024,
		ChunkSize: 256,
	}

	session, err := manager.CreateSession(req)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	chunkData := []byte("test chunk data")
	wrongChecksum := "wrong_checksum"

	// Upload chunk with wrong checksum - should fail
	err = manager.UploadChunk(session.ID, 0, chunkData, wrongChecksum)
	if err == nil {
		t.Error("Expected error for wrong checksum, got nil")
	}

	if err.Error() != "chunk checksum mismatch" {
		t.Errorf("Expected 'chunk checksum mismatch' error, got %v", err)
	}
}

func TestCompleteUpload(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, 5)

	req := &models.StartUploadRequest{
		FileName:  "test.jpg",
		FileSize:  20, // Small file for testing
		ChunkSize: 10,
	}

	session, err := manager.CreateSession(req)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Upload all chunks
	chunk1 := []byte("0123456789") // 10 bytes
	chunk2 := []byte("abcdefghij") // 10 bytes

	err = manager.UploadChunk(session.ID, 0, chunk1, "")
	if err != nil {
		t.Fatalf("UploadChunk 0 failed: %v", err)
	}

	err = manager.UploadChunk(session.ID, 1, chunk2, "")
	if err != nil {
		t.Fatalf("UploadChunk 1 failed: %v", err)
	}

	// Complete upload
	err = manager.CompleteUpload(session.ID, "")
	if err != nil {
		t.Fatalf("CompleteUpload failed: %v", err)
	}

	// Verify session status
	completedSession, err := manager.GetSession(session.ID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if completedSession.Status != models.StatusCompleted {
		t.Errorf("Expected status %s, got %s", models.StatusCompleted, completedSession.Status)
	}
}

func TestCompleteUploadSizeMismatch(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, 5)

	req := &models.StartUploadRequest{
		FileName:  "test.jpg",
		FileSize:  20,
		ChunkSize: 10,
	}

	session, err := manager.CreateSession(req)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Upload only one chunk (incomplete)
	chunk1 := []byte("0123456789") // 10 bytes
	err = manager.UploadChunk(session.ID, 0, chunk1, "")
	if err != nil {
		t.Fatalf("UploadChunk failed: %v", err)
	}

	// Try to complete upload - should fail due to size mismatch
	err = manager.CompleteUpload(session.ID, "")
	if err == nil {
		t.Error("Expected error for incomplete upload, got nil")
	}

	if err.Error() != "uploaded size mismatch: expected 20, got 10" {
		t.Errorf("Expected size mismatch error, got %v", err)
	}
}

func TestGetProgress(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, 5)

	req := &models.StartUploadRequest{
		FileName:  "test.jpg",
		FileSize:  100,
		ChunkSize: 25,
	}

	session, err := manager.CreateSession(req)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Get initial progress
	progress, err := manager.GetProgress(session.ID)
	if err != nil {
		t.Fatalf("GetProgress failed: %v", err)
	}

	if progress.PercentComplete != 0 {
		t.Errorf("Expected 0%% progress, got %.2f%%", progress.PercentComplete)
	}

	if progress.TotalChunks != 4 {
		t.Errorf("Expected 4 total chunks, got %d", progress.TotalChunks)
	}

	// Upload one chunk
	chunk := make([]byte, 25)
	err = manager.UploadChunk(session.ID, 0, chunk, "")
	if err != nil {
		t.Fatalf("UploadChunk failed: %v", err)
	}

	// Get updated progress
	progress, err = manager.GetProgress(session.ID)
	if err != nil {
		t.Fatalf("GetProgress failed: %v", err)
	}

	expectedPercent := float64(25) / float64(100) * 100
	if progress.PercentComplete != expectedPercent {
		t.Errorf("Expected %.2f%% progress, got %.2f%%", expectedPercent, progress.PercentComplete)
	}

	if progress.UploadedChunks != 1 {
		t.Errorf("Expected 1 uploaded chunk, got %d", progress.UploadedChunks)
	}
}

func TestPauseAndResumeUpload(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, 5)

	req := &models.StartUploadRequest{
		FileName:  "test.jpg",
		FileSize:  1024,
		ChunkSize: 256,
	}

	session, err := manager.CreateSession(req)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Pause upload
	err = manager.PauseUpload(session.ID)
	if err != nil {
		t.Fatalf("PauseUpload failed: %v", err)
	}

	// Verify status
	pausedSession, err := manager.GetSession(session.ID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if pausedSession.Status != models.StatusPaused {
		t.Errorf("Expected status %s, got %s", models.StatusPaused, pausedSession.Status)
	}

	// Resume upload
	err = manager.ResumeUpload(session.ID)
	if err != nil {
		t.Fatalf("ResumeUpload failed: %v", err)
	}

	// Verify status
	resumedSession, err := manager.GetSession(session.ID)
	if err != nil {
		t.Fatalf("GetSession failed: %v", err)
	}

	if resumedSession.Status != models.StatusUploading {
		t.Errorf("Expected status %s, got %s", models.StatusUploading, resumedSession.Status)
	}
}

func TestCancelUpload(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, 5)

	req := &models.StartUploadRequest{
		FileName:  "test.jpg",
		FileSize:  1024,
		ChunkSize: 256,
	}

	session, err := manager.CreateSession(req)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	tempPath := session.TempPath

	// Verify temp file exists
	if _, err := os.Stat(tempPath); os.IsNotExist(err) {
		t.Fatalf("Temp file should exist before cancel")
	}

	// Cancel upload
	err = manager.CancelUpload(session.ID)
	if err != nil {
		t.Fatalf("CancelUpload failed: %v", err)
	}

	// Verify temp file was removed
	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Error("Temp file should be removed after cancel")
	}

	// Verify session was removed
	_, err = manager.GetSession(session.ID)
	if err == nil {
		t.Error("Session should be removed after cancel")
	}
}

func TestCleanupSession(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, 5)

	req := &models.StartUploadRequest{
		FileName:  "test.jpg",
		FileSize:  10,
		ChunkSize: 10,
	}

	session, err := manager.CreateSession(req)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	tempPath := session.TempPath

	// Upload chunk and complete the upload first
	chunk := []byte("0123456789") // 10 bytes to match FileSize
	err = manager.UploadChunk(session.ID, 0, chunk, "")
	if err != nil {
		t.Fatalf("UploadChunk failed: %v", err)
	}

	err = manager.CompleteUpload(session.ID, "")
	if err != nil {
		t.Fatalf("CompleteUpload failed: %v", err)
	}

	// Cleanup session
	err = manager.CleanupSession(session.ID)
	if err != nil {
		t.Fatalf("CleanupSession failed: %v", err)
	}

	// Verify temp file was removed
	if _, err := os.Stat(tempPath); !os.IsNotExist(err) {
		t.Error("Temp file should be removed after cleanup")
	}

	// Verify session was removed
	_, err = manager.GetSession(session.ID)
	if err == nil {
		t.Error("Session should be removed after cleanup")
	}
}

func TestGetTempFilePath(t *testing.T) {
	tempDir := t.TempDir()
	manager := NewManager(tempDir, 5)

	req := &models.StartUploadRequest{
		FileName:  "test.jpg",
		FileSize:  10,
		ChunkSize: 10,
	}

	session, err := manager.CreateSession(req)
	if err != nil {
		t.Fatalf("CreateSession failed: %v", err)
	}

	// Try to get temp path before completion - should fail
	_, err = manager.GetTempFilePath(session.ID)
	if err == nil {
		t.Error("Expected error for incomplete session, got nil")
	}

	// Upload chunk and complete
	chunk := []byte("0123456789")
	err = manager.UploadChunk(session.ID, 0, chunk, "")
	if err != nil {
		t.Fatalf("UploadChunk failed: %v", err)
	}

	err = manager.CompleteUpload(session.ID, "")
	if err != nil {
		t.Fatalf("CompleteUpload failed: %v", err)
	}

	// Now should get temp path successfully
	tempPath, err := manager.GetTempFilePath(session.ID)
	if err != nil {
		t.Fatalf("GetTempFilePath failed: %v", err)
	}

	if tempPath != session.TempPath {
		t.Errorf("Expected temp path %s, got %s", session.TempPath, tempPath)
	}
}
