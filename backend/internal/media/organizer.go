package media

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// Organizer handles organizing media files into date-based directory structure
type Organizer struct {
	mediaPath string
	extractor *Extractor
}

// NewOrganizer creates a new media organizer
func NewOrganizer(mediaPath string) *Organizer {
	return &Organizer{
		mediaPath: mediaPath,
		extractor: NewExtractor(),
	}
}

// OrganizeFile processes and organizes a completed upload
func (o *Organizer) OrganizeFile(tempFilePath, originalFileName string) (*MediaInfo, error) {
	// Extract metadata from the temporary file
	info, err := o.extractor.ExtractMetadata(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	// Update filename to original name for proper date extraction
	info.FileName = originalFileName

	// If date was extracted from filename and original filename is different,
	// try to re-extract from the original filename
	tempFileName := filepath.Base(tempFilePath)
	if info.DateSource == "filename" && tempFileName != originalFileName {
		// Clear the date extracted from temp filename
		info.DateTaken = nil
		info.DateSource = ""

		// Try to extract from original filename
		o.extractor.ExtractDateFromFilename(originalFileName, info)

		// If still no date found, fall back to file time
		if info.DateTaken == nil {
			if fileInfo, err := os.Stat(tempFilePath); err == nil {
				if fileInfo.ModTime().Year() > 1970 { // Reasonable date check
					info.DateTaken = &[]time.Time{fileInfo.ModTime()}[0]
					info.DateSource = "file_time"
				}
			}
		}
	} // Check for duplicates
	if duplicate, err := o.checkDuplicate(tempFilePath, info); err != nil {
		slog.Error("Failed to check for duplicates", "error", err, "file", originalFileName)
	} else if duplicate {
		slog.Info("Duplicate file detected, skipping", "file", originalFileName)
		os.Remove(tempFilePath) // Clean up temp file
		return info, nil
	}

	// Determine target directory based on date
	targetDir, err := o.getTargetDirectory(info.DateTaken)
	if err != nil {
		return nil, fmt.Errorf("failed to determine target directory: %w", err)
	}

	// Ensure target directory exists
	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	// Determine final file path (handle naming conflicts)
	finalPath := o.getFinalPath(targetDir, originalFileName)

	// Move file from temp location to final location
	if err := o.moveFile(tempFilePath, finalPath); err != nil {
		return nil, fmt.Errorf("failed to move file: %w", err)
	}

	slog.Info("File organized successfully",
		"original_file", originalFileName,
		"final_path", finalPath,
		"date_taken", info.DateTaken,
		"date_source", info.DateSource,
	)

	return info, nil
}

// checkDuplicate checks if a file with the same content already exists
func (o *Organizer) checkDuplicate(filePath string, info *MediaInfo) (bool, error) {
	// Calculate file hash
	hash, err := o.calculateFileHash(filePath)
	if err != nil {
		return false, err
	}

	// Search for files with the same hash in the organized directories
	// This is a simplified approach - in production, you might want to use a database
	targetDir, err := o.getTargetDirectory(info.DateTaken)
	if err != nil {
		return false, err
	}

	// Check if target directory exists
	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return false, nil // Directory doesn't exist, so no duplicates
	}

	// Walk through files in the target directory
	var foundDuplicate bool
	err = filepath.Walk(targetDir, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue walking even if there's an error with a specific file
		}

		if fileInfo.IsDir() {
			return nil
		}

		// Calculate hash of existing file
		existingHash, err := o.calculateFileHash(path)
		if err != nil {
			return nil // Continue even if we can't hash this file
		}

		if existingHash == hash {
			foundDuplicate = true
			slog.Info("Duplicate found", "original", filePath, "existing", path)
			return filepath.SkipAll // Stop walking
		}

		return nil
	})

	return foundDuplicate, err
}

// calculateFileHash calculates SHA256 hash of a file
func (o *Organizer) calculateFileHash(filePath string) (string, error) {
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

// getTargetDirectory determines the target directory based on date
func (o *Organizer) getTargetDirectory(dateTaken *time.Time) (string, error) {
	if dateTaken == nil {
		// Use current date if no date could be determined
		now := time.Now()
		dateTaken = &now
	}

	year := dateTaken.Format("2006")
	month := dateTaken.Format("01")

	targetDir := filepath.Join(o.mediaPath, year, month)
	return targetDir, nil
}

// getFinalPath determines the final file path, handling naming conflicts
func (o *Organizer) getFinalPath(targetDir, fileName string) string {
	basePath := filepath.Join(targetDir, fileName)

	// If file doesn't exist, use the original name
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return basePath
	}

	// File exists, append a number
	ext := filepath.Ext(fileName)
	nameWithoutExt := fileName[:len(fileName)-len(ext)]

	for i := 1; i < 1000; i++ {
		newName := fmt.Sprintf("%s(%d)%s", nameWithoutExt, i, ext)
		newPath := filepath.Join(targetDir, newName)

		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}

	// Fallback - use timestamp
	timestamp := time.Now().Unix()
	newName := fmt.Sprintf("%s_%d%s", nameWithoutExt, timestamp, ext)
	return filepath.Join(targetDir, newName)
}

// moveFile moves a file from source to destination
func (o *Organizer) moveFile(src, dst string) error {
	// Try rename first (fastest if on same filesystem)
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	// If rename fails, copy and delete
	return o.copyAndDelete(src, dst)
}

// copyAndDelete copies a file and deletes the source
func (o *Organizer) copyAndDelete(src, dst string) error {
	srcFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		os.Remove(dst) // Clean up partial file
		return err
	}

	// Ensure data is written to disk
	if err := dstFile.Sync(); err != nil {
		os.Remove(dst)
		return err
	}

	// Remove source file
	return os.Remove(src)
}

// GetDirectoryStructure returns the organized directory structure
func (o *Organizer) GetDirectoryStructure() (map[string]interface{}, error) {
	structure := make(map[string]interface{})

	err := filepath.Walk(o.mediaPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Continue even if there's an error
		}

		// Skip temp directory
		if info.IsDir() && info.Name() == "temp" {
			return filepath.SkipDir
		}

		// Only include directories that match YYYY/MM pattern
		if info.IsDir() {
			relPath, err := filepath.Rel(o.mediaPath, path)
			if err != nil {
				return nil
			}

			// Skip root directory
			if relPath == "." {
				return nil
			}

			parts := filepath.SplitList(relPath)
			if len(parts) == 1 && len(parts[0]) == 4 { // Year directory
				if structure[parts[0]] == nil {
					structure[parts[0]] = make(map[string]int)
				}
			} else if len(parts) == 2 && len(parts[1]) == 2 { // Month directory
				year := parts[0]
				month := parts[1]

				if structure[year] == nil {
					structure[year] = make(map[string]int)
				}

				// Count files in this month
				fileCount := o.countFilesInDirectory(path)
				structure[year].(map[string]int)[month] = fileCount
			}
		}

		return nil
	})

	return structure, err
}

// countFilesInDirectory counts the number of files in a directory
func (o *Organizer) countFilesInDirectory(dirPath string) int {
	count := 0
	filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			count++
		}
		return nil
	})
	return count
}
