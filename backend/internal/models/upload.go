package models

import (
	"time"
)

// UploadSession represents an active upload session
type UploadSession struct {
	ID           string            `json:"id"`
	FileName     string            `json:"filename"`
	FileSize     int64             `json:"file_size"`
	ChunkSize    int64             `json:"chunk_size"`
	TotalChunks  int               `json:"total_chunks"`
	UploadedSize int64             `json:"uploaded_size"`
	Checksum     string            `json:"checksum"`      // Expected SHA256 checksum
	TempPath     string            `json:"temp_path"`     // Temporary file path
	Metadata     map[string]string `json:"metadata"`      // Additional metadata
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Status       UploadStatus      `json:"status"`
}

// UploadStatus represents the status of an upload
type UploadStatus string

const (
	StatusInitialized UploadStatus = "initialized"
	StatusUploading   UploadStatus = "uploading"
	StatusPaused      UploadStatus = "paused"
	StatusCompleted   UploadStatus = "completed"
	StatusFailed      UploadStatus = "failed"
	StatusCancelled   UploadStatus = "cancelled"
)

// ChunkInfo represents information about an uploaded chunk
type ChunkInfo struct {
	SessionID   string `json:"session_id"`
	ChunkNumber int    `json:"chunk_number"`
	ChunkSize   int64  `json:"chunk_size"`
	Checksum    string `json:"checksum"` // SHA256 of this specific chunk
}

// UploadProgress represents the current progress of an upload
type UploadProgress struct {
	SessionID       string  `json:"session_id"`
	FileName        string  `json:"filename"`
	UploadedBytes   int64   `json:"uploaded_bytes"`
	TotalBytes      int64   `json:"total_bytes"`
	UploadedChunks  int     `json:"uploaded_chunks"`
	TotalChunks     int     `json:"total_chunks"`
	PercentComplete float64 `json:"percent_complete"`
	Status          string  `json:"status"`
}

// StartUploadRequest represents the request to start an upload
type StartUploadRequest struct {
	FileName  string            `json:"filename"`
	FileSize  int64             `json:"file_size"`
	ChunkSize int64             `json:"chunk_size"`
	Checksum  string            `json:"checksum"`
	Metadata  map[string]string `json:"metadata"`
}

// UploadChunkRequest represents the request to upload a chunk
type UploadChunkRequest struct {
	SessionID   string `json:"session_id"`
	ChunkNumber int    `json:"chunk_number"`
	ChunkSize   int64  `json:"chunk_size"`
	Checksum    string `json:"checksum"`
}

// CompleteUploadRequest represents the request to complete an upload
type CompleteUploadRequest struct {
	SessionID string `json:"session_id"`
	Checksum  string `json:"checksum"`
}
