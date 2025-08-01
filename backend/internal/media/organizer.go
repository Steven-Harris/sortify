package media

import (
	"crypto/sha256"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
)

type Organizer struct {
	mediaPath string
	extractor *Extractor
}

func NewOrganizer(mediaPath string) *Organizer {
	return &Organizer{
		mediaPath: mediaPath,
		extractor: NewExtractor(),
	}
}

func (o *Organizer) OrganizeFile(tempFilePath, originalFileName string) (*MediaInfo, error) {
	info, err := o.extractor.ExtractMetadata(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	info.FileName = originalFileName

	tempFileName := filepath.Base(tempFilePath)
	if info.DateSource == "filename" && tempFileName != originalFileName {
		info.DateTaken = nil
		info.DateSource = ""

		o.extractor.ExtractDateFromFilename(originalFileName, info)

		if info.DateTaken == nil {
			if fileInfo, err := os.Stat(tempFilePath); err == nil {
				if fileInfo.ModTime().Year() > 1970 { // Reasonable date check
					info.DateTaken = &[]time.Time{fileInfo.ModTime()}[0]
					info.DateSource = "file_time"
				}
			}
		}
	}

	if duplicate, err := o.checkDuplicate(tempFilePath, info); err != nil {
		slog.Error("Failed to check for duplicates", "error", err, "file", originalFileName)
	} else if duplicate {
		slog.Info("Duplicate file detected, skipping", "file", originalFileName)
		os.Remove(tempFilePath) // Clean up temp file
		return info, nil
	}

	targetDir, err := o.getTargetDirectory(info.DateTaken)
	if err != nil {
		return nil, fmt.Errorf("failed to determine target directory: %w", err)
	}

	if err := os.MkdirAll(targetDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	sanitizedFilename := o.sanitizeFileName(originalFileName)
	finalPath := filepath.Join(targetDir, sanitizedFilename)
	finalPath = o.handleDuplicates(finalPath)

	if err := o.moveFile(tempFilePath, finalPath); err != nil {
		return nil, fmt.Errorf("failed to move file: %w", err)
	}

	slog.Info("File organized successfully",
		"originalFile", originalFileName,
		"finalPath", finalPath,
		"dateTaken", info.DateTaken,
		"dateSource", info.DateSource,
	)

	return info, nil
}

func (o *Organizer) handleDuplicates(targetPath string) string {
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return targetPath
	}

	dir := filepath.Dir(targetPath)
	filename := filepath.Base(targetPath)
	ext := filepath.Ext(filename)
	nameWithoutExt := strings.TrimSuffix(filename, ext)

	counter := 1
	for {
		newFilename := fmt.Sprintf("%s(%d)%s", nameWithoutExt, counter, ext)
		newPath := filepath.Join(dir, newFilename)

		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
		counter++
	}
}

func (o *Organizer) checkDuplicate(filePath string, info *MediaInfo) (bool, error) {
	hash, err := o.calculateFileHash(filePath)
	if err != nil {
		return false, err
	}

	targetDir, err := o.getTargetDirectory(info.DateTaken)
	if err != nil {
		return false, err
	}

	if _, err := os.Stat(targetDir); os.IsNotExist(err) {
		return false, nil
	}

	var foundDuplicate bool
	err = filepath.Walk(targetDir, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if fileInfo.IsDir() {
			return nil
		}

		existingHash, err := o.calculateFileHash(path)
		if err != nil {
			return nil
		}

		if existingHash == hash {
			foundDuplicate = true
			slog.Info("Duplicate found", "original", filePath, "existing", path)
			return filepath.SkipAll
		}

		return nil
	})

	return foundDuplicate, err
}

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

func (o *Organizer) getTargetDirectory(dateTaken *time.Time) (string, error) {
	// Validate and sanitize the date
	validatedDate := o.validateDate(dateTaken)

	year := validatedDate.Format("2006")
	month := validatedDate.Format("January") // Use full English month name

	targetDir := filepath.Join(o.mediaPath, year, month)
	return targetDir, nil
}

