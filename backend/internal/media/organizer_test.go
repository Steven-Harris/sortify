package media

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewOrganizer(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	assert.NotNil(t, organizer, "NewOrganizer should not return nil")
	assert.Equal(t, tempDir, organizer.mediaPath, "Expected media path to match")
	assert.NotNil(t, organizer.extractor, "Expected extractor to be initialized")
}

func TestGetTargetDirectory(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	t.Run("valid date", func(t *testing.T) {
		date := time.Date(2024, 3, 15, 14, 30, 22, 0, time.UTC)
		expected := filepath.Join(tempDir, "2024", "March")

		result, err := organizer.getTargetDirectory(&date)
		assert.NoError(t, err, "getTargetDirectory should not fail")
		assert.Equal(t, expected, result, "Expected target directory to match")
	})

	t.Run("different year and month", func(t *testing.T) {
		date := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)
		expected := filepath.Join(tempDir, "2023", "December")

		result, err := organizer.getTargetDirectory(&date)
		assert.NoError(t, err, "getTargetDirectory should not fail")
		assert.Equal(t, expected, result, "Expected target directory to match")
	})

	t.Run("single digit month", func(t *testing.T) {
		date := time.Date(2022, 7, 8, 0, 0, 0, 0, time.UTC)
		expected := filepath.Join(tempDir, "2022", "July")

		result, err := organizer.getTargetDirectory(&date)
		assert.NoError(t, err, "getTargetDirectory should not fail")
		assert.Equal(t, expected, result, "Expected target directory to match")
	})

	t.Run("nil date uses current date", func(t *testing.T) {
		result, err := organizer.getTargetDirectory(nil)
		assert.NoError(t, err, "getTargetDirectory should not fail")

		// For nil date, just verify it uses a valid year/month format
		assert.True(t, filepath.IsAbs(result), "Expected absolute path")
		// Should contain current year
		currentYear := time.Now().Format("2006")
		assert.Contains(t, result, currentYear, "Expected path to contain current year")
	})
}

