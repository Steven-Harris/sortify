package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Steven-harris/sortify/backend/internal/media"
	"github.com/Steven-harris/sortify/backend/internal/models"
	"github.com/Steven-harris/sortify/backend/internal/upload"
	"github.com/Steven-harris/sortify/backend/pkg/response"
)

type UploadHandlers struct {
	manager   *upload.Manager
	organizer *media.Organizer
}

func NewUploadHandlers(tempDir, mediaPath string) *UploadHandlers {
	manager := upload.NewManager(tempDir, 10)
	organizer := media.NewOrganizer(mediaPath)
	return &UploadHandlers{
		manager:   manager,
		organizer: organizer,
	}
}

func (h *UploadHandlers) StartUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.StartUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode start upload request", "error", err)
		response.BadRequest(w, "Invalid request body")
		return
	}

	if req.FileName == "" {
		response.BadRequest(w, "Filename is required")
		return
	}
	if req.FileSize <= 0 {
		response.BadRequest(w, "File size must be greater than 0")
		return
	}
	if req.ChunkSize <= 0 {
		req.ChunkSize = 1024 * 1024
	}

	session, err := h.manager.CreateSession(&req)
	if err != nil {
		slog.Error("Failed to create upload session", "error", err)
		response.InternalError(w, "Failed to create upload session")
		return
	}

	slog.Info("Upload session created",
		"session_id", session.ID,
		"filename", session.FileName,
		"file_size", session.FileSize,
	)

	response.NoContent(w)
}

func (h *UploadHandlers) UploadChunkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		slog.Error("Failed to parse multipart form", "error", err)
		response.BadRequest(w, "Failed to parse form data")
		return
	}

	sessionID := r.FormValue("session_id")
	chunkNumberStr := r.FormValue("chunk_number")
	expectedChecksum := r.FormValue("checksum")

	if sessionID == "" {
		response.BadRequest(w, "Session ID is required")
		return
	}

	chunkNumber, err := strconv.Atoi(chunkNumberStr)
	if err != nil {
		response.BadRequest(w, "Invalid chunk number")
		return
	}

	file, _, err := r.FormFile("chunk")
	if err != nil {
		slog.Error("Failed to get chunk file", "error", err)
		response.BadRequest(w, "Chunk file is required")
		return
	}
	defer file.Close()

	chunkData, err := io.ReadAll(file)
	if err != nil {
		slog.Error("Failed to read chunk data", "error", err)
		response.InternalError(w, "Failed to read chunk data")
		return
	}

	if err := h.manager.UploadChunk(sessionID, chunkNumber, chunkData, expectedChecksum); err != nil {
		slog.Error("Failed to upload chunk",
			"error", err,
			"session_id", sessionID,
			"chunk_number", chunkNumber,
		)
		response.InternalError(w, fmt.Sprintf("Failed to upload chunk: %v", err))
		return
	}

	progress, err := h.manager.GetProgress(sessionID)
	if err != nil {
		slog.Error("Failed to get upload progress", "error", err)
		response.InternalError(w, "Failed to get progress")
		return
	}

	slog.Info("Chunk uploaded successfully",
		"session_id", sessionID,
		"chunk_number", chunkNumber,
		"chunk_size", len(chunkData),
		"progress", fmt.Sprintf("%.2f%%", progress.PercentComplete),
	)

	response.Success(w, progress)
}

func (h *UploadHandlers) CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.CompleteUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode complete upload request", "error", err)
		response.BadRequest(w, "Invalid request body")
		return
	}

	if req.SessionID == "" {
		response.BadRequest(w, "Session ID is required")
		return
	}

	if err := h.manager.CompleteUpload(req.SessionID, req.Checksum); err != nil {
		slog.Error("Failed to complete upload",
			"error", err,
			"session_id", req.SessionID,
		)
		response.InternalError(w, fmt.Sprintf("Failed to complete upload: %v", err))
		return
	}

	tempPath, err := h.manager.GetTempFilePath(req.SessionID)
	if err != nil {
		slog.Error("Failed to get temp file path",
			"error", err,
			"session_id", req.SessionID,
		)
		response.InternalError(w, "Failed to get temporary file path")
		return
	}

	session, err := h.manager.GetSession(req.SessionID)
	if err != nil {
		slog.Error("Failed to get session",
			"error", err,
			"session_id", req.SessionID,
		)
		response.InternalError(w, "Failed to get session information")
		return
	}

	mediaInfo, err := h.organizer.OrganizeFile(tempPath, session.FileName)
	if err != nil {
		slog.Error("Failed to organize file",
			"error", err,
			"session_id", req.SessionID,
			"filename", session.FileName,
		)
		response.InternalError(w, fmt.Sprintf("Failed to organize file: %v", err))
		return
	}

	if err := h.manager.CleanupSession(req.SessionID); err != nil {
		slog.Warn("Failed to cleanup session",
			"error", err,
			"session_id", req.SessionID,
		)
	}

	slog.Info("Upload completed and organized successfully",
		"session_id", req.SessionID,
		"filename", mediaInfo.FileName,
		"media_type", mediaInfo.MediaType,
		"date_taken", mediaInfo.DateTaken,
		"date_source", mediaInfo.DateSource,
	)

	result := map[string]any{
		"session_id": req.SessionID,
		"filename":   mediaInfo.FileName,
		"media_info": mediaInfo,
		"organized":  true,
	}

	response.Success(w, result)
}

func (h *UploadHandlers) GetProgressHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		response.BadRequest(w, "Session ID is required")
		return
	}

	progress, err := h.manager.GetProgress(sessionID)
	if err != nil {
		slog.Error("Failed to get upload progress",
			"error", err,
			"session_id", sessionID,
		)
		response.NotFound(w, "Session not found")
		return
	}

	response.Success(w, progress)
}

func (h *UploadHandlers) PauseUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		response.BadRequest(w, "Session ID is required")
		return
	}

	if err := h.manager.PauseUpload(sessionID); err != nil {
		slog.Error("Failed to pause upload",
			"error", err,
			"session_id", sessionID,
		)
		response.InternalError(w, fmt.Sprintf("Failed to pause upload: %v", err))
		return
	}

	slog.Info("Upload paused", "session_id", sessionID)
	response.NoContent(w)
}

func (h *UploadHandlers) ResumeUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		response.BadRequest(w, "Session ID is required")
		return
	}

	if err := h.manager.ResumeUpload(sessionID); err != nil {
		slog.Error("Failed to resume upload",
			"error", err,
			"session_id", sessionID,
		)
		response.InternalError(w, fmt.Sprintf("Failed to resume upload: %v", err))
		return
	}

	slog.Info("Upload resumed", "session_id", sessionID)
	response.NoContent(w)
}

func (h *UploadHandlers) CancelUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		response.Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		response.BadRequest(w, "Session ID is required")
		return
	}

	if err := h.manager.CancelUpload(sessionID); err != nil {
		slog.Error("Failed to cancel upload",
			"error", err,
			"session_id", sessionID,
		)
		response.InternalError(w, fmt.Sprintf("Failed to cancel upload: %v", err))
		return
	}

	slog.Info("Upload cancelled", "session_id", sessionID)
	response.NoContent(w)
}
