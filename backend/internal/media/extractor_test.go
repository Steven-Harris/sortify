package media

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestNewExtractor(t *testing.T) {
	extractor := NewExtractor()
	if extractor == nil {
		t.Fatal("NewExtractor should not return nil")
	}

	if len(extractor.filenamePatterns) == 0 {
		t.Error("Expected filename patterns to be initialized")
	}
}

func TestExtractDateFromFilename(t *testing.T) {
	tests := []struct {
		filename     string
		expectedDate *time.Time
		hasDate      bool
	}{
		{
			filename:     "IMG_20240315_143022.jpg",
			expectedDate: timePtr(time.Date(2024, 3, 15, 14, 30, 22, 0, time.UTC)),
			hasDate:      true,
		},
		{
			filename:     "VID_20231225_120000.mp4",
			expectedDate: timePtr(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC)),
			hasDate:      true,
		},
		{
			filename:     "2021-06-14_16-45-12.png",
			expectedDate: timePtr(time.Date(2021, 6, 14, 16, 45, 12, 0, time.UTC)),
			hasDate:      true,
		},
		{
			filename:     "Screenshot_2020-01-01-10-30-45.png",
			expectedDate: timePtr(time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC)), // The pattern only extracts date, not time
			hasDate:      true,
		},
		{
			filename:     "random_filename.jpg",
			expectedDate: nil,
			hasDate:      false,
		},
	}

	extractor := NewExtractor()

	for _, test := range tests {
		t.Run(test.filename, func(t *testing.T) {
			info := &MediaInfo{
				FileName:      test.filename,
				ExtraMetadata: make(map[string]string),
			}

			extractor.extractDateFromFilename(test.filename, info)

			if test.hasDate {
				if info.DateTaken == nil {
					t.Error("Expected date to be extracted, got nil")
					return
				}
				if !info.DateTaken.Equal(*test.expectedDate) {
					t.Errorf("Expected date %v, got %v", test.expectedDate, info.DateTaken)
				}
				if info.DateSource != DateSourceFileName {
					t.Errorf("Expected date source %s, got %s", DateSourceFileName, info.DateSource)
				}
			} else {
				if info.DateTaken != nil {
					t.Errorf("Expected no date, got %v", info.DateTaken)
				}
			}
		})
	}
}

