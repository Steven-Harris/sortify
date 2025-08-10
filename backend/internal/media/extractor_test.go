package media

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewExtractor(t *testing.T) {
	extractor := NewExtractor()
	assert.NotNil(t, extractor, "NewExtractor should not return nil")
	assert.NotEmpty(t, extractor.filenamePatterns, "Expected filename patterns to be initialized")
}

func TestExtractDateFromFilename(t *testing.T) {
	extractor := NewExtractor()

	t.Run("IMG_20240315_143022.jpg", func(t *testing.T) {
		filename := "IMG_20240315_143022.jpg"
		expectedDate := time.Date(2024, 3, 15, 14, 30, 22, 0, time.UTC)

		info := &MediaInfo{
			FileName:      filename,
			ExtraMetadata: make(map[string]string),
		}

		extractor.extractDateFromFilename(filename, info)

		assert.NotNil(t, info.DateTaken, "Expected date to be extracted")
		assert.True(t, info.DateTaken.Equal(expectedDate), "Expected date %v, got %v", expectedDate, info.DateTaken)
		assert.Equal(t, DateSourceFileName, info.DateSource, "Expected date source to be filename")
	})

	t.Run("VID_20231225_120000.mp4", func(t *testing.T) {
		filename := "VID_20231225_120000.mp4"
		expectedDate := time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC)

		info := &MediaInfo{
			FileName:      filename,
			ExtraMetadata: make(map[string]string),
		}

		extractor.extractDateFromFilename(filename, info)

		assert.NotNil(t, info.DateTaken, "Expected date to be extracted")
		assert.True(t, info.DateTaken.Equal(expectedDate), "Expected date %v, got %v", expectedDate, info.DateTaken)
		assert.Equal(t, DateSourceFileName, info.DateSource, "Expected date source to be filename")
	})

	t.Run("2021-06-14_16-45-12.png", func(t *testing.T) {
		filename := "2021-06-14_16-45-12.png"
		expectedDate := time.Date(2021, 6, 14, 16, 45, 12, 0, time.UTC)

		info := &MediaInfo{
			FileName:      filename,
			ExtraMetadata: make(map[string]string),
		}

		extractor.extractDateFromFilename(filename, info)

		assert.NotNil(t, info.DateTaken, "Expected date to be extracted")
		assert.True(t, info.DateTaken.Equal(expectedDate), "Expected date %v, got %v", expectedDate, info.DateTaken)
		assert.Equal(t, DateSourceFileName, info.DateSource, "Expected date source to be filename")
	})

	t.Run("Screenshot_2020-01-01-10-30-45.png", func(t *testing.T) {
		filename := "Screenshot_2020-01-01-10-30-45.png"
		expectedDate := time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC) // The pattern only extracts date, not time

		info := &MediaInfo{
			FileName:      filename,
			ExtraMetadata: make(map[string]string),
		}

		extractor.extractDateFromFilename(filename, info)

		assert.NotNil(t, info.DateTaken, "Expected date to be extracted")
		assert.True(t, info.DateTaken.Equal(expectedDate), "Expected date %v, got %v", expectedDate, info.DateTaken)
		assert.Equal(t, DateSourceFileName, info.DateSource, "Expected date source to be filename")
	})

	t.Run("random_filename.jpg", func(t *testing.T) {
		filename := "random_filename.jpg"

		info := &MediaInfo{
			FileName:      filename,
			ExtraMetadata: make(map[string]string),
		}

		extractor.extractDateFromFilename(filename, info)

		assert.Nil(t, info.DateTaken, "Expected no date to be extracted")
	})
}

