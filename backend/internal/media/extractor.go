package media

import (
	"fmt"
	"log/slog"
	"mime"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/rwcarlsen/goexif/exif"
)

// Extractor handles metadata extraction from media files
type Extractor struct {
	filenamePatterns []*regexp.Regexp
}

// NewExtractor creates a new metadata extractor
func NewExtractor() *Extractor {
	return &Extractor{
		filenamePatterns: buildFilenamePatterns(),
	}
}

// ExtractMetadata extracts metadata from a file
func (e *Extractor) ExtractMetadata(filePath string) (*MediaInfo, error) {
	info := &MediaInfo{
		FileName:      filepath.Base(filePath),
		ExtraMetadata: make(map[string]string),
	}

	// Get file info
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}
	info.FileSize = fileInfo.Size()

	// Determine MIME type
	info.MimeType = mime.TypeByExtension(filepath.Ext(filePath))
	info.MediaType = e.determineMediaType(info.MimeType)

	// Try extracting date from different sources
	e.extractDateFromEXIF(filePath, info)
	if info.DateTaken == nil {
		e.extractDateFromFilename(info.FileName, info)
	}
	if info.DateTaken == nil {
		e.extractDateFromFileTime(fileInfo, info)
	}

	// Extract additional metadata based on media type
	switch info.MediaType {
	case MediaTypePhoto:
		e.extractPhotoMetadata(filePath, info)
	case MediaTypeVideo:
		e.extractVideoMetadata(filePath, info)
	}

	slog.Info("Metadata extracted",
		"filename", info.FileName,
		"media_type", info.MediaType,
		"date_source", info.DateSource,
		"date_taken", info.DateTaken,
	)

	return info, nil
}

// ExtractDateFromFilename is a public wrapper for extracting date from filename
func (e *Extractor) ExtractDateFromFilename(filename string, info *MediaInfo) {
	e.extractDateFromFilename(filename, info)
}

// determineMediaType determines the media type from MIME type
func (e *Extractor) determineMediaType(mimeType string) MediaType {
	if strings.HasPrefix(mimeType, "image/") {
		return MediaTypePhoto
	}
	if strings.HasPrefix(mimeType, "video/") {
		return MediaTypeVideo
	}
	return MediaTypeOther
}

// extractDateFromEXIF extracts date from EXIF data
func (e *Extractor) extractDateFromEXIF(filePath string, info *MediaInfo) {
	if info.MediaType != MediaTypePhoto {
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		slog.Debug("Failed to open file for EXIF", "error", err, "file", filePath)
		return
	}
	defer file.Close()

	x, err := exif.Decode(file)
	if err != nil {
		slog.Debug("Failed to decode EXIF data", "error", err, "file", filePath)
		return
	}

	// Try to extract date taken
	if dt, err := x.DateTime(); err == nil {
		info.DateTaken = &dt
		info.DateSource = DateSourceEXIF
		slog.Debug("Date extracted from EXIF", "date", dt, "file", filePath)
	}

	// Extract camera information
	if info.Camera == nil {
		info.Camera = &CameraInfo{}
	}

	// Camera make
	if make, err := x.Get(exif.Make); err == nil {
		if s, err := make.StringVal(); err == nil {
			info.Camera.Make = strings.TrimSpace(s)
		}
	}

	// Camera model
	if model, err := x.Get(exif.Model); err == nil {
		if s, err := model.StringVal(); err == nil {
			info.Camera.Model = strings.TrimSpace(s)
		}
	}

	// Lens model
	if lens, err := x.Get(exif.LensModel); err == nil {
		if s, err := lens.StringVal(); err == nil {
			info.Camera.LensModel = strings.TrimSpace(s)
		}
	}

	// ISO - skip for now due to API differences
	// TODO: Implement ISO extraction with correct goexif API

	// Focal length - skip for now due to API differences
	// TODO: Implement focal length extraction with correct goexif API

	// Aperture - skip for now due to API differences
	// TODO: Implement aperture extraction with correct goexif API

	// Extract GPS data
	if lat, long, err := x.LatLong(); err == nil {
		info.Location = &LocationInfo{
			Latitude:  lat,
			Longitude: long,
		}
	}
}

