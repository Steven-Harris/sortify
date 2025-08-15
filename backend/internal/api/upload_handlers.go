package api

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Steven-Harris/sortify/backend/internal/media"
	"github.com/Steven-Harris/sortify/backend/internal/models"
	"github.com/Steven-Harris/sortify/backend/internal/upload"
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
		Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.StartUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode start upload request", "error", err)
		BadRequest(w, "Invalid request body")
		return
	}

	if req.FileName == "" {
		BadRequest(w, "Filename is required")
		return
	}
	if req.FileSize <= 0 {
		BadRequest(w, "File size must be greater than 0")
		return
	}
	if req.ChunkSize <= 0 {
		req.ChunkSize = 1024 * 1024
	}

	// Pass algorithm to manager if needed
	if req.Algorithm == "" {
		req.Algorithm = "sha256"
	}

	session, err := h.manager.CreateSession(&req)
	if err != nil {
		slog.Error("Failed to create upload session", "error", err)
		InternalError(w, "Failed to create upload session")
		return
	}

	slog.Info("Upload session created",
		"sessionId", session.ID,
		"filename", session.FileName,
		"fileSize", session.FileSize,
		"algorithm", req.Algorithm,
	)

	totalChunks := int((req.FileSize + req.ChunkSize - 1) / req.ChunkSize)

	result := map[string]any{
		"uploadId":    session.ID,
		"sessionId":   session.ID, // For backward compatibility
		"totalChunks": totalChunks,
	}

	Success(w, result)
}

func (h *UploadHandlers) UploadChunkHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if err := r.ParseMultipartForm(0); err != nil { // 0 = no limit
		slog.Error("Failed to parse multipart form", "error", err)
		BadRequest(w, "Failed to parse form data")
		return
	}

	sessionID := r.FormValue("sessionId")
	if sessionID == "" {
		sessionID = r.FormValue("sessionId")
	}

	chunkNumberStr := r.FormValue("chunkNumber")
	if chunkNumberStr == "" {
		chunkNumberStr = r.FormValue("chunk_number")
	}

	expectedChecksum := r.FormValue("checksum")
	algorithm := r.FormValue("algorithm")
	if algorithm == "" {
		algorithm = "sha256"
	}

	if sessionID == "" {
		BadRequest(w, "Session ID is required")
		return
	}

	chunkNumber, err := strconv.Atoi(chunkNumberStr)
	if err != nil {
		BadRequest(w, "Invalid chunk number")
		return
	}

	file, _, err := r.FormFile("chunk")
	if err != nil {
		slog.Error("Failed to get chunk file", "error", err)
		BadRequest(w, "Chunk file is required")
		return
	}
	defer file.Close()

	chunkData, err := io.ReadAll(file)
	if err != nil {
		slog.Error("Failed to read chunk data", "error", err)
		BadRequest(w, "Failed to read chunk data")
		return
	}

	if err := h.manager.UploadChunk(sessionID, chunkNumber, chunkData, expectedChecksum, algorithm); err != nil {
		slog.Error("Failed to upload chunk",
			"error", err,
			"sessionId", sessionID,
			"chunk_number", chunkNumber,
		)
		InternalError(w, fmt.Sprintf("Failed to upload chunk: %v", err))
		return
	}

	progress, err := h.manager.GetProgress(sessionID)
	if err != nil {
		slog.Error("Failed to get upload progress", "error", err)
		InternalError(w, "Failed to get progress")
		return
	}

	slog.Info("Chunk uploaded successfully",
		"sessionId", sessionID,
		"chunk_number", chunkNumber,
		"chunk_size", len(chunkData),
		"progress", fmt.Sprintf("%.2f%%", progress.PercentComplete),
	)

	Success(w, progress)
}