func TestExtractMetadataFromJPEG(t *testing.T) {
	// Create a test JPEG file without EXIF data
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "IMG_20240315_143022.jpg")

	// Create a simple test file (not a real JPEG, but sufficient for filename parsing)
	err := os.WriteFile(testFile, []byte("fake jpeg content"), 0644)
	assert.NoError(t, err, "Failed to create test file")

	extractor := NewExtractor()
	metadata, err := extractor.ExtractMetadata(testFile)
	assert.NoError(t, err, "ExtractMetadata should not fail")

	// Should extract date from filename since no EXIF data
	expectedDate := time.Date(2024, 3, 15, 14, 30, 22, 0, time.UTC)
	assert.NotNil(t, metadata.DateTaken, "Expected date to be extracted")
	assert.True(t, metadata.DateTaken.Equal(expectedDate), "Expected date %v, got %v", expectedDate, metadata.DateTaken)

	if metadata.Camera != nil {
		assert.Empty(t, metadata.Camera.Make, "Expected empty camera make")
	}

	if metadata.Location != nil {
		assert.Zero(t, metadata.Location.Latitude, "Expected zero latitude")
		assert.Zero(t, metadata.Location.Longitude, "Expected zero longitude")
	}

	assert.Equal(t, DateSourceFileName, metadata.DateSource, "Expected date source to be filename")
	assert.Equal(t, MediaTypePhoto, metadata.MediaType, "Expected media type to be photo")
}

func TestExtractMetadataFallbackToFileTime(t *testing.T) {
	// Create a test file with no date in filename
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "random_name.jpg")

	err := os.WriteFile(testFile, []byte("fake content"), 0644)
	assert.NoError(t, err, "Failed to create test file")

	extractor := NewExtractor()
	metadata, err := extractor.ExtractMetadata(testFile)
	assert.NoError(t, err, "ExtractMetadata should not fail")

	// Should fall back to file modification time
	fileInfo, err := os.Stat(testFile)
	assert.NoError(t, err, "Failed to stat file")

	assert.NotNil(t, metadata.DateTaken, "Expected date taken to be set")

	// Allow some tolerance for file time comparison (within 1 second)
	timeDiff := metadata.DateTaken.Sub(fileInfo.ModTime()).Abs()
	assert.LessOrEqual(t, timeDiff, time.Second, "Date taken should be close to file mod time")

	assert.Equal(t, DateSourceFileTime, metadata.DateSource, "Expected date source to be file time")
}

func TestExtractMetadataVideoFile(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "VID_20231225_120000.mp4")

	err := os.WriteFile(testFile, []byte("fake video content"), 0644)
	assert.NoError(t, err, "Failed to create test file")

	extractor := NewExtractor()
	metadata, err := extractor.ExtractMetadata(testFile)
	assert.NoError(t, err, "ExtractMetadata should not fail")

	expectedDate := time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC)
	assert.NotNil(t, metadata.DateTaken, "Expected date to be extracted")
	assert.True(t, metadata.DateTaken.Equal(expectedDate), "Expected date %v, got %v", expectedDate, metadata.DateTaken)

	assert.Equal(t, DateSourceFileName, metadata.DateSource, "Expected date source to be filename")
	assert.Equal(t, MediaTypeVideo, metadata.MediaType, "Expected media type to be video")
}

func TestExtractMetadataNonExistentFile(t *testing.T) {
	extractor := NewExtractor()
	_, err := extractor.ExtractMetadata("/non/existent/file.jpg")
	assert.Error(t, err, "Expected error for non-existent file")
}