// extractDateFromFilename extracts date from filename using patterns
func (e *Extractor) extractDateFromFilename(filename string, info *MediaInfo) {
	for _, pattern := range e.filenamePatterns {
		matches := pattern.FindStringSubmatch(filename)
		if len(matches) > 0 {
			if date := e.parseFilenameMatches(matches); date != nil {
				info.DateTaken = date
				info.DateSource = DateSourceFileName
				slog.Debug("Date extracted from filename", "filename", filename, "date", date)
				return
			}
		}
	}
}

// parseFilenameMatches parses date from regex matches
func (e *Extractor) parseFilenameMatches(matches []string) *time.Time {
	if len(matches) < 4 {
		return nil
	}

	year, err1 := strconv.Atoi(matches[1])
	month, err2 := strconv.Atoi(matches[2])
	day, err3 := strconv.Atoi(matches[3])

	if err1 != nil || err2 != nil || err3 != nil {
		return nil
	}

	// Optional time components
	hour, minute, second := 0, 0, 0
	if len(matches) >= 7 {
		if h, err := strconv.Atoi(matches[4]); err == nil {
			hour = h
		}
		if m, err := strconv.Atoi(matches[5]); err == nil {
			minute = m
		}
		if s, err := strconv.Atoi(matches[6]); err == nil {
			second = s
		}
	}

	date := time.Date(year, time.Month(month), day, hour, minute, second, 0, time.UTC)
	return &date
}

// extractDateFromFileTime uses file modification time as fallback
func (e *Extractor) extractDateFromFileTime(fileInfo os.FileInfo, info *MediaInfo) {
	modTime := fileInfo.ModTime()
	info.DateTaken = &modTime
	info.DateSource = DateSourceFileTime
	slog.Debug("Using file modification time", "date", modTime)
}

// extractPhotoMetadata extracts additional photo-specific metadata
func (e *Extractor) extractPhotoMetadata(filePath string, info *MediaInfo) {
	// The basic EXIF data is already extracted in extractDateFromEXIF
	// This function can be extended for additional photo-specific metadata
	slog.Debug("Photo metadata extraction completed", "file", filePath)
}

// NeedsUserInput determines if user input is required for date extraction
func (e *Extractor) NeedsUserInput(info *MediaInfo) bool {
	return info.DateSource == DateSourceFileTime || info.DateSource == DateSourceUnknown
}

// extractVideoMetadata extracts video-specific metadata (placeholder for future implementation)
func (e *Extractor) extractVideoMetadata(filePath string, info *MediaInfo) {
	// TODO: Implement video metadata extraction using FFprobe or similar
	// For now, we'll just mark it as video type
	slog.Debug("Video metadata extraction not yet implemented", "file", filePath)
}

// buildFilenamePatterns creates regex patterns for extracting dates from filenames
func buildFilenamePatterns() []*regexp.Regexp {
	patterns := []string{
		// IMG_20231225_143022.jpg
		`IMG_(\d{4})(\d{2})(\d{2})_(\d{2})(\d{2})(\d{2})`,
		// 20231225_143022.jpg
		`(\d{4})(\d{2})(\d{2})_(\d{2})(\d{2})(\d{2})`,
		// 2023-12-25_14-30-22.jpg
		`(\d{4})-(\d{2})-(\d{2})_(\d{2})-(\d{2})-(\d{2})`,
		// 2023-12-25.jpg
		`(\d{4})-(\d{2})-(\d{2})`,
		// 20231225.jpg
		`(\d{4})(\d{2})(\d{2})`,
		// VID_20231225_143022.mp4
		`VID_(\d{4})(\d{2})(\d{2})_(\d{2})(\d{2})(\d{2})`,
		// Screenshot_2023-12-25-14-30-22.png
		`Screenshot_(\d{4})-(\d{2})-(\d{2})-(\d{2})-(\d{2})-(\d{2})`,
		// WhatsApp Image 2023-12-25 at 14.30.22.jpeg
		`WhatsApp.+(\d{4})-(\d{2})-(\d{2}).+(\d{2})\.(\d{2})\.(\d{2})`,
	}

	var compiledPatterns []*regexp.Regexp
	for _, pattern := range patterns {
		if compiled, err := regexp.Compile(pattern); err == nil {
			compiledPatterns = append(compiledPatterns, compiled)
		} else {
			slog.Error("Failed to compile filename pattern", "pattern", pattern, "error", err)
		}
	}

	return compiledPatterns
}
