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
	assert.NotNil(t, organizer)
	assert.Equal(t, tempDir, organizer.mediaPath)
	assert.NotNil(t, organizer.extractor)
}

func TestGetTargetDirectory(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	date := time.Date(2024, 3, 15, 14, 30, 22, 0, time.UTC)
	result, err := organizer.getTargetDirectory(&date)
	assert.NoError(t, err)
	assert.Equal(t, filepath.Join(tempDir, "2024", "March"), result)
}

func TestOrganizeFileUsesOriginalFilenameForDateExtraction(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	sourceFile := filepath.Join(tempDir, "source", "upload-blob.tmp")
	assert.NoError(t, os.MkdirAll(filepath.Dir(sourceFile), 0755))
	assert.NoError(t, os.WriteFile(sourceFile, []byte("test image content"), 0644))

	mediaInfo, err := organizer.OrganizeFile(sourceFile, "IMG_20240315_143022.jpg")
	assert.NoError(t, err)
	assert.Equal(t, DateSourceFileName, mediaInfo.DateSource)
	assert.NotNil(t, mediaInfo.DateTaken)
	assert.True(t, mediaInfo.DateTaken.Equal(time.Date(2024, 3, 15, 14, 30, 22, 0, time.UTC)))
	assert.FileExists(t, filepath.Join(tempDir, "2024", "March", "IMG_20240315_143022.jpg"))
}

func TestOrganizeFileDeduplicatesAcrossLibrary(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	existingDir := filepath.Join(tempDir, "2020", "January")
	assert.NoError(t, os.MkdirAll(existingDir, 0755))
	existingFile := filepath.Join(existingDir, "old-name.jpg")
	content := []byte("same-content")
	assert.NoError(t, os.WriteFile(existingFile, content, 0644))

	sourceFile := filepath.Join(tempDir, "source", "incoming.tmp")
	assert.NoError(t, os.MkdirAll(filepath.Dir(sourceFile), 0755))
	assert.NoError(t, os.WriteFile(sourceFile, content, 0644))

	mediaInfo, err := organizer.OrganizeFile(sourceFile, "IMG_20240315_143022.jpg")
	assert.NoError(t, err)
	assert.True(t, mediaInfo.IsDuplicate)
	assert.Equal(t, existingFile, mediaInfo.CanonicalPath)
	assert.NoFileExists(t, sourceFile)
	assert.FileExists(t, existingFile)
	assert.NoFileExists(t, filepath.Join(tempDir, "2024", "March", "IMG_20240315_143022.jpg"))
}

func TestOrganizeFileUsesDeterministicConflictNameForDifferentContent(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	targetDir := filepath.Join(tempDir, "2024", "March")
	assert.NoError(t, os.MkdirAll(targetDir, 0755))

	existingFile := filepath.Join(targetDir, "IMG_20240315_143022.jpg")
	assert.NoError(t, os.WriteFile(existingFile, []byte("existing"), 0644))

	sourceFile := filepath.Join(tempDir, "source", "incoming.tmp")
	assert.NoError(t, os.MkdirAll(filepath.Dir(sourceFile), 0755))
	assert.NoError(t, os.WriteFile(sourceFile, []byte("different"), 0644))

	mediaInfo, err := organizer.OrganizeFile(sourceFile, "IMG_20240315_143022.jpg")
	assert.NoError(t, err)
	assert.Equal(t, "IMG_20240315_143022 (1).jpg", mediaInfo.FileName)
	assert.Equal(t, filepath.Join(targetDir, "IMG_20240315_143022 (1).jpg"), mediaInfo.CanonicalPath)
	assert.FileExists(t, mediaInfo.CanonicalPath)
}

func TestCheckHashCalculation(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	testFile := filepath.Join(tempDir, "test.txt")
	content := []byte("test content for checksum")
	assert.NoError(t, os.WriteFile(testFile, content, 0644))

	checksum, err := organizer.calculateFileHash(testFile)
	assert.NoError(t, err)
	assert.Equal(t, fmt.Sprintf("%x", sha256.Sum256(content)), checksum)
}

func TestGetFinalPathConflict(t *testing.T) {
	tempDir := t.TempDir()
	organizer := NewOrganizer(tempDir)

	targetDir := filepath.Join(tempDir, "test")
	assert.NoError(t, os.MkdirAll(targetDir, 0755))
	assert.NoError(t, os.WriteFile(filepath.Join(targetDir, "test.jpg"), []byte("existing"), 0644))

	finalPath := organizer.getFinalPath(targetDir, "test.jpg")
	assert.Equal(t, filepath.Join(targetDir, "test (1).jpg"), finalPath)
}