func TestOrganizeFile(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	t.Run("successful file organization", func(t *testing.T) {
		// Create a test file with date in filename
		sourceFile := filepath.Join(tempDir, "source", "IMG_20240315_143022.jpg")
		err := os.MkdirAll(filepath.Dir(sourceFile), 0755)
		assert.NoError(t, err, "Failed to create source directory")

		testContent := []byte("test image content")
		err = os.WriteFile(sourceFile, testContent, 0644)
		assert.NoError(t, err, "Failed to create test file")

		// Organize the file
		mediaInfo, err := organizer.OrganizeFile(sourceFile, "IMG_20240315_143022.jpg")
		assert.NoError(t, err, "OrganizeFile should not fail")

		// Verify metadata
		expectedDate := time.Date(2024, 3, 15, 14, 30, 22, 0, time.UTC)
		assert.NotNil(t, mediaInfo.DateTaken, "Expected date to be extracted")
		assert.True(t, mediaInfo.DateTaken.Equal(expectedDate), "Expected date %v, got %v", expectedDate, mediaInfo.DateTaken)
		assert.Equal(t, DateSourceFileName, mediaInfo.DateSource, "Expected date source to be filename")
		assert.Equal(t, MediaTypePhoto, mediaInfo.MediaType, "Expected media type to be photo")

		// Verify file was moved to correct location
		expectedDir := filepath.Join(tempDir, "2024", "March")
		expectedFile := filepath.Join(expectedDir, "IMG_20240315_143022.jpg")
		assert.FileExists(t, expectedFile, "File should exist in target location")

		// Verify file content
		movedContent, err := os.ReadFile(expectedFile)
		assert.NoError(t, err, "Failed to read moved file")
		assert.Equal(t, string(testContent), string(movedContent), "File content should match original")

		// Verify source file was removed
		assert.NoFileExists(t, sourceFile, "Source file should be removed after organizing")
	})

	t.Run("duplicate file handling", func(t *testing.T) {
		// Create target directory and existing file
		targetDir := filepath.Join(tempDir, "2024", "March")
		err := os.MkdirAll(targetDir, 0755)
		assert.NoError(t, err, "Failed to create target directory")

		existingFile := filepath.Join(targetDir, "IMG_20240315_143022_dup.jpg")
		existingContent := []byte("existing content")
		err = os.WriteFile(existingFile, existingContent, 0644)
		assert.NoError(t, err, "Failed to create existing file")

		// Create source file with same content (exact duplicate)
		sourceFile := filepath.Join(tempDir, "source", "IMG_20240315_143022_dup.jpg")
		err = os.MkdirAll(filepath.Dir(sourceFile), 0755)
		assert.NoError(t, err, "Failed to create source directory")

		err = os.WriteFile(sourceFile, existingContent, 0644)
		assert.NoError(t, err, "Failed to create source file")

		// Organize the duplicate file
		mediaInfo, err := organizer.OrganizeFile(sourceFile, "IMG_20240315_143022_dup.jpg")
		assert.NoError(t, err, "OrganizeFile should not fail for duplicates")

		// Should have detected duplicate and removed source without error
		assert.NoFileExists(t, sourceFile, "Source file should be removed after duplicate detection")
		assert.FileExists(t, existingFile, "Existing file should remain")
		assert.NotNil(t, mediaInfo, "MediaInfo should be returned even for duplicates")
	})

	t.Run("file name conflict resolution", func(t *testing.T) {
		// Create target directory and existing file with different content
		targetDir := filepath.Join(tempDir, "2024", "March")
		err := os.MkdirAll(targetDir, 0755)
		assert.NoError(t, err, "Failed to create target directory")

		existingFile := filepath.Join(targetDir, "IMG_20240315_143022_conflict.jpg")
		existingContent := []byte("existing different content")
		err = os.WriteFile(existingFile, existingContent, 0644)
		assert.NoError(t, err, "Failed to create existing file")

		// Create source file with different content
		sourceFile := filepath.Join(tempDir, "source", "IMG_20240315_143022_conflict.jpg")
		err = os.MkdirAll(filepath.Dir(sourceFile), 0755)
		assert.NoError(t, err, "Failed to create source directory")

		sourceContent := []byte("new different content")
		err = os.WriteFile(sourceFile, sourceContent, 0644)
		assert.NoError(t, err, "Failed to create source file")

		// Organize the conflicting file
		mediaInfo, err := organizer.OrganizeFile(sourceFile, "IMG_20240315_143022_conflict.jpg")
		assert.NoError(t, err, "OrganizeFile should not fail for conflicts")

		// Should have renamed the new file with (1) format
		renamedFile := filepath.Join(targetDir, "IMG_20240315_143022_conflict(1).jpg")
		assert.FileExists(t, renamedFile, "Renamed file should exist")

		// Verify content of renamed file
		renamedContent, err := os.ReadFile(renamedFile)
		assert.NoError(t, err, "Failed to read renamed file")
		assert.Equal(t, string(sourceContent), string(renamedContent), "Renamed file content should match source")

		// Original existing file should remain unchanged
		originalContent, err := os.ReadFile(existingFile)
		assert.NoError(t, err, "Failed to read original file")
		assert.Equal(t, string(existingContent), string(originalContent), "Original file content should remain unchanged")

		// MediaInfo should reflect the original filename
		assert.Equal(t, "IMG_20240315_143022_conflict.jpg", mediaInfo.FileName, "Expected original filename in metadata")
	})

	t.Run("non existent source file", func(t *testing.T) {
		_, err := organizer.OrganizeFile("/non/existent/file.jpg", "test.jpg")
		assert.Error(t, err, "Expected error for non-existent source file")
	})

	t.Run("file with unknown date pattern", func(t *testing.T) {
		// Create a file with no recognizable date pattern
		sourceFile := filepath.Join(tempDir, "source", "random_name.jpg")
		err := os.MkdirAll(filepath.Dir(sourceFile), 0755)
		assert.NoError(t, err, "Failed to create source directory")

		err = os.WriteFile(sourceFile, []byte("content"), 0644)
		assert.NoError(t, err, "Failed to create source file")

		mediaInfo, err := organizer.OrganizeFile(sourceFile, "random_name.jpg")
		assert.NoError(t, err, "OrganizeFile should not fail for unknown date")

		// Should fall back to file time
		assert.Equal(t, DateSourceFileTime, mediaInfo.DateSource, "Expected date source to be file time")
		assert.NotNil(t, mediaInfo.DateTaken, "Date taken should be set even when falling back to file time")

		// Should be placed in a year/month directory based on file time
		expectedPattern := filepath.Join(tempDir, "*", "*", "random_name.jpg")
		matches, err := filepath.Glob(expectedPattern)
		assert.NoError(t, err, "Failed to glob for organized file")
		assert.Len(t, matches, 1, "Expected exactly one organized file")
	})
}

