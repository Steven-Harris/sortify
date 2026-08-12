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

	"github.com/Steven-Harris/sortify/backend/internal/security"
	"github.com/rwcarlsen/goexif/exif"
)

type Extractor struct {
	filenamePatterns []*regexp.Regexp
}

func NewExtractor() *Extractor {
	return &Extractor{
		filenamePatterns: buildFilenamePatterns(),
	}
}

func (e *Extractor) ExtractMetadata(filePath string) (*MediaInfo, error) {
	return e.ExtractMetadataWithOriginalName(filePath, "")
}

func (e *Extractor) ExtractMetadataWithOriginalName(filePath, originalFileName string) (*MediaInfo, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info: %w", err)
	}

	sourceFileName := filepath.Base(filePath)
	detectedName := sourceFileName
	if originalFileName != "" {
		detectedName = filepath.Base(originalFileName)
	}

	info := &MediaInfo{
		FileName:         sourceFileName,
		OriginalFileName: originalFileName,
		FileSize:         fileInfo.Size(),
		ExtraMetadata:    make(map[string]string),
	}

	info.MimeType = mime.TypeByExtension(filepath.Ext(detectedName))
	info.MediaType = e.determineMediaType(detectedName, info.MimeType)

	e.extractDateFromEmbeddedMetadata(filePath, info)

	filenameForParsing := detectedName
	if info.DateTaken == nil {
		e.extractDateFromFilename(filenameForParsing, info)
	}
	if info.DateTaken == nil {
		e.extractDateFromFileTime(fileInfo, info)
	}

	slog.Info("Metadata extracted",
		"filename", info.FileName,
		"originalFilename", info.OriginalFileName,
		"media_type", info.MediaType,
		"date_source", info.DateSource,
		"date_taken", info.DateTaken,
	)

	return info, nil
}

func (e *Extractor) ExtractDateFromFilename(filename string, info *MediaInfo) {
	e.extractDateFromFilename(filename, info)
}

func (e *Extractor) determineMediaType(fileName, mimeType string) MediaType {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return MediaTypePhoto
	case strings.HasPrefix(mimeType, "video/"):
		return MediaTypeVideo
	}

	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".jpg", ".jpeg", ".png", ".gif", ".bmp", ".tiff", ".tif", ".webp", ".heic", ".heif":
		return MediaTypePhoto
	case ".mp4", ".mov", ".avi", ".mkv", ".webm", ".m4v", ".3gp", ".wmv", ".flv":
		return MediaTypeVideo
	default:
		return MediaTypeOther
	}
}

func (e *Extractor) extractDateFromEmbeddedMetadata(filePath string, info *MediaInfo) {
	if info.MediaType != MediaTypePhoto && info.MediaType != MediaTypeVideo {
		return
	}

	e.extractDateFromEXIF(filePath, info)
}

func (e *Extractor) extractDateFromEXIF(filePath string, info *MediaInfo) {
	cleanPath, err := security.ValidateFilePath(filePath)
	if err != nil {
		slog.Warn("Invalid file path for metadata extraction", "error", err, "path", filePath)
		return
	}

	file, err := os.Open(cleanPath) // #nosec G304 - path validated by security.ValidateFilePath
	if err != nil {
		slog.Debug("Failed to open file for metadata extraction", "error", err, "file", filePath)
		return
	}
	defer file.Close()

	x, err := exif.Decode(file)
	if err != nil {
		slog.Debug("Failed to decode EXIF data", "error", err, "file", filePath)
		return
	}

	if dt := extractEXIFDate(x); dt != nil {
		info.DateTaken = dt
		info.DateSource = DateSourceEmbeddedMetadata
		slog.Debug("Date extracted from embedded metadata", "date", dt, "file", filePath)
	}

	if info.Camera == nil {
		info.Camera = &CameraInfo{}
	}

	if make, err := x.Get(exif.Make); err == nil {
		if s, err := make.StringVal(); err == nil {
			info.Camera.Make = strings.TrimSpace(s)
		}
	}

	if model, err := x.Get(exif.Model); err == nil {
		if s, err := model.StringVal(); err == nil {
			info.Camera.Model = strings.TrimSpace(s)
		}
	}

	if lens, err := x.Get(exif.LensModel); err == nil {
		if s, err := lens.StringVal(); err == nil {
			info.Camera.LensModel = strings.TrimSpace(s)
		}
	}

	if lat, long, err := x.LatLong(); err == nil {
		info.Location = &LocationInfo{
			Latitude:  lat,
			Longitude: long,
		}
	}
}

func extractEXIFDate(x *exif.Exif) *time.Time {
	if dt, err := x.DateTime(); err == nil {
		return &dt
	}

	tags := []exif.FieldName{
		exif.DateTimeOriginal,
		exif.DateTimeDigitized,
		exif.DateTime,
	}

	for _, tag := range tags {
		field, err := x.Get(tag)
		if err != nil {
			continue
		}
		value, err := field.StringVal()
		if err != nil {
			continue
		}
		if parsed := parseEXIFTime(strings.TrimSpace(value)); parsed != nil {
			return parsed
		}
	}

	return nil
}

func parseEXIFTime(value string) *time.Time {
	layouts := []string{
		"2006:01:02 15:04:05",
		time.RFC3339,
		"2006-01-02 15:04:05",
	}

	for _, layout := range layouts {
		if parsed, err := time.Parse(layout, value); err == nil {
			return &parsed
		}
	}

	return nil
}

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
	if date.Year() != year || int(date.Month()) != month || date.Day() != day {
		return nil
	}

	return &date
}

func (e *Extractor) extractDateFromFileTime(fileInfo os.FileInfo, info *MediaInfo) {
	modTime := fileInfo.ModTime()
	info.DateTaken = &modTime
	info.DateSource = DateSourceFileTime
	slog.Debug("Using file modification time", "date", modTime)
}

func (e *Extractor) NeedsUserInput(info *MediaInfo) bool {
	return info.DateSource == DateSourceFileTime || info.DateSource == DateSourceUnknown
}

func buildFilenamePatterns() []*regexp.Regexp {
	patterns := []string{
		`IMG_(\d{4})(\d{2})(\d{2})_(\d{2})(\d{2})(\d{2})`,
		`VID_(\d{4})(\d{2})(\d{2})_(\d{2})(\d{2})(\d{2})`,
		`PXL_(\d{4})(\d{2})(\d{2})_(\d{2})(\d{2})(\d{2})`,
		`MVIMG_(\d{4})(\d{2})(\d{2})_(\d{2})(\d{2})(\d{2})`,
		`Screenshot_(\d{4})-(\d{2})-(\d{2})-(\d{2})-(\d{2})-(\d{2})`,
		`WhatsApp.+(\d{4})-(\d{2})-(\d{2}).+(\d{2})\.(\d{2})\.(\d{2})`,
		`(\d{4})(\d{2})(\d{2})_(\d{2})(\d{2})(\d{2})`,
		`(\d{4})-(\d{2})-(\d{2})[_ -](\d{2})[-:.](\d{2})[-:.](\d{2})`,
		`(\d{4})-(\d{2})-(\d{2})`,
		`(\d{4})(\d{2})(\d{2})`,
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
