package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Steven-harris/sortify/backend/internal/media"
	"github.com/Steven-harris/sortify/backend/pkg/response"
)

// MediaHandlers contains all media-related HTTP handlers
type MediaHandlers struct {
	organizer *media.Organizer
}

// NewMediaHandlers creates a new media handlers instance
func NewMediaHandlers(mediaPath string) *MediaHandlers {
	return &MediaHandlers{
		organizer: media.NewOrganizer(mediaPath),
	}
}

// BrowseHandler handles browsing the organized media structure
func (h *MediaHandlers) BrowseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Get query parameters
	year := r.URL.Query().Get("year")
	month := r.URL.Query().Get("month")
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	// Convert limit and offset to integers
	limitInt := 50 // Default limit
	offsetInt := 0 // Default offset

	if limit != "" {
		if l, err := strconv.Atoi(limit); err == nil && l > 0 {
			limitInt = l
		}
	}

	if offset != "" {
		if o, err := strconv.Atoi(offset); err == nil && o >= 0 {
			offsetInt = o
		}
	}

	if year == "" {
		// Return directory structure overview
		structure, err := h.organizer.GetDirectoryStructure()
		if err != nil {
			slog.Error("Failed to get directory structure", "error", err)
			response.InternalError(w, "Failed to retrieve media structure")
			return
		}

		response.Success(w, map[string]interface{}{
			"type":      "structure",
			"structure": structure,
		})
		return
	}

	// Return files for specific year/month
	files, err := h.getFilesInDirectory(year, month, limitInt, offsetInt)
	if err != nil {
		slog.Error("Failed to get files", "error", err, "year", year, "month", month)
		response.InternalError(w, "Failed to retrieve files")
		return
	}

	response.Success(w, map[string]interface{}{
		"type":   "files",
		"year":   year,
		"month":  month,
		"files":  files,
		"limit":  limitInt,
		"offset": offsetInt,
	})
}

// MetadataHandler handles extracting metadata from uploaded files
func (h *MediaHandlers) MetadataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		FilePath string `json:"file_path"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode metadata request", "error", err)
		response.BadRequest(w, "Invalid request body")
		return
	}

	if req.FilePath == "" {
		response.BadRequest(w, "File path is required")
		return
	}

	// Extract metadata
	extractor := media.NewExtractor()
	info, err := extractor.ExtractMetadata(req.FilePath)
	if err != nil {
		slog.Error("Failed to extract metadata", "error", err, "file_path", req.FilePath)
		response.InternalError(w, "Failed to extract metadata")
		return
	}

	response.Success(w, info)
}

// UserDateHandler handles user input for date extraction when metadata fails
func (h *MediaHandlers) UserDateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req media.DateExtractionResponse
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode user date request", "error", err)
		response.BadRequest(w, "Invalid request body")
		return
	}

	if req.SessionID == "" {
		response.BadRequest(w, "Session ID is required")
		return
	}

	// TODO: Store the user-provided date and complete the upload process
	// For now, just acknowledge the request
	slog.Info("User provided date for upload",
		"session_id", req.SessionID,
		"date_taken", req.DateTaken,
	)

	response.SuccessWithMessage(w, nil, "Date recorded successfully")
}

// getFilesInDirectory retrieves files from a specific year/month directory
func (h *MediaHandlers) getFilesInDirectory(year, month string, limit, offset int) ([]map[string]interface{}, error) {
	// This is a placeholder implementation
	// In a real application, you'd scan the filesystem or query a database

	files := []map[string]interface{}{
		{
			"name":        "example1.jpg",
			"size":        1024768,
			"date_taken":  "2023-12-25T14:30:22Z",
			"media_type":  "photo",
			"camera_make": "Canon",
		},
		{
			"name":       "example2.mp4",
			"size":       10485760,
			"date_taken": "2023-12-25T15:45:10Z",
			"media_type": "video",
			"duration":   "00:02:30",
		},
	}

	// Apply pagination
	start := offset
	end := offset + limit

	if start >= len(files) {
		return []map[string]interface{}{}, nil
	}

	if end > len(files) {
		end = len(files)
	}

	return files[start:end], nil
}
