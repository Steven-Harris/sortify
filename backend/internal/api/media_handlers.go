package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/Steven-harris/sortify/backend/internal/media"
	"github.com/Steven-harris/sortify/backend/pkg/response"
)

type MediaHandlers struct {
	organizer *media.Organizer
}

func NewMediaHandlers(mediaPath string) *MediaHandlers {
	return &MediaHandlers{
		organizer: media.NewOrganizer(mediaPath),
	}
}

func (h *MediaHandlers) BrowseHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	year := r.URL.Query().Get("year")
	month := r.URL.Query().Get("month")
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	limitInt := 50
	offsetInt := 0

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
		structure, err := h.organizer.GetDirectoryStructure()
		if err != nil {
			slog.Error("Failed to get directory structure", "error", err)
			response.InternalError(w, "Failed to retrieve media structure")
			return
		}

		response.Success(w, map[string]any{
			"type":      "structure",
			"structure": structure,
		})
		return
	}

	files, err := h.getFilesInDirectory(year, month, limitInt, offsetInt)
	if err != nil {
		slog.Error("Failed to get files", "error", err, "year", year, "month", month)
		response.InternalError(w, "Failed to retrieve files")
		return
	}

	response.Success(w, map[string]any{
		"type":   "files",
		"year":   year,
		"month":  month,
		"files":  files,
		"limit":  limitInt,
		"offset": offsetInt,
	})
}

func (h *MediaHandlers) MetadataHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		FilePath string `json:"filePath"`
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

	extractor := media.NewExtractor()
	info, err := extractor.ExtractMetadata(req.FilePath)
	if err != nil {
		slog.Error("Failed to extract metadata", "error", err, "filePath", req.FilePath)
		response.InternalError(w, "Failed to extract metadata")
		return
	}

	response.Success(w, info)
}

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
	slog.Info("User provided date for upload",
		"sessionId", req.SessionID,
		"dateTaken", req.DateTaken,
	)

	response.NoContent(w)
}

func (h *MediaHandlers) ListFilesHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	query := r.URL.Query().Get("q")
	mediaType := r.URL.Query().Get("type")
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")

	limitInt := 50
	offsetInt := 0

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

	// Get all files without pagination first
	allFiles, err := h.organizer.ScanFiles("", "", 10000, 0)
	if err != nil {
		slog.Error("Failed to scan files", "error", err)
		response.InternalError(w, "Failed to retrieve files")
		return
	}

	var filteredFiles []media.MediaFileInfo
	for _, file := range allFiles {
		if query != "" {
			queryMatch := false
			queryLower := strings.ToLower(query)
			if strings.Contains(strings.ToLower(file.FileName), queryLower) ||
				strings.Contains(strings.ToLower(file.Camera), queryLower) ||
				strings.Contains(strings.ToLower(file.Location), queryLower) {
				queryMatch = true
			}
			if !queryMatch {
				continue
			}
		}

		if mediaType != "" && mediaType != "all" && file.MediaType != mediaType {
			continue
		}

		filteredFiles = append(filteredFiles, file)
	}

	start := offsetInt
	end := offsetInt + limitInt

	if start >= len(filteredFiles) {
		filteredFiles = []media.MediaFileInfo{}
	} else {
		if end > len(filteredFiles) {
			end = len(filteredFiles)
		}
		filteredFiles = filteredFiles[start:end]
	}

	response.Success(w, map[string]any{
		"files":  filteredFiles,
		"total":  len(allFiles),
		"limit":  limitInt,
		"offset": offsetInt,
	})
}

func (h *MediaHandlers) getFilesInDirectory(year, month string, limit, offset int) ([]media.MediaFileInfo, error) {
	return h.organizer.ScanFiles(year, month, limit, offset)
}
