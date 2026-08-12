package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Steven-Harris/sortify/backend/internal/media"
	"github.com/Steven-Harris/sortify/backend/internal/models"
	"github.com/stretchr/testify/assert"
)

func createTestSession(t *testing.T, handler *UploadHandlers) string {
	t.Helper()
	reqBody := `{"filename":"test.jpg","fileSize":1024,"chunkSize":256}`
	req := httptest.NewRequest("POST", "/api/upload/start", strings.NewReader(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.StartUploadHandler(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code, "Failed to create test session")

	var response map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err, "Failed to unmarshal session response")

	sessionID, ok := response["sessionId"].(string)
	if !ok {
		t.Fatal("Expected sessionId in response")
	}
	return sessionID
}

func TestStartUploadHandler(t *testing.T) {
	tempDir := t.TempDir()
	mediaDir := t.TempDir()
	handler := NewUploadHandlers(tempDir, mediaDir)

	t.Run("valid upload request", func(t *testing.T) {
		request := &models.StartUploadRequest{
			FileName:  "test.jpg",
			FileSize:  1024,
			ChunkSize: 256,
			Checksum:  "abc123",
			Metadata:  map[string]string{"type": "photo"},
		}

		body, err := json.Marshal(request)
		assert.NoError(t, err, "Failed to marshal request")

		req := httptest.NewRequest("POST", "/api/upload/start", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.StartUploadHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Expected successful response")

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err, "Failed to unmarshal response")

		assert.NotNil(t, response["uploadId"], "Expected uploadId in response")
		assert.NotNil(t, response["sessionId"], "Expected sessionId in response")
		assert.NotNil(t, response["totalChunks"], "Expected totalChunks in response")
	})

	t.Run("zero file size", func(t *testing.T) {
		request := &models.StartUploadRequest{
			FileName:  "test.jpg",
			FileSize:  0,
			ChunkSize: 256,
		}

		body, err := json.Marshal(request)
		assert.NoError(t, err, "Failed to marshal request")

		req := httptest.NewRequest("POST", "/api/upload/start", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.StartUploadHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected bad request for zero file size")
	})

	t.Run("zero chunk size defaults to 1MB", func(t *testing.T) {
		request := &models.StartUploadRequest{
			FileName:  "test.jpg",
			FileSize:  1024,
			ChunkSize: 0,
		}

		body, err := json.Marshal(request)
		assert.NoError(t, err, "Failed to marshal request")

		req := httptest.NewRequest("POST", "/api/upload/start", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.StartUploadHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Expected successful response with default chunk size")

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err, "Failed to unmarshal response")

		assert.NotNil(t, response["totalChunks"], "Expected totalChunks in response")
	})

	t.Run("empty filename", func(t *testing.T) {
		request := &models.StartUploadRequest{
			FileName:  "",
			FileSize:  1024,
			ChunkSize: 256,
		}

		body, err := json.Marshal(request)
		assert.NoError(t, err, "Failed to marshal request")

		req := httptest.NewRequest("POST", "/api/upload/start", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.StartUploadHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected bad request for empty filename")
	})

	t.Run("invalid JSON request", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/upload/start", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.StartUploadHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected bad request for invalid JSON")
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/upload/start", nil)
		rr := httptest.NewRecorder()

		handler.StartUploadHandler(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code, "Expected method not allowed")
	})
}

func TestUploadChunkHandler(t *testing.T) {
	tempDir := t.TempDir()
	mediaDir := t.TempDir()
	handler := NewUploadHandlers(tempDir, mediaDir)

	t.Run("valid chunk upload", func(t *testing.T) {
		sessionID := createTestSession(t, handler)

		// Create multipart form data
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		sessionField, err := writer.CreateFormField("sessionId")
		assert.NoError(t, err, "Failed to create sessionId field")
		sessionField.Write([]byte(sessionID))

		chunkField, err := writer.CreateFormField("chunkNumber")
		assert.NoError(t, err, "Failed to create chunkNumber field")
		chunkField.Write([]byte("0"))

		fileField, err := writer.CreateFormFile("chunk", "chunk")
		assert.NoError(t, err, "Failed to create chunk field")
		fileField.Write([]byte("test chunk data"))

		writer.Close()

		req := httptest.NewRequest("POST", "/api/upload/chunk", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()

		handler.UploadChunkHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Expected successful chunk upload")

		var response map[string]interface{}
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err, "Failed to unmarshal chunk response")

		assert.NotNil(t, response["percentComplete"], "Expected progress in response")
	})

	t.Run("missing session ID", func(t *testing.T) {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		chunkField, err := writer.CreateFormField("chunkNumber")
		assert.NoError(t, err, "Failed to create chunkNumber field")
		chunkField.Write([]byte("0"))

		writer.Close()

		req := httptest.NewRequest("POST", "/api/upload/chunk", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()

		handler.UploadChunkHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected bad request for missing session ID")
	})

	t.Run("invalid session ID", func(t *testing.T) {
		var buf bytes.Buffer
		writer := multipart.NewWriter(&buf)

		sessionField, err := writer.CreateFormField("sessionId")
		assert.NoError(t, err, "Failed to create sessionId field")
		sessionField.Write([]byte("invalid-session"))

		chunkField, err := writer.CreateFormField("chunkNumber")
		assert.NoError(t, err, "Failed to create chunkNumber field")
		chunkField.Write([]byte("0"))

		writer.Close()

		req := httptest.NewRequest("POST", "/api/upload/chunk", &buf)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		rr := httptest.NewRecorder()

		handler.UploadChunkHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected bad request for invalid session")
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/upload/chunk", nil)
		rr := httptest.NewRecorder()

		handler.UploadChunkHandler(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code, "Expected method not allowed")
	})
}

func TestGetProgressHandler(t *testing.T) {
	tempDir := t.TempDir()
	mediaDir := t.TempDir()
	handler := NewUploadHandlers(tempDir, mediaDir)

	t.Run("valid session ID", func(t *testing.T) {
		sessionID := createTestSession(t, handler)

		req := httptest.NewRequest("GET", "/api/upload/progress?sessionId="+sessionID, nil)
		rr := httptest.NewRecorder()

		handler.GetProgressHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Expected successful progress request")

		var response map[string]interface{}
		err := json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err, "Failed to unmarshal progress response")

		assert.NotNil(t, response["percentComplete"], "Expected percentComplete in response")
		assert.NotNil(t, response["uploadedChunks"], "Expected uploadedChunks in response")
		assert.NotNil(t, response["totalChunks"], "Expected totalChunks in response")
	})

	t.Run("missing session ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/upload/progress", nil)
		rr := httptest.NewRecorder()

		handler.GetProgressHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected bad request for missing session ID")
	})

	t.Run("invalid session ID", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/upload/progress?sessionId=invalid", nil)
		rr := httptest.NewRecorder()

		handler.GetProgressHandler(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code, "Expected not found for invalid session")
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/upload/progress", nil)
		rr := httptest.NewRecorder()

		handler.GetProgressHandler(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code, "Expected method not allowed")
	})
}

