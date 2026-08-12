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
	assert.NotNil(t, extractor)
	assert.NotEmpty(t, extractor.filenamePatterns)
}

func TestExtractDateFromFilename(t *testing.T) {
	extractor := NewExtractor()

	tests := []struct {
		name     string
		filename string
		expected *time.Time
	}{
		{"img pattern", "IMG_20240315_143022.jpg", timePtr(time.Date(2024, 3, 15, 14, 30, 22, 0, time.UTC))},
		{"vid pattern", "VID_20231225_120000.mp4", timePtr(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC))},
		{"screenshot pattern", "Screenshot_2020-01-01-10-30-45.png", timePtr(time.Date(2020, 1, 1, 10, 30, 45, 0, time.UTC))},
		{"hyphenated datetime", "2021-06-14_16-45-12.png", timePtr(time.Date(2021, 6, 14, 16, 45, 12, 0, time.UTC))},
		{"plain date", "2023-12-25.jpg", timePtr(time.Date(2023, 12, 25, 0, 0, 0, 0, time.UTC))},
		{"unknown", "random_filename.jpg", nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			info := &MediaInfo{FileName: tc.filename, ExtraMetadata: map[string]string{}}
			extractor.extractDateFromFilename(tc.filename, info)

			if tc.expected == nil {
				assert.Nil(t, info.DateTaken)
				return
			}

			assert.NotNil(t, info.DateTaken)
			assert.True(t, info.DateTaken.Equal(*tc.expected))
			assert.Equal(t, DateSourceFileName, info.DateSource)
		})
	}
}

func TestExtractMetadataPrefersOriginalFilenameOverTempFilename(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "upload-session-123.tmp")
	err := os.WriteFile(testFile, []byte("fake video content"), 0644)
	assert.NoError(t, err)

	extractor := NewExtractor()
	metadata, err := extractor.ExtractMetadataWithOriginalName(testFile, "VID_20231225_120000.mp4")
	assert.NoError(t, err)

	assert.NotNil(t, metadata.DateTaken)
	assert.True(t, metadata.DateTaken.Equal(time.Date(2023, 12, 25, 12, 0, 0, 0, time.UTC)))
	assert.Equal(t, DateSourceFileName, metadata.DateSource)
	assert.Equal(t, MediaTypeVideo, metadata.MediaType)
	assert.Equal(t, "VID_20231225_120000.mp4", metadata.OriginalFileName)
}

func TestExtractMetadataFallsBackToFileTime(t *testing.T) {
	tempDir := t.TempDir()
	testFile := filepath.Join(tempDir, "random_name.jpg")

	err := os.WriteFile(testFile, []byte("fake content"), 0644)
	assert.NoError(t, err)

	extractor := NewExtractor()
	metadata, err := extractor.ExtractMetadata(testFile)
	assert.NoError(t, err)

	fileInfo, err := os.Stat(testFile)
	assert.NoError(t, err)
	assert.NotNil(t, metadata.DateTaken)
	assert.LessOrEqual(t, metadata.DateTaken.Sub(fileInfo.ModTime()).Abs(), time.Second)
	assert.Equal(t, DateSourceFileTime, metadata.DateSource)
}

func TestDetermineMediaTypeUsesExtensionFallback(t *testing.T) {
	extractor := NewExtractor()
	assert.Equal(t, MediaTypePhoto, extractor.determineMediaType("photo.heic", ""))
	assert.Equal(t, MediaTypeVideo, extractor.determineMediaType("clip.mp4", ""))
	assert.Equal(t, MediaTypeOther, extractor.determineMediaType("doc.txt", ""))
}

func TestParseFilenameMatches(t *testing.T) {
	extractor := NewExtractor()

	assert.NotNil(t, extractor.parseFilenameMatches([]string{"IMG_20240315_143022", "2024", "03", "15", "14", "30", "22"}))
	assert.NotNil(t, extractor.parseFilenameMatches([]string{"20231225", "2023", "12", "25"}))
	assert.Nil(t, extractor.parseFilenameMatches([]string{"invalid"}))
	assert.Nil(t, extractor.parseFilenameMatches([]string{"bad", "2024", "02", "31"}))
}

func TestNeedsUserInput(t *testing.T) {
	extractor := NewExtractor()
	assert.False(t, extractor.NeedsUserInput(&MediaInfo{DateSource: DateSourceEmbeddedMetadata}))
	assert.False(t, extractor.NeedsUserInput(&MediaInfo{DateSource: DateSourceFileName}))
	assert.True(t, extractor.NeedsUserInput(&MediaInfo{DateSource: DateSourceFileTime}))
	assert.False(t, extractor.NeedsUserInput(&MediaInfo{DateSource: DateSourceUserInput}))
	assert.True(t, extractor.NeedsUserInput(&MediaInfo{DateSource: DateSourceUnknown}))
}

func timePtr(t time.Time) *time.Time {
	return &t
}
