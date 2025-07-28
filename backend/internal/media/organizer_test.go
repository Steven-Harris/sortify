package media

import (
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestNewOrganizer(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	if organizer == nil {
		t.Fatal("NewOrganizer should not return nil")
	}

	if organizer.mediaPath != tempDir {
		t.Errorf("Expected media path %s, got %s", tempDir, organizer.mediaPath)
	}

	if organizer.extractor == nil {
		t.Error("Expected extractor to be initialized")
	}
}

func TestGetTargetDirectory(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	tests := []struct {
		name     string
		date     *time.Time
		expected string
	}{
		{
			name:     "Valid date",
			date:     timePtr(time.Date(2024, 3, 15, 14, 30, 22, 0, time.UTC)),
			expected: filepath.Join(tempDir, "2024", "March"),
		},
		{
			name:     "Different year and month",
			date:     timePtr(time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)),
			expected: filepath.Join(tempDir, "2023", "December"),
		},
		{
			name:     "Single digit month",
			date:     timePtr(time.Date(2022, 7, 8, 0, 0, 0, 0, time.UTC)),
			expected: filepath.Join(tempDir, "2022", "July"),
		},
		{
			name:     "Nil date",
			date:     nil,
			expected: "uses_current_date", // Will use current time, so we'll test differently
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := organizer.getTargetDirectory(test.date)
			if err != nil {
				t.Fatalf("getTargetDirectory failed: %v", err)
			}

			if test.expected == "uses_current_date" {
				// For nil date, just verify it uses a valid year/month format
				if !filepath.IsAbs(result) {
					t.Errorf("Expected absolute path, got %s", result)
				}
				// Should contain current year
				currentYear := time.Now().Format("2006")
				if !strings.Contains(result, currentYear) {
					t.Errorf("Expected path to contain current year %s, got %s", currentYear, result)
				}
			} else {
				if result != test.expected {
					t.Errorf("Expected %s, got %s", test.expected, result)
				}
			}
		})
	}
}

func TestOrganizeFileSuccess(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	// Create a test file with date in filename
	sourceFile := filepath.Join(tempDir, "source", "IMG_20240315_143022.jpg")
	err := os.MkdirAll(filepath.Dir(sourceFile), 0755)
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	testContent := []byte("test image content")
	err = os.WriteFile(sourceFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	// Organize the file
	mediaInfo, err := organizer.OrganizeFile(sourceFile, "IMG_20240315_143022.jpg")
	if err != nil {
		t.Fatalf("OrganizeFile failed: %v", err)
	}

	// Verify metadata
	expectedDate := time.Date(2024, 3, 15, 14, 30, 22, 0, time.UTC)
	if mediaInfo.DateTaken == nil || !mediaInfo.DateTaken.Equal(expectedDate) {
		t.Errorf("Expected date %v, got %v", expectedDate, mediaInfo.DateTaken)
	}

	if mediaInfo.DateSource != DateSourceFileName {
		t.Errorf("Expected date source %s, got %s", DateSourceFileName, mediaInfo.DateSource)
	}

	if mediaInfo.MediaType != MediaTypePhoto {
		t.Errorf("Expected media type %s, got %s", MediaTypePhoto, mediaInfo.MediaType)
	}

	// Verify file was moved to correct location
	expectedDir := filepath.Join(tempDir, "2024", "March")
	expectedFile := filepath.Join(expectedDir, "IMG_20240315_143022.jpg")

	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("File should exist at %s", expectedFile)
	}

	// Verify file content
	movedContent, err := os.ReadFile(expectedFile)
	if err != nil {
		t.Fatalf("Failed to read moved file: %v", err)
	}

	if string(movedContent) != string(testContent) {
		t.Error("File content should match original")
	}

	// Verify source file was removed
	if _, err := os.Stat(sourceFile); !os.IsNotExist(err) {
		t.Error("Source file should be removed after organizing")
	}
}

