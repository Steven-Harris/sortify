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

	"github.com/Steven-Harris/sortify/backend/internal/security"
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

func (o *Organizer) MediaPath() string {
	return o.mediaPath
}

func (o *Organizer) OrganizeFile(tempFilePath, originalFileName string) (*MediaInfo, error) {
	fileExt := strings.ToLower(filepath.Ext(originalFileName))
	fileSize := int64(0)
	if stat, err := os.Stat(tempFilePath); err == nil {
		fileSize = stat.Size()
	}

	slog.Info("Processing file for organization",
		"tempPath", tempFilePath,
		"originalFileName", originalFileName,
		"fileType", fileExt,
		"fileSize", fileSize)

	info, err := o.extractor.ExtractMetadataWithOriginalName(tempFilePath, originalFileName)
	if err != nil {
		return nil, fmt.Errorf("failed to extract metadata: %w", err)
	}

	info.FileName = originalFileName

	hash, err := o.calculateFileHash(tempFilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to hash uploaded file: %w", err)
	}

	if existingPath, duplicate, err := o.findExistingByHash(hash); err != nil {
		slog.Error("Failed to check for duplicates", "error", err, "file", originalFileName)
	} else if duplicate {
		info.IsDuplicate = true
		info.CanonicalPath = existingPath
		slog.Info("Duplicate file detected, skipping", "file", originalFileName, "canonicalPath", existingPath)
		if err := os.Remove(tempFilePath); err != nil {
			slog.Warn("Failed to remove duplicate temp file", "error", err, "path", tempFilePath)
		}
		return info, nil
	}

	targetDir, err := o.getTargetDirectory(info.DateTaken)
	if err != nil {
		return nil, fmt.Errorf("failed to determine target directory: %w", err)
	}

	if err := os.MkdirAll(targetDir, 0750); err != nil {
		return nil, fmt.Errorf("failed to create target directory: %w", err)
	}

	finalPath, err := o.resolveConflictPath(targetDir, originalFileName, hash)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve target path: %w", err)
	}

	if err := o.moveFile(tempFilePath, finalPath); err != nil {
		return nil, fmt.Errorf("failed to move file: %w", err)
	}

	info.FileName = filepath.Base(finalPath)
	info.CanonicalPath = finalPath

	slog.Info("File organized successfully",
		"originalFile", originalFileName,
		"finalPath", finalPath,
		"dateTaken", info.DateTaken,
		"dateSource", info.DateSource,
	)

	return info, nil
}

func (o *Organizer) findExistingByHash(hash string) (string, bool, error) {
	var existingPath string
	err := filepath.Walk(o.mediaPath, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if fileInfo.IsDir() {
			return nil
		}
		relToMedia, err := filepath.Rel(o.mediaPath, path)
		if err != nil || relToMedia == "." || strings.HasPrefix(relToMedia, "..") {
			return nil
		}
		parts := strings.Split(relToMedia, string(filepath.Separator))
		if len(parts) < 3 {
			return nil
		}
		if strings.Contains(path, string(filepath.Separator)+"temp"+string(filepath.Separator)) {
			return nil
		}

		existingHash, err := o.calculateFileHash(path)
		if err != nil {
			return nil
		}

		if existingHash == hash {
			existingPath = path
			return filepath.SkipAll
		}

		return nil
	})
	if err != nil {
		return "", false, err
	}
	return existingPath, existingPath != "", nil
}

func (o *Organizer) resolveConflictPath(targetDir, originalFileName, newHash string) (string, error) {
	sanitizedFilename := o.sanitizeFileName(originalFileName)
	ext := filepath.Ext(sanitizedFilename)
	nameWithoutExt := strings.TrimSuffix(sanitizedFilename, ext)

	basePath := filepath.Join(targetDir, sanitizedFilename)
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return basePath, nil
	}

	for counter := 1; ; counter++ {
		candidateName := fmt.Sprintf("%s (%d)%s", nameWithoutExt, counter, ext)
		candidatePath := filepath.Join(targetDir, candidateName)
		if _, err := os.Stat(candidatePath); os.IsNotExist(err) {
			return candidatePath, nil
		}

		existingHash, err := o.calculateFileHash(candidatePath)
		if err != nil {
			return "", err
		}
		if existingHash == newHash {
			return candidatePath, nil
		}
	}
}

func (o *Organizer) calculateFileHash(filePath string) (string, error) {
	cleanPath, err := security.ValidateFilePath(filePath)
	if err != nil {
		return "", fmt.Errorf("path validation failed: %w", err)
	}

	file, err := os.Open(cleanPath) // #nosec G304 - path validated by security.ValidateFilePath
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
	validatedDate := o.validateDate(dateTaken)
	year := validatedDate.Format("2006")
	month := validatedDate.Format("January")
	return filepath.Join(o.mediaPath, year, month), nil
}