func (o *Organizer) getFinalPath(targetDir, fileName string) string {
	// Sanitize the filename first
	sanitizedFileName := o.sanitizeFileName(fileName)
	basePath := filepath.Join(targetDir, sanitizedFileName)

	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return basePath
	}

	ext := filepath.Ext(sanitizedFileName)
	nameWithoutExt := sanitizedFileName[:len(sanitizedFileName)-len(ext)]

	for i := 1; i < 1000; i++ {
		newName := fmt.Sprintf("%s(%d)%s", nameWithoutExt, i, ext)
		newPath := filepath.Join(targetDir, newName)

		if _, err := os.Stat(newPath); os.IsNotExist(err) {
			return newPath
		}
	}

	timestamp := time.Now().Unix()
	newName := fmt.Sprintf("%s_%d%s", nameWithoutExt, timestamp, ext)
	return filepath.Join(targetDir, newName)
}

// sanitizeFileName removes or replaces problematic characters in filenames
func (o *Organizer) sanitizeFileName(fileName string) string {
	if fileName == "" {
		return "untitled"
	}

	// Remove or replace problematic characters
	// Replace common problematic characters with safe alternatives
	replacements := map[string]string{
		"/":  "_", // Forward slash
		"\\": "_", // Backslash
		":":  "_", // Colon
		"*":  "_", // Asterisk
		"?":  "_", // Question mark
		"\"": "_", // Double quote
		"<":  "_", // Less than
		">":  "_", // Greater than
		"|":  "_", // Pipe
	}

	result := fileName
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	// Remove control characters and other problematic Unicode characters
	var sanitized strings.Builder
	for _, r := range result {
		if unicode.IsControl(r) || r == 0 {
			continue // Skip control characters
		}
		sanitized.WriteRune(r)
	}

	result = sanitized.String()

	// Trim whitespace and dots from beginning and end
	result = strings.Trim(result, " .")

	// Handle empty result after sanitization
	if result == "" {
		return "untitled"
	}

	// Ensure filename isn't too long (most filesystems support 255 characters)
	if len(result) > 200 { // Leave some room for numbering if needed
		ext := filepath.Ext(result)
		nameWithoutExt := result[:len(result)-len(ext)]
		if len(nameWithoutExt) > 200-len(ext) {
			nameWithoutExt = nameWithoutExt[:200-len(ext)]
		}
		result = nameWithoutExt + ext
	}

	return result
}

// validateDate ensures the date is reasonable and handles edge cases
func (o *Organizer) validateDate(dateTaken *time.Time) *time.Time {
	if dateTaken == nil {
		now := time.Now()
		return &now
	}

	// Check for unreasonable dates (before digital photography era or too far in future)
	minDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	maxDate := time.Now().AddDate(1, 0, 0) // One year in the future

	if dateTaken.Before(minDate) || dateTaken.After(maxDate) {
		slog.Warn("Date outside reasonable range, using current time",
			"original_date", dateTaken,
			"min_date", minDate,
			"max_date", maxDate,
		)
		now := time.Now()
		return &now
	}

	return dateTaken
}

func (o *Organizer) moveFile(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}

	return o.copyAndDelete(src, dst)
}

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
		os.Remove(dst)
		return err
	}

	if err := dstFile.Sync(); err != nil {
		os.Remove(dst)
		return err
	}

	return os.Remove(src)
}

func (o *Organizer) GetDirectoryStructure() (map[string]any, error) {
	structure := make(map[string]any)

	err := filepath.Walk(o.mediaPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if info.IsDir() && info.Name() == "temp" {
			return filepath.SkipDir
		}

		if info.IsDir() {
			relPath, err := filepath.Rel(o.mediaPath, path)
			if err != nil {
				return nil
			}

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

				fileCount := o.countFilesInDirectory(path)
				structure[year].(map[string]int)[month] = fileCount
			}
		}

		return nil
	})

	return structure, err
}

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