func (h *UploadHandlers) CompleteUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req models.CompleteUploadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.Error("Failed to decode complete upload request", "error", err)
		BadRequest(w, "Invalid request body")
		return
	}

	if req.SessionID == "" {
		BadRequest(w, "Session ID is required")
		return
	}

	if req.Algorithm == "" {
		req.Algorithm = "sha256"
	}

	if err := h.manager.CompleteUpload(req.SessionID, req.Checksum, req.Algorithm); err != nil {
		slog.Error("Failed to complete upload",
			"error", err,
			"sessionId", req.SessionID,
		)
		InternalError(w, fmt.Sprintf("Failed to complete upload: %v", err))
		return
	}

	tempPath, err := h.manager.GetTempFilePath(req.SessionID)
	if err != nil {
		slog.Error("Failed to get temp file path",
			"error", err,
			"sessionId", req.SessionID,
		)
		InternalError(w, "Failed to get temporary file path")
		return
	}

	session, err := h.manager.GetSession(req.SessionID)
	if err != nil {
		slog.Error("Failed to get session",
			"error", err,
			"sessionId", req.SessionID,
		)
		InternalError(w, "Failed to get session information")
		return
	}

	mediaInfo, err := h.organizer.OrganizeFile(tempPath, session.FileName)
	if err != nil {
		slog.Error("Failed to organize file",
			"error", err,
			"sessionId", req.SessionID,
			"filename", session.FileName,
		)
		InternalError(w, fmt.Sprintf("Failed to organize file: %v", err))
		return
	}

	if err := h.manager.CleanupSession(req.SessionID); err != nil {
		slog.Warn("Failed to cleanup session",
			"error", err,
			"sessionId", req.SessionID,
		)
	}

	slog.Info("Upload completed and organized successfully",
		"sessionId", req.SessionID,
		"filename", mediaInfo.FileName,
		"media_type", mediaInfo.MediaType,
		"date_taken", mediaInfo.DateTaken,
		"date_source", mediaInfo.DateSource,
	)

	result := map[string]any{
		"sessionId": req.SessionID,
		"filename":  mediaInfo.FileName,
		"mediaInfo": mediaInfo,
		"organized": true,
	}

	Success(w, result)
}

func (h *UploadHandlers) GetProgressHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		sessionID = r.URL.Query().Get("sessionId")
	}
	if sessionID == "" {
		BadRequest(w, "Session ID is required")
		return
	}

	progress, err := h.manager.GetProgress(sessionID)
	if err != nil {
		slog.Error("Failed to get upload progress",
			"error", err,
			"sessionId", sessionID,
		)
		NotFound(w, "Session not found")
		return
	}

	Success(w, progress)
}

func (h *UploadHandlers) PauseUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		sessionID = r.URL.Query().Get("sessionId")
	}
	if sessionID == "" {
		BadRequest(w, "Session ID is required")
		return
	}

	if err := h.manager.PauseUpload(sessionID); err != nil {
		slog.Error("Failed to pause upload",
			"error", err,
			"sessionId", sessionID,
		)
		InternalError(w, fmt.Sprintf("Failed to pause upload: %v", err))
		return
	}

	slog.Info("Upload paused", "sessionId", sessionID)
	NoContent(w)
}

func (h *UploadHandlers) ResumeUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		sessionID = r.URL.Query().Get("sessionId")
	}
	if sessionID == "" {
		BadRequest(w, "Session ID is required")
		return
	}

	if err := h.manager.ResumeUpload(sessionID); err != nil {
		slog.Error("Failed to resume upload",
			"error", err,
			"sessionId", sessionID,
		)
		InternalError(w, fmt.Sprintf("Failed to resume upload: %v", err))
		return
	}

	slog.Info("Upload resumed", "sessionId", sessionID)
	NoContent(w)
}

func (h *UploadHandlers) CancelUploadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		Error(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	sessionID := r.URL.Query().Get("sessionId")
	if sessionID == "" {
		sessionID = r.URL.Query().Get("sessionId")
	}
	if sessionID == "" {
		BadRequest(w, "Session ID is required")
		return
	}

	if err := h.manager.CancelUpload(sessionID); err != nil {
		slog.Error("Failed to cancel upload",
			"error", err,
			"sessionId", sessionID,
		)
		InternalError(w, fmt.Sprintf("Failed to cancel upload: %v", err))
		return
	}

	slog.Info("Upload cancelled", "sessionId", sessionID)
	NoContent(w)
}
