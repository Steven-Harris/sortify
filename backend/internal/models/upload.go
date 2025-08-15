package models

import (
	"time"
)

// UploadSession represents an active upload session
type UploadSession struct {
	ID           string            `json:"id"`
	FileName     string            `json:"filename"`
	FileSize     int64             `json:"fileSize"`
	ChunkSize    int64             `json:"chunkSize"`
	TotalChunks  int               `json:"totalChunks"`
	UploadedSize int64             `json:"uploadedSize"`
	Checksum     string            `json:"checksum"`
	TempPath     string            `json:"tempPath"`
	Metadata     map[string]string `json:"metadata"`
	CreatedAt    time.Time         `json:"createdAt"`
	UpdatedAt    time.Time         `json:"updatedAt"`
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
	SessionID   string `json:"sessionId"`
	ChunkNumber int    `json:"chunkNumber"`
	ChunkSize   int64  `json:"chunkSize"`
	Checksum    string `json:"checksum"`
}

// UploadProgress represents the current progress of an upload
type UploadProgress struct {
	SessionID       string  `json:"sessionId"`
	FileName        string  `json:"fileName"`
	UploadedBytes   int64   `json:"uploadedBytes"`
	TotalBytes      int64   `json:"totalBytes"`
	UploadedChunks  int     `json:"uploadedChunks"`
	TotalChunks     int     `json:"totalChunks"`
	PercentComplete float64 `json:"percentComplete"`
	Status          string  `json:"status"`
}

// StartUploadRequest represents the request to start an upload
type StartUploadRequest struct {
	FileName  string            `json:"fileName"`
	FileSize  int64             `json:"fileSize"`
	ChunkSize int64             `json:"chunkSize"`
	Checksum  string            `json:"checksum"`
	Algorithm string            `json:"algorithm"`
	Metadata  map[string]string `json:"metadata"`
}

// UploadChunkRequest represents the request to upload a chunk
type UploadChunkRequest struct {
	SessionID   string `json:"sessionId"`
	ChunkNumber int    `json:"chunkNumber"`
	ChunkSize   int64  `json:"chunkSize"`
	Checksum    string `json:"checksum"`
	Algorithm   string `json:"algorithm"`
}

// CompleteUploadRequest represents the request to complete an upload
type CompleteUploadRequest struct {
	SessionID string `json:"sessionId"`
	Checksum  string `json:"checksum"`
	Algorithm string `json:"algorithm"`
}