func (o *Organizer) getFinalPath(targetDir, fileName string) string {
	sanitizedFileName := o.sanitizeFileName(fileName)
	basePath := filepath.Join(targetDir, sanitizedFileName)
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		return basePath
	}

	ext := filepath.Ext(sanitizedFileName)
	nameWithoutExt := strings.TrimSuffix(sanitizedFileName, ext)
	for counter := 1; ; counter++ {
		candidatePath := filepath.Join(targetDir, fmt.Sprintf("%s (%d)%s", nameWithoutExt, counter, ext))
		if _, err := os.Stat(candidatePath); os.IsNotExist(err) {
			return candidatePath
		}
	}
}

func (o *Organizer) sanitizeFileName(fileName string) string {
	if fileName == "" {
		return "untitled"
	}

	replacements := map[string]string{
		"/":  "_",
		"\\": "_",
		":":  "_",
		"*":  "_",
		"?":  "_",
		"\"": "_",
		"<":  "_",
		">":  "_",
		"|":  "_",
	}

	result := fileName
	for old, new := range replacements {
		result = strings.ReplaceAll(result, old, new)
	}

	var sanitized strings.Builder
	for _, r := range result {
		if unicode.IsControl(r) || r == 0 {
			continue
		}
		sanitized.WriteRune(r)
	}

	result = strings.Trim(sanitized.String(), " .")
	if result == "" {
		return "untitled"
	}

	if len(result) > 200 {
		ext := filepath.Ext(result)
		nameWithoutExt := result[:len(result)-len(ext)]
		if len(nameWithoutExt) > 200-len(ext) {
			nameWithoutExt = nameWithoutExt[:200-len(ext)]
		}
		result = nameWithoutExt + ext
	}

	return result
}

func (o *Organizer) validateDate(dateTaken *time.Time) *time.Time {
	if dateTaken == nil {
		now := time.Now()
		return &now
	}

	minDate := time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC)
	maxDate := time.Now().AddDate(1, 0, 0)
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
	cleanSrc, cleanDst, err := security.ValidateFilePathPair(src, dst)
	if err != nil {
		return fmt.Errorf("path validation failed: %w", err)
	}

	if err := os.Rename(cleanSrc, cleanDst); err == nil {
		return nil
	}

	return o.copyAndDelete(cleanSrc, cleanDst)
}

func (o *Organizer) copyAndDelete(src, dst string) error {
	cleanSrc, cleanDst, err := security.ValidateFilePathPair(src, dst)
	if err != nil {
		return fmt.Errorf("path validation failed: %w", err)
	}

	srcFile, err := os.Open(cleanSrc) // #nosec G304 - path validated by security.ValidateFilePathPair
	if err != nil {
		return err
	}
	defer srcFile.Close()

	dstFile, err := os.Create(cleanDst) // #nosec G304 - path validated by security.ValidateFilePathPair
	if err != nil {
		return err
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		if removeErr := os.Remove(cleanDst); removeErr != nil {
			slog.Warn("Failed to remove incomplete destination file", "error", removeErr, "path", cleanDst)
		}
		return err
	}

	if err := dstFile.Sync(); err != nil {
		if removeErr := os.Remove(cleanDst); removeErr != nil {
			slog.Warn("Failed to remove incomplete destination file", "error", removeErr, "path", cleanDst)
		}
		return err
	}

	return os.Remove(cleanSrc)
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
			if err != nil || relPath == "." {
				return nil
			}

			parts := strings.Split(relPath, string(filepath.Separator))
			if len(parts) == 1 && len(parts[0]) == 4 {
				if structure[parts[0]] == nil {
					structure[parts[0]] = make(map[string]int)
				}
			} else if len(parts) == 2 {
				year := parts[0]
				month := parts[1]
				if structure[year] == nil {
					structure[year] = make(map[string]int)
				}
				structure[year].(map[string]int)[month] = o.countFilesInDirectory(path)
			}
		}
		return nil
	})

	return structure, err
}

func (o *Organizer) countFilesInDirectory(dirPath string) int {
	count := 0
	if err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if !info.IsDir() {
			count++
		}
		return nil
	}); err != nil {
		slog.Warn("Failed to walk directory for file count", "error", err, "path", dirPath)
	}
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

	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		return files, nil
	}

	err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.Contains(path, string(filepath.Separator)+"temp"+string(filepath.Separator)) {
			return nil
		}
		if !o.isMediaFile(path) {
			return nil
		}

		relPath, err := filepath.Rel(o.mediaPath, path)
		if err != nil {
			relPath = path
		}

		mediaInfo, err := o.extractor.ExtractMetadata(path)
		if err != nil {
			slog.Warn("Failed to extract metadata", "file", path, "error", err)
			mediaInfo = &MediaInfo{FileName: info.Name(), FileSize: info.Size()}
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
