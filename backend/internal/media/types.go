package media

import (
	"time"
)

// MediaInfo represents extracted metadata from a media file
type MediaInfo struct {
	FileName      string            `json:"filename"`
	FileSize      int64             `json:"file_size"`
	MimeType      string            `json:"mime_type"`
	MediaType     MediaType         `json:"media_type"`
	DateTaken     *time.Time        `json:"date_taken,omitempty"`
	DateSource    DateSource        `json:"date_source"`
	Width         int               `json:"width,omitempty"`
	Height        int               `json:"height,omitempty"`
	Duration      *time.Duration    `json:"duration,omitempty"`
	Camera        *CameraInfo       `json:"camera,omitempty"`
	Location      *LocationInfo     `json:"location,omitempty"`
	ExtraMetadata map[string]string `json:"extra_metadata,omitempty"`
}

// MediaType represents the type of media file
type MediaType string

const (
	MediaTypePhoto MediaType = "photo"
	MediaTypeVideo MediaType = "video"
	MediaTypeOther MediaType = "other"
)

// DateSource indicates how the date was determined
type DateSource string

const (
	DateSourceEXIF       DateSource = "exif"
	DateSourceFileName   DateSource = "filename"
	DateSourceFileTime   DateSource = "file_time"
	DateSourceUserInput  DateSource = "user_input"
	DateSourceUnknown    DateSource = "unknown"
)

// CameraInfo contains camera-specific metadata
type CameraInfo struct {
	Make         string `json:"make,omitempty"`
	Model        string `json:"model,omitempty"`
	Software     string `json:"software,omitempty"`
	LensModel    string `json:"lens_model,omitempty"`
	FocalLength  string `json:"focal_length,omitempty"`
	Aperture     string `json:"aperture,omitempty"`
	ShutterSpeed string `json:"shutter_speed,omitempty"`
	ISO          string `json:"iso,omitempty"`
	Flash        string `json:"flash,omitempty"`
}

// LocationInfo contains GPS location data
type LocationInfo struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Altitude  float64 `json:"altitude,omitempty"`
}

// DateExtractionRequest represents a request for date extraction when metadata fails
type DateExtractionRequest struct {
	FileName     string `json:"filename"`
	OriginalPath string `json:"original_path"`
	SessionID    string `json:"session_id"`
}

// DateExtractionResponse represents the user's response for date extraction
type DateExtractionResponse struct {
	SessionID string    `json:"session_id"`
	DateTaken time.Time `json:"date_taken"`
}
