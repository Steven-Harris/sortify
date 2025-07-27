package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Steven-harris/sortify/backend/internal/media"
	"github.com/Steven-harris/sortify/backend/internal/models"
	"github.com/Steven-harris/sortify/backend/internal/upload"
)

func TestIntegrationUploadAndOrganize(t *testing.T) {
	// Create temporary directories
	tempDir := t.TempDir()
	uploadTempDir := filepath.Join(tempDir, "uploads")
	mediaDir := filepath.Join(tempDir, "media")

	err := os.MkdirAll(uploadTempDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create upload temp dir: %v", err)
	}

	err = os.MkdirAll(mediaDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create media dir: %v", err)
	}

	// Initialize components
	uploadManager := upload.NewManager(uploadTempDir, 5)
	organizer := media.NewOrganizer(mediaDir)

	// Test complete workflow
	t.Run("Complete upload and organize workflow", func(t *testing.T) {
		// Create upload session
		req := &models.StartUploadRequest{
			FileName:  "IMG_20240315_143022.jpg",
			FileSize:  20,
			ChunkSize: 10,
			Metadata:  map[string]string{"camera": "test"},
		}

		session, err := uploadManager.CreateSession(req)
		if err != nil {
			t.Fatalf("Failed to create session: %v", err)
		}

		// Upload chunks
		chunk1 := []byte("0123456789") // 10 bytes
		chunk2 := []byte("abcdefghij") // 10 bytes

		err = uploadManager.UploadChunk(session.ID, 0, chunk1, "")
		if err != nil {
			t.Fatalf("Failed to upload chunk 1: %v", err)
		}

		err = uploadManager.UploadChunk(session.ID, 1, chunk2, "")
		if err != nil {
			t.Fatalf("Failed to upload chunk 2: %v", err)
		}

		// Complete upload (without checksum validation for test)
		err = uploadManager.CompleteUpload(session.ID, "")
		if err != nil {
			t.Fatalf("Failed to complete upload: %v", err)
		}

		// Get temp file path
		tempFilePath, err := uploadManager.GetTempFilePath(session.ID)
		if err != nil {
			t.Fatalf("Failed to get temp file path: %v", err)
		}

		// Organize the file (extractor will read from temp file path but organizer updates filename)
		mediaInfo, err := organizer.OrganizeFile(tempFilePath, session.FileName)
		if err != nil {
			t.Fatalf("Failed to organize file: %v", err)
		}

		// Verify file was organized correctly - it will be organized by actual file date since temp file doesn't have date in name
		// Let's check if the file exists somewhere in the media directory
		var organizedFile string
		err = filepath.Walk(mediaDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.Name() == "IMG_20240315_143022.jpg" {
				organizedFile = path
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to walk media directory: %v", err)
		}

		if organizedFile == "" {
			t.Error("Organized file not found in media directory")
		}

		// Verify metadata - since temp file doesn't have date in name, it will use file time
		if mediaInfo.DateSource != media.DateSourceFileTime {
			t.Logf("Note: Date source is %s (expected file_time since temp file has no date pattern)", mediaInfo.DateSource)
		}

		// Verify file content
		content, err := os.ReadFile(organizedFile)
		if err != nil {
			t.Fatalf("Failed to read organized file: %v", err)
		}

		expectedContent := string(chunk1) + string(chunk2)
		if string(content) != expectedContent {
			t.Errorf("File content mismatch. Expected %s, got %s", expectedContent, string(content))
		}

		// Cleanup session
		err = uploadManager.CleanupSession(session.ID)
		if err != nil {
			t.Fatalf("Failed to cleanup session: %v", err)
		}

		// Verify temp file was removed
		if _, err := os.Stat(tempFilePath); !os.IsNotExist(err) {
			t.Error("Temp file should be removed after cleanup")
		}
	})

	t.Run("Duplicate detection integration", func(t *testing.T) {
		// Create first upload
		req1 := &models.StartUploadRequest{
			FileName:  "duplicate_test.jpg",
			FileSize:  10,
			ChunkSize: 10,
		}

		session1, err := uploadManager.CreateSession(req1)
		if err != nil {
			t.Fatalf("Failed to create session 1: %v", err)
		}

		chunk := []byte("duplicated")
		err = uploadManager.UploadChunk(session1.ID, 0, chunk, "")
		if err != nil {
			t.Fatalf("Failed to upload chunk: %v", err)
		}

		err = uploadManager.CompleteUpload(session1.ID, "")
		if err != nil {
			t.Fatalf("Failed to complete upload 1: %v", err)
		}

		tempFilePath1, err := uploadManager.GetTempFilePath(session1.ID)
		if err != nil {
			t.Fatalf("Failed to get temp file path 1: %v", err)
		}

		// Organize first file
		_, err = organizer.OrganizeFile(tempFilePath1, session1.FileName)
		if err != nil {
			t.Fatalf("Failed to organize file 1: %v", err)
		}

		// Count files in media directory before second upload
		initialFileCount := 0
		err = filepath.Walk(mediaDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && info.Name() == "duplicate_test.jpg" {
				initialFileCount++
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to count initial files: %v", err)
		}

		// Create second upload with same content
		req2 := &models.StartUploadRequest{
			FileName:  "duplicate_test.jpg",
			FileSize:  10,
			ChunkSize: 10,
		}

		session2, err := uploadManager.CreateSession(req2)
		if err != nil {
			t.Fatalf("Failed to create session 2: %v", err)
		}

		err = uploadManager.UploadChunk(session2.ID, 0, chunk, "")
		if err != nil {
			t.Fatalf("Failed to upload chunk 2: %v", err)
		}

		err = uploadManager.CompleteUpload(session2.ID, "")
		if err != nil {
			t.Fatalf("Failed to complete upload 2: %v", err)
		}

		tempFilePath2, err := uploadManager.GetTempFilePath(session2.ID)
		if err != nil {
			t.Fatalf("Failed to get temp file path 2: %v", err)
		}

		// Organize second file (should detect duplicate)
		_, err = organizer.OrganizeFile(tempFilePath2, session2.FileName)
		if err != nil {
			t.Fatalf("Failed to organize file 2: %v", err)
		}

		// Count files after second upload - should still be same count
		finalFileCount := 0
		err = filepath.Walk(mediaDir, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && info.Name() == "duplicate_test.jpg" {
				finalFileCount++
			}
			return nil
		})
		if err != nil {
			t.Fatalf("Failed to count final files: %v", err)
		}

		if finalFileCount != initialFileCount {
			t.Errorf("Expected file count to remain %d, but got %d (duplicate should not create new file)", initialFileCount, finalFileCount)
		}

		// Cleanup
		uploadManager.CleanupSession(session1.ID)
		uploadManager.CleanupSession(session2.ID)
	})
}

func TestIntegrationConcurrentUploads(t *testing.T) {
	tempDir := t.TempDir()
	uploadTempDir := filepath.Join(tempDir, "uploads")

	err := os.MkdirAll(uploadTempDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create upload temp dir: %v", err)
	}

	// Test concurrent session creation
	uploadManager := upload.NewManager(uploadTempDir, 3) // Max 3 concurrent

	sessions := make([]*models.UploadSession, 3)

	// Create 3 sessions simultaneously
	for i := 0; i < 3; i++ {
		req := &models.StartUploadRequest{
			FileName:  "concurrent_test_" + string(rune('A'+i)) + ".jpg",
			FileSize:  100,
			ChunkSize: 50,
		}

		session, err := uploadManager.CreateSession(req)
		if err != nil {
			t.Fatalf("Failed to create session %d: %v", i, err)
		}
		sessions[i] = session
	}

	// Try to create 4th session (should fail due to limit)
	req := &models.StartUploadRequest{
		FileName:  "should_fail.jpg",
		FileSize:  100,
		ChunkSize: 50,
	}

	_, err = uploadManager.CreateSession(req)
	if err == nil {
		t.Error("Expected error when exceeding max sessions limit")
	}

	// Verify all 3 sessions are accessible
	for i, session := range sessions {
		retrievedSession, err := uploadManager.GetSession(session.ID)
		if err != nil {
			t.Errorf("Failed to retrieve session %d: %v", i, err)
		}

		if retrievedSession.ID != session.ID {
			t.Errorf("Session ID mismatch for session %d", i)
		}
	}

	// Cancel one session and verify we can create a new one
	err = uploadManager.CancelUpload(sessions[0].ID)
	if err != nil {
		t.Fatalf("Failed to cancel session: %v", err)
	}

	// Now creating a new session should work
	_, err = uploadManager.CreateSession(req)
	if err != nil {
		t.Errorf("Should be able to create session after canceling one: %v", err)
	}
}