func TestOrganizeFileDuplicate(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	// Create target directory and existing file
	targetDir := filepath.Join(tempDir, "2024", "March")
	err := os.MkdirAll(targetDir, 0755)
	assert.NoError(t, err, "Failed to create target directory")

	existingFile := filepath.Join(targetDir, "IMG_20240315_143022.jpg")
	existingContent := []byte("existing content")
	err = os.WriteFile(existingFile, existingContent, 0644)
	assert.NoError(t, err, "Failed to create existing file")

	// Create source file with same content (exact duplicate)
	sourceFile := filepath.Join(tempDir, "source", "IMG_20240315_143022.jpg")
	err = os.MkdirAll(filepath.Dir(sourceFile), 0755)
	assert.NoError(t, err, "Failed to create source directory")

	err = os.WriteFile(sourceFile, existingContent, 0644)
	assert.NoError(t, err, "Failed to create source file")

	// Organize the duplicate file
	mediaInfo, err := organizer.OrganizeFile(sourceFile, "IMG_20240315_143022.jpg")
	assert.NoError(t, err, "OrganizeFile should not fail")

	// Should have detected duplicate and removed source without error
	_, err = os.Stat(sourceFile)
	assert.True(t, os.IsNotExist(err), "Source file should be removed after duplicate detection")

	// Original file should still exist
	_, err = os.Stat(existingFile)
	assert.False(t, os.IsNotExist(err), "Existing file should remain")

	// Metadata should still be returned
	assert.NotNil(t, mediaInfo, "MediaInfo should be returned even for duplicates")
}

func TestOrganizeFileConflict(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	// Create target directory and existing file with different content
	targetDir := filepath.Join(tempDir, "2024", "March")
	err := os.MkdirAll(targetDir, 0755)
	assert.NoError(t, err, "Failed to create target directory")

	existingFile := filepath.Join(targetDir, "IMG_20240315_143022.jpg")
	existingContent := []byte("existing different content")
	err = os.WriteFile(existingFile, existingContent, 0644)
	assert.NoError(t, err, "Failed to create existing file")

	// Create source file with different content
	sourceFile := filepath.Join(tempDir, "source", "IMG_20240315_143022.jpg")
	err = os.MkdirAll(filepath.Dir(sourceFile), 0755)
	assert.NoError(t, err, "Failed to create source directory")

	sourceContent := []byte("new different content")
	err = os.WriteFile(sourceFile, sourceContent, 0644)
	assert.NoError(t, err, "Failed to create source file")

	// Organize the conflicting file
	mediaInfo, err := organizer.OrganizeFile(sourceFile, "IMG_20240315_143022.jpg")
	assert.NoError(t, err, "OrganizeFile should not fail")

	// Should have renamed the new file with (1) format
	renamedFile := filepath.Join(targetDir, "IMG_20240315_143022(1).jpg")
	_, err = os.Stat(renamedFile)
	assert.False(t, os.IsNotExist(err), "Renamed file should exist at expected location")

	// Verify content of renamed file
	renamedContent, err := os.ReadFile(renamedFile)
	assert.NoError(t, err, "Failed to read renamed file")
	assert.Equal(t, string(sourceContent), string(renamedContent), "Renamed file content should match source")

	// Original existing file should remain unchanged
	originalContent, err := os.ReadFile(existingFile)
	assert.NoError(t, err, "Failed to read original file")
	assert.Equal(t, string(existingContent), string(originalContent), "Original file content should remain unchanged")

	// MediaInfo should reflect the renamed file (filename won't be updated by OrganizeFile)
	assert.Equal(t, "IMG_20240315_143022.jpg", mediaInfo.FileName, "Expected original filename in metadata")
}

func TestCheckDuplicate(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	t.Run("no duplicate when no organized files exist", func(t *testing.T) {
		// Create test files
		file1Content := []byte("identical content")
		file1 := filepath.Join(tempDir, "file1.jpg")

		err := os.WriteFile(file1, file1Content, 0644)
		assert.NoError(t, err, "Failed to create file1")

		// Test with a fake MediaInfo that has a date
		date := time.Date(2024, 3, 15, 14, 30, 22, 0, time.UTC)
		info := &MediaInfo{
			FileName:  "test.jpg",
			DateTaken: &date,
		}

		// This should not be a duplicate since no organized files exist yet
		isDuplicate, err := organizer.checkDuplicate(file1, info)
		assert.NoError(t, err, "checkDuplicate should not fail")
		assert.False(t, isDuplicate, "Should not be duplicate when no organized files exist")
	})
}

