package media

import (
	"time"
)

type MediaInfo struct {
	FileName      string            `json:"filename"`
	FileSize      int64             `json:"fileSize"`
	MimeType      string            `json:"mimeType"`
	MediaType     MediaType         `json:"mediaType"`
	DateTaken     *time.Time        `json:"dateTaken,omitempty"`
	DateSource    DateSource        `json:"dateSource"`
	Width         int               `json:"width,omitempty"`
	Height        int               `json:"height,omitempty"`
	Duration      *time.Duration    `json:"duration,omitempty"`
	Camera        *CameraInfo       `json:"camera,omitempty"`
	Location      *LocationInfo     `json:"location,omitempty"`
	ExtraMetadata map[string]string `json:"extraMetadata,omitempty"`
}

type MediaType string

const (
	MediaTypePhoto MediaType = "photo"
	MediaTypeVideo MediaType = "video"
	MediaTypeOther MediaType = "other"
)

type DateSource string

const (
	DateSourceEXIF      DateSource = "exif"
	DateSourceFileName  DateSource = "filename"
	DateSourceFileTime  DateSource = "fileTime"
	DateSourceUserInput DateSource = "userInput"
	DateSourceUnknown   DateSource = "unknown"
)

type CameraInfo struct {
	Make         string `json:"make,omitempty"`
	Model        string `json:"model,omitempty"`
	Software     string `json:"software,omitempty"`
	LensModel    string `json:"lensModel,omitempty"`
	FocalLength  string `json:"focalLength,omitempty"`
	Aperture     string `json:"aperture,omitempty"`
	ShutterSpeed string `json:"shutterSpeed,omitempty"`
	ISO          string `json:"iso,omitempty"`
	Flash        string `json:"flash,omitempty"`
}

type LocationInfo struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude,omitempty"`
}

type DateExtractionRequest struct {
	FileName     string `json:"filename"`
	OriginalPath string `json:"originalPath"`
	SessionID    string `json:"sessionId"`
}

type DateExtractionResponse struct {
	SessionID string    `json:"sessionId"`
	DateTaken time.Time `json:"dateTaken"`
}

type MediaFileInfo struct {
	ID           string         `json:"id"`
	FileName     string         `json:"filename"`
	RelativePath string         `json:"relativePath"`
	Size         int64          `json:"size"`
	ModTime      time.Time      `json:"modTime"`
	MediaType    string         `json:"type"`
	URL          string         `json:"url"`
	DateTaken    *time.Time     `json:"dateTaken,omitempty"`
	Camera       string         `json:"camera,omitempty"`
	Location     string         `json:"location,omitempty"`
	Width        int            `json:"width,omitempty"`
	Height       int            `json:"height,omitempty"`
	Duration     *time.Duration `json:"duration,omitempty"`
}