func TestDetermineMediaType(t *testing.T) {
	extractor := NewExtractor()

	t.Run("image/jpeg", func(t *testing.T) {
		mediaType := extractor.determineMediaType("image/jpeg")
		assert.Equal(t, MediaTypePhoto, mediaType, "Expected photo media type for JPEG")
	})

	t.Run("image/png", func(t *testing.T) {
		mediaType := extractor.determineMediaType("image/png")
		assert.Equal(t, MediaTypePhoto, mediaType, "Expected photo media type for PNG")
	})

	t.Run("image/gif", func(t *testing.T) {
		mediaType := extractor.determineMediaType("image/gif")
		assert.Equal(t, MediaTypePhoto, mediaType, "Expected photo media type for GIF")
	})

	t.Run("video/mp4", func(t *testing.T) {
		mediaType := extractor.determineMediaType("video/mp4")
		assert.Equal(t, MediaTypeVideo, mediaType, "Expected video media type for MP4")
	})

	t.Run("video/quicktime", func(t *testing.T) {
		mediaType := extractor.determineMediaType("video/quicktime")
		assert.Equal(t, MediaTypeVideo, mediaType, "Expected video media type for QuickTime")
	})

	t.Run("application/pdf", func(t *testing.T) {
		mediaType := extractor.determineMediaType("application/pdf")
		assert.Equal(t, MediaTypeOther, mediaType, "Expected other media type for PDF")
	})

	t.Run("text/plain", func(t *testing.T) {
		mediaType := extractor.determineMediaType("text/plain")
		assert.Equal(t, MediaTypeOther, mediaType, "Expected other media type for text")
	})

	t.Run("empty mime type", func(t *testing.T) {
		mediaType := extractor.determineMediaType("")
		assert.Equal(t, MediaTypeOther, mediaType, "Expected other media type for empty string")
	})
}

func TestBuildFilenamePatterns(t *testing.T) {
	patterns := buildFilenamePatterns()

	assert.NotEmpty(t, patterns, "Expected non-empty patterns slice")

	// Test that each pattern compiles
	for i, pattern := range patterns {
		assert.NotNil(t, pattern, "Pattern %d should not be nil", i)
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

	assert.True(t, matched, "Expected at least one pattern to match IMG_20240315_143022 format")
}

func TestParseFilenameMatches(t *testing.T) {
	extractor := NewExtractor()

	t.Run("full timestamp match", func(t *testing.T) {
		matches := []string{"IMG_20240315_143022", "2024", "03", "15", "14", "30", "22"}
		expected := time.Date(2024, 3, 15, 14, 30, 22, 0, time.UTC)

		result := extractor.parseFilenameMatches(matches)

		assert.NotNil(t, result, "Expected parsed time to not be nil")
		assert.True(t, result.Equal(expected), "Expected %v, got %v", expected, result)
	})

	t.Run("date only match", func(t *testing.T) {
		matches := []string{"20231225", "2023", "12", "25"}
		expected := time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC)

		result := extractor.parseFilenameMatches(matches)

		assert.NotNil(t, result, "Expected parsed time to not be nil")
		assert.True(t, result.Equal(expected), "Expected %v, got %v", expected, result)
	})

	t.Run("invalid single match", func(t *testing.T) {
		matches := []string{"invalid"}

		result := extractor.parseFilenameMatches(matches)

		assert.Nil(t, result, "Expected nil for invalid match")
	})

	t.Run("invalid multiple matches", func(t *testing.T) {
		matches := []string{"", "invalid", "date", "parts"}

		result := extractor.parseFilenameMatches(matches)

		assert.Nil(t, result, "Expected nil for invalid matches")
	})
}

func TestNeedsUserInput(t *testing.T) {
	extractor := NewExtractor()

	t.Run("exif", func(t *testing.T) {
		info := &MediaInfo{DateSource: DateSourceEXIF}
		result := extractor.NeedsUserInput(info)
		assert.False(t, result, "EXIF date source should not need user input")
	})

	t.Run("filename", func(t *testing.T) {
		info := &MediaInfo{DateSource: DateSourceFileName}
		result := extractor.NeedsUserInput(info)
		assert.False(t, result, "Filename date source should not need user input")
	})

	t.Run("fileTime", func(t *testing.T) {
		info := &MediaInfo{DateSource: DateSourceFileTime}
		result := extractor.NeedsUserInput(info)
		assert.True(t, result, "File time date source should need user input")
	})

	t.Run("userInput", func(t *testing.T) {
		info := &MediaInfo{DateSource: DateSourceUserInput}
		result := extractor.NeedsUserInput(info)
		assert.False(t, result, "User input date source should not need user input")
	})

	t.Run("unknown", func(t *testing.T) {
		info := &MediaInfo{DateSource: DateSourceUnknown}
		result := extractor.NeedsUserInput(info)
		assert.True(t, result, "Unknown date source should need user input")
	})
}