func TestExtractMetadataFromJPEG(t *testing.T) {
	// Create a test JPEG file without EXIF data
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "IMG_20240315_143022.jpg")

	// Create a simple test file (not a real JPEG, but sufficient for filename parsing)
	err := os.WriteFile(testFile, []byte("fake jpeg content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	extractor := NewExtractor()
	metadata, err := extractor.ExtractMetadata(testFile)
	if err != nil {
		t.Fatalf("ExtractMetadata failed: %v", err)
	}

	// Should extract date from filename since no EXIF data
	expectedDate := time.Date(2024, 3, 15, 14, 30, 22, 0, time.UTC)
	if metadata.DateTaken == nil || !metadata.DateTaken.Equal(expectedDate) {
		t.Errorf("Expected date %v, got %v", expectedDate, metadata.DateTaken)
	}

	if metadata.Camera != nil && metadata.Camera.Make != "" {
		t.Errorf("Expected empty camera make, got %s", metadata.Camera.Make)
	}

	if metadata.Location != nil && (metadata.Location.Latitude != 0 || metadata.Location.Longitude != 0) {
		t.Errorf("Expected zero coordinates, got lat=%f, lon=%f",
			metadata.Location.Latitude, metadata.Location.Longitude)
	}

	if metadata.DateSource != DateSourceFileName {
		t.Errorf("Expected date source %s, got %s", DateSourceFileName, metadata.DateSource)
	}

	if metadata.MediaType != MediaTypePhoto {
		t.Errorf("Expected media type %s, got %s", MediaTypePhoto, metadata.MediaType)
	}
}

func TestExtractMetadataFallbackToFileTime(t *testing.T) {
	// Create a test file with no date in filename
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "random_name.jpg")

	err := os.WriteFile(testFile, []byte("fake content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	extractor := NewExtractor()
	metadata, err := extractor.ExtractMetadata(testFile)
	if err != nil {
		t.Fatalf("ExtractMetadata failed: %v", err)
	}

	// Should fall back to file modification time
	fileInfo, err := os.Stat(testFile)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	if metadata.DateTaken == nil {
		t.Fatal("Expected date taken to be set")
	}

	// Allow some tolerance for file time comparison (within 1 second)
	timeDiff := metadata.DateTaken.Sub(fileInfo.ModTime()).Abs()
	if timeDiff > time.Second {
		t.Errorf("Date taken should be close to file mod time. Diff: %v", timeDiff)
	}

	if metadata.DateSource != DateSourceFileTime {
		t.Errorf("Expected date source %s, got %s", DateSourceFileTime, metadata.DateSource)
	}
}

func TestExtractMetadataVideoFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "VID_20231225_120000.mp4")

	err := os.WriteFile(testFile, []byte("fake video content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	extractor := NewExtractor()
	metadata, err := extractor.ExtractMetadata(testFile)
	if err != nil {
		t.Fatalf("ExtractMetadata failed: %v", err)
	}

	expectedDate := time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC)
	if metadata.DateTaken == nil || !metadata.DateTaken.Equal(expectedDate) {
		t.Errorf("Expected date %v, got %v", expectedDate, metadata.DateTaken)
	}

	if metadata.DateSource != DateSourceFileName {
		t.Errorf("Expected date source %s, got %s", DateSourceFileName, metadata.DateSource)
	}

	if metadata.MediaType != MediaTypeVideo {
		t.Errorf("Expected media type %s, got %s", MediaTypeVideo, metadata.MediaType)
	}
}

func TestExtractMetadataNonExistentFile(t *testing.T) {
	extractor := NewExtractor()
	_, err := extractor.ExtractMetadata("/non/existent/file.jpg")
	if err == nil {
		t.Error("Expected error for non-existent file, got nil")
	}
}

func TestDetermineMediaType(t *testing.T) {
	tests := []struct {
		mimeType string
		expected MediaType
	}{
		{"image/jpeg", MediaTypePhoto},
		{"image/png", MediaTypePhoto},
		{"image/gif", MediaTypePhoto},
		{"video/mp4", MediaTypeVideo},
		{"video/quicktime", MediaTypeVideo},
		{"application/pdf", MediaTypeOther},
		{"text/plain", MediaTypeOther},
		{"", MediaTypeOther},
	}

	extractor := NewExtractor()

	for _, test := range tests {
		t.Run(test.mimeType, func(t *testing.T) {
			mediaType := extractor.determineMediaType(test.mimeType)
			if mediaType != test.expected {
				t.Errorf("Expected media type %s, got %s", test.expected, mediaType)
			}
		})
	}
}

func TestBuildFilenamePatterns(t *testing.T) {
	patterns := buildFilenamePatterns()

	if len(patterns) == 0 {
		t.Error("Expected non-empty patterns slice")
	}

	// Test that each pattern compiles
	for i, pattern := range patterns {
		if pattern == nil {
			t.Errorf("Pattern %d should not be nil", i)
		}
	}

	// Test a known pattern matches expected format
	testString := "IMG_20240315_143022"
	matched := false
	for _, pattern := range patterns {
		if pattern.MatchString(testString) {
			matched = true
			break
		}
	}

	if !matched {
		t.Error("Expected at least one pattern to match IMG_20240315_143022 format")
	}
}

func TestParseFilenameMatches(t *testing.T) {
	extractor := NewExtractor()

	tests := []struct {
		matches  []string
		expected *time.Time
	}{
		{
			matches:  []string{"IMG_20240315_143022", "2024", "03", "15", "14", "30", "22"},
			expected: timePtr(time.Date(2024, 3, 15, 14, 30, 22, 0, time.UTC)),
		},
		{
			matches:  []string{"20231225", "2023", "12", "25"},
			expected: timePtr(time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)),
		},
		{
			matches:  []string{"invalid"},
			expected: nil,
		},
		{
			matches:  []string{"", "invalid", "date", "parts"},
			expected: nil,
		},
	}

	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			result := extractor.parseFilenameMatches(test.matches)

			if test.expected == nil {
				if result != nil {
					t.Errorf("Expected nil, got %v", result)
				}
			} else {
				if result == nil {
					t.Errorf("Expected %v, got nil", test.expected)
				} else if !result.Equal(*test.expected) {
					t.Errorf("Expected %v, got %v", test.expected, result)
				}
			}
		})
	}
}

func TestNeedsUserInput(t *testing.T) {
	extractor := NewExtractor()

	tests := []struct {
		dateSource DateSource
		expected   bool
	}{
		{DateSourceEXIF, false},
		{DateSourceFileName, false},
		{DateSourceFileTime, true},
		{DateSourceUserInput, false},
		{DateSourceUnknown, true},
	}

	for _, test := range tests {
		t.Run(string(test.dateSource), func(t *testing.T) {
			info := &MediaInfo{DateSource: test.dateSource}
			result := extractor.NeedsUserInput(info)

			if result != test.expected {
				t.Errorf("Expected %v, got %v", test.expected, result)
			}
		})
	}
}

// Helper function to create time pointer
func timePtr(t time.Time) *time.Time {
	return &t
}