func TestGetFinalPath(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	targetDir := filepath.Join(tempDir, "test")
	err := os.MkdirAll(targetDir, 0755)
	assert.NoError(t, err, "Failed to create target directory")

	t.Run("no existing file", func(t *testing.T) {
		finalPath := organizer.getFinalPath(targetDir, "test.jpg")
		expectedPath := filepath.Join(targetDir, "test.jpg")
		assert.Equal(t, expectedPath, finalPath, "Expected path to match when no conflict")
	})

	t.Run("with existing file generates unique name", func(t *testing.T) {
		// Create existing file
		existingFile := filepath.Join(targetDir, "test_conflict.jpg")
		err := os.WriteFile(existingFile, []byte("existing"), 0644)
		assert.NoError(t, err, "Failed to create existing file")

		// Test with existing file - should generate unique name with (1) format
		finalPath := organizer.getFinalPath(targetDir, "test_conflict.jpg")
		expectedPath := filepath.Join(targetDir, "test_conflict(1).jpg")
		assert.Equal(t, expectedPath, finalPath, "Expected path with (1) suffix when conflict exists")
	})
}

func TestCalculateFileHash(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	t.Run("calculates correct checksum", func(t *testing.T) {
		testFile := filepath.Join(tempDir, "test.txt")
		testContent := []byte("test content for checksum")

		err := os.WriteFile(testFile, testContent, 0644)
		assert.NoError(t, err, "Failed to create test file")

		checksum, err := organizer.calculateFileHash(testFile)
		assert.NoError(t, err, "calculateFileHash should not fail")

		// Calculate expected checksum
		expectedChecksum := fmt.Sprintf("%x", sha256.Sum256(testContent))
		assert.Equal(t, expectedChecksum, checksum, "Expected checksum to match")
	})
}

func TestOrganizeFileNonExistentSource(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	_, err := organizer.OrganizeFile("/non/existent/file.jpg", "test.jpg")
	assert.Error(t, err, "Expected error for non-existent source file")
}

func TestOrganizeFileWithUnknownDate(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	// Create a file with no recognizable date pattern
	sourceFile := filepath.Join(tempDir, "source", "random_name.jpg")
	err := os.MkdirAll(filepath.Dir(sourceFile), 0755)
	assert.NoError(t, err, "Failed to create source directory")

	err = os.WriteFile(sourceFile, []byte("content"), 0644)
	assert.NoError(t, err, "Failed to create source file")

	mediaInfo, err := organizer.OrganizeFile(sourceFile, "random_name.jpg")
	assert.NoError(t, err, "OrganizeFile should not fail")

	// Should fall back to file time
	assert.Equal(t, DateSourceFileTime, mediaInfo.DateSource, "Expected date source to be file time")

	// File should be organized based on file modification time
	assert.NotNil(t, mediaInfo.DateTaken, "Date taken should be set even when falling back to file time")

	// Should be placed in a year/month directory based on file time
	expectedPattern := filepath.Join(tempDir, "*", "*", "random_name.jpg")
	matches, err := filepath.Glob(expectedPattern)
	assert.NoError(t, err, "Failed to glob for organized file")
	assert.Len(t, matches, 1, "Expected exactly one organized file")
}

func TestGetDirectoryStructure(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	t.Run("returns correct directory structure", func(t *testing.T) {
		// Create some test files
		testFiles := []struct {
			path    string
			content string
		}{
			{"2024/March/IMG_20240315_143022.jpg", "content1"},
			{"2024/March/IMG_20240315_150000.jpg", "content2"},
			{"2023/December/VID_20231225_120000.mp4", "content3"},
		}

		for _, file := range testFiles {
			fullPath := filepath.Join(tempDir, file.path)
			err := os.MkdirAll(filepath.Dir(fullPath), 0755)
			assert.NoError(t, err, "Failed to create directory for %s", file.path)

			err = os.WriteFile(fullPath, []byte(file.content), 0644)
			assert.NoError(t, err, "Failed to create file %s", file.path)
		}

		structure, err := organizer.GetDirectoryStructure()
		assert.NoError(t, err, "GetDirectoryStructure should not fail")
		assert.NotNil(t, structure, "Expected non-nil directory structure")

		// Verify structure contains expected years
		assert.Contains(t, structure, "2024", "Expected 2024 to exist in directory structure")
		assert.Contains(t, structure, "2023", "Expected 2023 to exist in directory structure")
	})
}