func (o *Organizer) ScanFiles(year, month string, limit, offset int) ([]MediaFileInfo, error) {
	var files []MediaFileInfo
	var targetPath string

	if year == "" {
		targetPath = o.mediaPath
	} else if month == "" {
		targetPath = filepath.Join(o.mediaPath, year)
	} else {
		targetPath = filepath.Join(o.mediaPath, year, month)
	}

	slog.Debug("ScanFiles called", "year", year, "month", month, "limit", limit, "offset", offset)
	slog.Debug("Media path configuration", "mediaPath", o.mediaPath, "targetPath", targetPath)

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		slog.Debug("Target directory does not exist", "targetPath", targetPath)
		return files, nil
	}

	slog.Debug("Starting filepath.Walk", "targetPath", targetPath)

	err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			slog.Warn("Error walking file", "path", path, "error", err)
			return nil // Continue walking even if there's an error with one file
		}

		slog.Debug("Walking path", "path", path, "isDir", info.IsDir(), "name", info.Name())

		if info.IsDir() {
			slog.Debug("Skipping directory", "path", path)
			return nil
		}
		if strings.Contains(path, "/temp/") || strings.Contains(path, "\\temp\\") {
			slog.Debug("Skipping temp file", "path", path)
			return nil
		}

		slog.Debug("Processing file", "path", path, "name", info.Name())

		if !o.isMediaFile(path) {
			slog.Debug("Skipping non-media file", "path", path, "ext", filepath.Ext(path))
			return nil
		}

		slog.Debug("Found media file", "path", path, "ext", filepath.Ext(path))

		relPath, err := filepath.Rel(o.mediaPath, path)
		if err != nil {
			relPath = path
		}

		mediaInfo, err := o.extractor.ExtractMetadata(path)
		if err != nil {
			slog.Warn("Failed to extract metadata", "file", path, "error", err)
			mediaInfo = &MediaInfo{
				FileName: info.Name(),
				FileSize: info.Size(),
			}
		}

		fileInfo := MediaFileInfo{
			ID:           o.generateFileID(relPath),
			FileName:     info.Name(),
			RelativePath: relPath,
			Size:         info.Size(),
			ModTime:      info.ModTime(),
			MediaType:    o.getMediaType(path),
			URL:          fmt.Sprintf("/media/%s", relPath),
		}

		if mediaInfo != nil {
			if mediaInfo.DateTaken != nil {
				fileInfo.DateTaken = mediaInfo.DateTaken
			}
			if mediaInfo.Camera != nil {
				camera := mediaInfo.Camera.Make
				if mediaInfo.Camera.Model != "" {
					if camera != "" {
						camera += " " + mediaInfo.Camera.Model
					} else {
						camera = mediaInfo.Camera.Model
					}
				}
				fileInfo.Camera = camera
			}
			if mediaInfo.Location != nil {
				fileInfo.Location = fmt.Sprintf("%f,%f", mediaInfo.Location.Latitude, mediaInfo.Location.Longitude)
			}
			fileInfo.Width = mediaInfo.Width
			fileInfo.Height = mediaInfo.Height
			fileInfo.Duration = mediaInfo.Duration
		}

		files = append(files, fileInfo)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to scan files: %w", err)
	}

	o.sortFiles(files)

	start := offset
	end := offset + limit

	if start >= len(files) {
		return []MediaFileInfo{}, nil
	}

	if end > len(files) {
		end = len(files)
	}

	return files[start:end], nil
}

func (o *Organizer) isMediaFile(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	supportedExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".tiff": true,
		".mp4": true, ".mov": true, ".avi": true, ".mkv": true, ".webm": true, ".m4v": true,
		".3gp": true, ".wmv": true, ".flv": true,
	}
	return supportedExts[ext]
}

func (o *Organizer) getMediaType(filePath string) string {
	ext := strings.ToLower(filepath.Ext(filePath))
	imageExts := map[string]bool{
		".jpg": true, ".jpeg": true, ".png": true, ".gif": true, ".bmp": true, ".tiff": true,
	}
	if imageExts[ext] {
		return "image"
	}
	return "video"
}

func (o *Organizer) generateFileID(relPath string) string {
	hash := sha256.Sum256([]byte(relPath))
	return fmt.Sprintf("%x", hash[:8])
}

func (o *Organizer) sortFiles(files []MediaFileInfo) {
	for i := 0; i < len(files)-1; i++ {
		for j := i + 1; j < len(files); j++ {
			var timeI, timeJ time.Time

			if files[i].DateTaken != nil {
				timeI = *files[i].DateTaken
			} else {
				timeI = files[i].ModTime
			}

			if files[j].DateTaken != nil {
				timeJ = *files[j].DateTaken
			} else {
				timeJ = files[j].ModTime
			}

			if timeI.Before(timeJ) {
				files[i], files[j] = files[j], files[i]
			}
		}
	}
}
