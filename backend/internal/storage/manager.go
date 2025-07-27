package storage

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"

	"github.com/Steven-harris/sortify/backend/internal/media"
)

// Manager handles file organization and storage
type Manager struct {
	mediaPath string
	extractor *media.Extractor
}

// NewManager creates a new storage manager
func NewManager(mediaPath string) *Manager {
	return &Manager{
		mediaPath: mediaPath,
		extractor: media.NewExtractor(),
	}
}

// OrganizeFile moves a file from temporary location to organized storage
func (m *Manager) OrganizeFile(tempPath string, originalFilename string) (*media.MediaInfo, error) {
	// Extract metadata from the file
	mediaInfo, err := m.extractor.ExtractMetadata(tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	// Update filename if needed
	if originalFilename != "" {
		mediaInfo.FileName = originalFilename
	}

	// Determine the target directory based on date
	targetDir := m.getTargetDirectory(mediaInfo.DateTaken)

	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Check for duplicates and get final filename
	finalPath, isDuplicate, err := m.getFinalPath(targetDir, mediaInfo.FileName, tempPath)
	if err != nil {
		return nil, fmt.Errorf("failed to determine final path: %w", err)
	}

	if isDuplicate {
		slog.Info("Duplicate file detected, skipping copy", 
			"original", mediaInfo.FileName, 
			"existing", finalPath,
		)
		// Clean up temp file
		os.Remove(tempPath)
		return mediaInfo, nil
	}

	// Move file to final location
	if err := m.moveFile(tempPath, finalPath); err != nil {
		return nil, fmt.Errorf("failed to move file: %w", err)
	}

	// Update media info with final path
	mediaInfo.FileName = filepath.Base(finalPath)

	slog.Info("File organized successfully",
		"original", originalFilename,
		"final_path", finalPath,
		"date_taken", mediaInfo.DateTaken,
		"date_source", mediaInfo.DateSource,
	)

	return mediaInfo, nil
}

// getTargetDirectory returns the target directory path based on date
func (m *Manager) getTargetDirectory(dateTaken *time.Time) string {
	if dateTaken == nil {
		// Use current date if no date available
		now := time.Now()
		dateTaken = &now
	}

	year := fmt.Sprintf("%04d", dateTaken.Year())
	month := fmt.Sprintf("%02d", dateTaken.Month())
	
	return filepath.Join(m.mediaPath, year, month)
}

// getFinalPath determines the final file path, handling duplicates
func (m *Manager) getFinalPath(targetDir, filename, tempPath string) (string, bool, error) {
	basePath := filepath.Join(targetDir, filename)
	
	// If file doesn't exist, we can use the original name
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return basePath, false, nil
	}

	// Check if it's a duplicate by comparing checksums
	tempChecksum, err := m.calculateChecksum(tempPath)
	if err != nil {
		return "", false, fmt.Errorf("failed to calculate temp file checksum: %w", err)
	}

	existingChecksum, err := m.calculateChecksum(basePath)
	if err != nil {
		return "", false, fmt.Errorf("failed to calculate existing file checksum: %w", err)
	}

	// If checksums match, it's a duplicate
	if tempChecksum == existingChecksum {
		return basePath, true, nil
	}

	// Find a unique filename by appending a number
	ext := filepath.Ext(filename)
	nameWithoutExt := filename[:len(filename)-len(ext)]
	
	for i := 1; i < 1000; i++ {
		newFilename := fmt.Sprintf("%s(%d)%s", nameWithoutExt, i, ext)
		newPath := filepath.Join(targetDir, newFilename)
		
		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath, false, nil
		}

		// Check if this variant is also a duplicate
		variantChecksum, err := m.calculateChecksum(newPath)
		if err != nil {
			continue // Skip this variant and try the next
		}

		if tempChecksum == variantChecksum {
			return newPath, true, nil
		}
	}

	return "", false, fmt.Errorf("could not find unique filename after 1000 attempts")
}

// moveFile moves a file from source to destination
func (m *Manager) moveFile(src, dst string) error {
	// Try to rename first (fastest if on same filesystem)
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// If rename fails, copy and delete
	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// Copy file contents
	if _, err := io.Copy(dstFile, srcFile); err != nil {
		os.Remove(dst) // Clean up partial file
		return fmt.Errorf("failed to copy file contents: %w", err)
	}

	// Sync to ensure data is written
	if err := dstFile.Sync(); err != nil {
		os.Remove(dst)
		return fmt.Errorf("failed to sync file: %w", err)
	}

	// Remove source file
	if err := os.Remove(src); err != nil {
		slog.Warn("Failed to remove source file", "error", err, "file", src)
	}

	return nil
}

// calculateChecksum calculates SHA256 checksum of a file
func (m *Manager) calculateChecksum(filePath string) (string, error) {
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

// GetFileInfo returns metadata for an organized file
func (m *Manager) GetFileInfo(relativePath string) (*media.MediaInfo, error) {
	fullPath := filepath.Join(m.mediaPath, relativePath)
	return m.extractor.ExtractMetadata(fullPath)
}

// ListFiles returns files in a directory organized by date
func (m *Manager) ListFiles(year, month string) ([]string, error) {
	dirPath := filepath.Join(m.mediaPath, year, month)
	
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		if os.IsNotExist(err) {
			return []string{}, nil // Return empty list if directory doesn't exist
		}
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() {
			files = append(files, entry.Name())
		}
	}

	return files, nil
}

// GetAvailableDates returns all available year/month combinations
func (m *Manager) GetAvailableDates() ([]DateInfo, error) {
	var dates []DateInfo

	yearEntries, err := os.ReadDir(m.mediaPath)
	if err != nil {
		if os.IsNotExist(err) {
			return dates, nil
		}
		return nil, fmt.Errorf("failed to read media directory: %w", err)
	}

	for _, yearEntry := range yearEntries {
		if !yearEntry.IsDir() || yearEntry.Name() == "temp" {
			continue
		}

		yearPath := filepath.Join(m.mediaPath, yearEntry.Name())
		monthEntries, err := os.ReadDir(yearPath)
		if err != nil {
			continue
		}

		for _, monthEntry := range monthEntries {
			if !monthEntry.IsDir() {
				continue
			}

			dates = append(dates, DateInfo{
				Year:  yearEntry.Name(),
				Month: monthEntry.Name(),
			})
		}
	}

	return dates, nil
}

// DateInfo represents a year/month combination
type DateInfo struct {
	Year  string `json:"year"`
	Month string `json:"month"`
}