func TestCompleteUploadHandler(t *testing.T) {
	tempDir := t.TempDir()
	mediaDir := t.TempDir()
	handler := NewUploadHandlers(tempDir, mediaDir)

	t.Run("valid completion request", func(t *testing.T) {
		sessionID := createTestSession(t, handler)
		session, err := handler.manager.GetSession(sessionID)
		assert.NoError(t, err)
		chunk := bytes.Repeat([]byte("a"), int(session.FileSize))
		err = handler.manager.UploadChunk(sessionID, 0, chunk, "", "sha256")
		assert.NoError(t, err)

		request := &models.CompleteUploadRequest{
			SessionID: sessionID,
		}

		body, err := json.Marshal(request)
		assert.NoError(t, err, "Failed to marshal completion request")

		req := httptest.NewRequest("POST", "/api/upload/complete", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CompleteUploadHandler(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code, "Expected successful completion")

		var response models.UploadCompletionResult
		err = json.Unmarshal(rr.Body.Bytes(), &response)
		assert.NoError(t, err, "Expected JSON response")
		assert.Equal(t, sessionID, response.SessionID)
		assert.True(t, response.Organized)
		assert.False(t, response.Duplicate)
		assert.NotEmpty(t, response.RelativePath)
		assert.NotEmpty(t, response.AbsolutePath)
	})

	t.Run("missing session ID", func(t *testing.T) {
		request := &models.CompleteUploadRequest{
			SessionID: "",
			Checksum:  "test-checksum",
		}

		body, err := json.Marshal(request)
		assert.NoError(t, err, "Failed to marshal completion request")

		req := httptest.NewRequest("POST", "/api/upload/complete", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CompleteUploadHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected bad request for missing session ID")
	})

	t.Run("invalid JSON request", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/upload/complete", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()

		handler.CompleteUploadHandler(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code, "Expected bad request for invalid JSON")
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/upload/complete", nil)
		rr := httptest.NewRecorder()

		handler.CompleteUploadHandler(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code, "Expected method not allowed")
	})
}

func TestBuildUploadCompletionResult(t *testing.T) {
	mediaRoot := t.TempDir()
	dateTaken := time.Date(2024, 3, 15, 14, 30, 22, 0, time.UTC)
	session := &models.UploadSession{
		FileName: "IMG_20240315_143022.jpg",
		Metadata: map[string]string{"camera": "test"},
	}
	mediaInfo := &media.MediaInfo{
		FileName:   "IMG_20240315_143022(1).jpg",
		DateSource: media.DateSourceFileName,
		DateTaken:  &dateTaken,
	}

	organizedPath := filepath.Join(mediaRoot, "2024", "03", "IMG_20240315_143022(1).jpg")
	assert.NoError(t, os.MkdirAll(filepath.Dir(organizedPath), 0o755))
	assert.NoError(t, os.WriteFile(organizedPath, []byte("x"), 0o644))

	result := buildUploadCompletionResult(
		"session-1",
		filepath.Join(mediaRoot, "upload.tmp"),
		session,
		mediaInfo,
		mediaRoot,
	)

	assert.True(t, result.Organized)
	assert.True(t, result.ConflictRenamed)
	assert.Equal(t, "IMG_20240315_143022.jpg", result.ConflictRenamedFrom)
	assert.Equal(t, "2024/03/IMG_20240315_143022(1).jpg", result.RelativePath)
	assert.Equal(t, "filename", result.MetadataDateSource)
	assert.Equal(t, session.Metadata, result.Metadata)
}

func TestPauseUploadHandler(t *testing.T) {
	tempDir := t.TempDir()
	mediaDir := t.TempDir()
	handler := NewUploadHandlers(tempDir, mediaDir)

	t.Run("valid pause request", func(t *testing.T) {
		sessionID := createTestSession(t, handler)

		req := httptest.NewRequest("POST", "/api/upload/pause?sessionId="+sessionID, nil)
		rr := httptest.NewRecorder()

		handler.PauseUploadHandler(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code, "Expected successful pause")
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/upload/pause", nil)
		rr := httptest.NewRecorder()

		handler.PauseUploadHandler(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code, "Expected method not allowed")
	})
}

func TestResumeUploadHandler(t *testing.T) {
	tempDir := t.TempDir()
	mediaDir := t.TempDir()
	handler := NewUploadHandlers(tempDir, mediaDir)

	t.Run("valid resume request", func(t *testing.T) {
		sessionID := createTestSession(t, handler)

		// First pause the session
		pauseReq := httptest.NewRequest("POST", "/api/upload/pause?sessionId="+sessionID, nil)
		pauseRr := httptest.NewRecorder()
		handler.PauseUploadHandler(pauseRr, pauseReq)
		assert.Equal(t, http.StatusNoContent, pauseRr.Code, "Failed to pause session")

		// Now resume it
		req := httptest.NewRequest("POST", "/api/upload/resume?sessionId="+sessionID, nil)
		rr := httptest.NewRecorder()

		handler.ResumeUploadHandler(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code, "Expected successful resume")
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/upload/resume", nil)
		rr := httptest.NewRecorder()

		handler.ResumeUploadHandler(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code, "Expected method not allowed")
	})
}

func TestCancelUploadHandler(t *testing.T) {
	tempDir := t.TempDir()
	mediaDir := t.TempDir()
	handler := NewUploadHandlers(tempDir, mediaDir)

	t.Run("valid cancel request", func(t *testing.T) {
		sessionID := createTestSession(t, handler)

		req := httptest.NewRequest("DELETE", "/api/upload/cancel?sessionId="+sessionID, nil)
		rr := httptest.NewRecorder()

		handler.CancelUploadHandler(rr, req)

		assert.Equal(t, http.StatusNoContent, rr.Code, "Expected successful cancel")
	})

	t.Run("method not allowed", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/upload/cancel", nil)
		rr := httptest.NewRecorder()

		handler.CancelUploadHandler(rr, req)

		assert.Equal(t, http.StatusMethodNotAllowed, rr.Code, "Expected method not allowed")
	})
}