func TestOrganizeFileDuplicate(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	// Create target directory and existing file
	targetDir := filepath.Join(tempDir, "2024", "March")
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	existingFile := filepath.Join(targetDir, "IMG_20240315_143022.jpg")
	existingContent := []byte("existing content")
	err = os.WriteFile(existingFile, existingContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	// Create source file with same content (exact duplicate)
	sourceFile := filepath.Join(tempDir, "source", "IMG_20240315_143022.jpg")
	err = os.MkdirAll(filepath.Dir(sourceFile), 0755)
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	err = os.WriteFile(sourceFile, existingContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Organize the duplicate file
	mediaInfo, err := organizer.OrganizeFile(sourceFile, "IMG_20240315_143022.jpg")
	if err != nil {
		t.Fatalf("OrganizeFile failed: %v", err)
	}

	// Should have detected duplicate and removed source without error
	if _, err := os.Stat(sourceFile); !os.IsNotExist(err) {
		t.Error("Source file should be removed after duplicate detection")
	}

	// Original file should still exist
	if _, err := os.Stat(existingFile); os.IsNotExist(err) {
		t.Error("Existing file should remain")
	}

	// Metadata should still be returned
	if mediaInfo == nil {
		t.Error("MediaInfo should be returned even for duplicates")
	}
}

func TestOrganizeFileConflict(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	// Create target directory and existing file with different content
	targetDir := filepath.Join(tempDir, "2024", "March")
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	existingFile := filepath.Join(targetDir, "IMG_20240315_143022.jpg")
	existingContent := []byte("existing different content")
	err = os.WriteFile(existingFile, existingContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	// Create source file with different content
	sourceFile := filepath.Join(tempDir, "source", "IMG_20240315_143022.jpg")
	err = os.MkdirAll(filepath.Dir(sourceFile), 0755)
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	sourceContent := []byte("new different content")
	err = os.WriteFile(sourceFile, sourceContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	// Organize the conflicting file
	mediaInfo, err := organizer.OrganizeFile(sourceFile, "IMG_20240315_143022.jpg")
	if err != nil {
		t.Fatalf("OrganizeFile failed: %v", err)
	}

	// Should have renamed the new file with (1) format
	renamedFile := filepath.Join(targetDir, "IMG_20240315_143022(1).jpg")
	if _, err := os.Stat(renamedFile); os.IsNotExist(err) {
		t.Errorf("Renamed file should exist at %s", renamedFile)
	}

	// Verify content of renamed file
	renamedContent, err := os.ReadFile(renamedFile)
	if err != nil {
		t.Fatalf("Failed to read renamed file: %v", err)
	}

	if string(renamedContent) != string(sourceContent) {
		t.Error("Renamed file content should match source")
	}

	// Original existing file should remain unchanged
	originalContent, err := os.ReadFile(existingFile)
	if err != nil {
		t.Fatalf("Failed to read original file: %v", err)
	}

	if string(originalContent) != string(existingContent) {
		t.Error("Original file content should remain unchanged")
	}

	// MediaInfo should reflect the renamed file (filename won't be updated by OrganizeFile)
	if mediaInfo.FileName != "IMG_20240315_143022.jpg" {
		t.Errorf("Expected original filename in metadata, got %s", mediaInfo.FileName)
	}
}

func TestCheckDuplicate(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	// Create test files
	file1Content := []byte("identical content")
	file2Content := []byte("different content")

	file1 := filepath.Join(tempDir, "file1.jpg")
	file2 := filepath.Join(tempDir, "file2.jpg")

	err := os.WriteFile(file1, file1Content, 0644)
	if err != nil {
		t.Fatalf("Failed to create file1: %v", err)
	}

	err = os.WriteFile(file2, file2Content, 0644)
	if err != nil {
		t.Fatalf("Failed to create file2: %v", err)
	}

	// Test with a fake MediaInfo that has a date
	date := time.Date(2024, 3, 15, 14, 30, 22, 0, time.UTC)
	info := &MediaInfo{
		FileName:  "test.jpg",
		DateTaken: &date,
	}

	// This should not be a duplicate since no organized files exist yet
	isDuplicate, err := organizer.checkDuplicate(file1, info)
	if err != nil {
		t.Fatalf("checkDuplicate failed: %v", err)
	}

	if isDuplicate {
		t.Error("Should not be duplicate when no organized files exist")
	}
}

func TestGetFinalPath(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	targetDir := filepath.Join(tempDir, "test")
	err := os.MkdirAll(targetDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create target directory: %v", err)
	}

	// Test with no existing file
	finalPath := organizer.getFinalPath(targetDir, "test.jpg")
	expectedPath := filepath.Join(targetDir, "test.jpg")
	if finalPath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, finalPath)
	}

	// Create existing file
	existingFile := filepath.Join(targetDir, "test.jpg")
	err = os.WriteFile(existingFile, []byte("existing"), 0644)
	if err != nil {
		t.Fatalf("Failed to create existing file: %v", err)
	}

	// Test with existing file - should generate unique name with (1) format
	finalPath = organizer.getFinalPath(targetDir, "test.jpg")
	expectedPath = filepath.Join(targetDir, "test(1).jpg")
	if finalPath != expectedPath {
		t.Errorf("Expected path %s, got %s", expectedPath, finalPath)
	}
}

func TestCalculateFileHash(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	testContent := []byte("test content for checksum")

	err := os.WriteFile(testFile, testContent, 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	checksum, err := organizer.calculateFileHash(testFile)
	if err != nil {
		t.Fatalf("calculateFileHash failed: %v", err)
	}

	// Calculate expected checksum
	expectedChecksum := fmt.Sprintf("%x", sha256.Sum256(testContent))

	if checksum != expectedChecksum {
		t.Errorf("Expected checksum %s, got %s", expectedChecksum, checksum)
	}
}

func TestOrganizeFileNonExistentSource(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	_, err := organizer.OrganizeFile("/non/existent/file.jpg", "test.jpg")
	if err == nil {
		t.Error("Expected error for non-existent source file, got nil")
	}
}

func TestOrganizeFileWithUnknownDate(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	// Create a file with no recognizable date pattern
	sourceFile := filepath.Join(tempDir, "source", "random_name.jpg")
	err := os.MkdirAll(filepath.Dir(sourceFile), 0755)
	if err != nil {
		t.Fatalf("Failed to create source directory: %v", err)
	}

	err = os.WriteFile(sourceFile, []byte("content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create source file: %v", err)
	}

	mediaInfo, err := organizer.OrganizeFile(sourceFile, "random_name.jpg")
	if err != nil {
		t.Fatalf("OrganizeFile failed: %v", err)
	}

	// Should fall back to file time
	if mediaInfo.DateSource != DateSourceFileTime {
		t.Errorf("Expected date source %s, got %s", DateSourceFileTime, mediaInfo.DateSource)
	}

	// File should be organized based on file modification time
	if mediaInfo.DateTaken == nil {
		t.Error("Date taken should be set even when falling back to file time")
	}

	// Should be placed in a year/month directory based on file time
	expectedPattern := filepath.Join(tempDir, "*", "*", "random_name.jpg")
	matches, err := filepath.Glob(expectedPattern)
	if err != nil {
		t.Fatalf("Failed to glob for organized file: %v", err)
	}

	if len(matches) != 1 {
		t.Errorf("Expected 1 organized file, found %d", len(matches))
	}
}

func TestGetDirectoryStructure(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

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
		if err != nil {
			t.Fatalf("Failed to create directory for %s: %v", file.path, err)
		}

		err = os.WriteFile(fullPath, []byte(file.content), 0644)
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", file.path, err)
		}
	}

	structure, err := organizer.GetDirectoryStructure()
	if err != nil {
		t.Fatalf("GetDirectoryStructure failed: %v", err)
	}

	if structure == nil {
		t.Fatal("Expected non-nil directory structure")
	}

	// Verify structure contains expected years
	if _, exists := structure["2024"]; !exists {
		t.Error("Expected 2024 to exist in directory structure")
	}

	if _, exists := structure["2023"]; !exists {
		t.Error("Expected 2023 to exist in directory structure")
	}
}
