package api

import (
	"bytes"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Steven-harris/sortify/backend/internal/models"
)

func TestStartUploadHandler(t *testing.T) {
	tempDir := t.TempDir()
	mediaDir := t.TempDir()
	handler := NewUploadHandlers(tempDir, mediaDir)

	tests := []struct {
		name           string
		request        *models.StartUploadRequest
		expectedStatus int
	}{
		{
			name: "Valid upload request",
			request: &models.StartUploadRequest{
				FileName:  "test.jpg",
				FileSize:  1024,
				ChunkSize: 256,
				Checksum:  "abc123",
				Metadata:  map[string]string{"type": "photo"},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Zero file size",
			request: &models.StartUploadRequest{
				FileName:  "test.jpg",
				FileSize:  0,
				ChunkSize: 256,
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Zero chunk size",
			request: &models.StartUploadRequest{
				FileName:  "test.jpg",
				FileSize:  1024,
				ChunkSize: 0,
			},
			expectedStatus: http.StatusOK, // Should default to 1MB chunks
		},
		{
			name: "Empty filename",
			request: &models.StartUploadRequest{
				FileName:  "",
				FileSize:  1024,
				ChunkSize: 256,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, _ := json.Marshal(test.request)
			req := httptest.NewRequest("POST", "/api/upload/start", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			handler.StartUploadHandler(rr, req)

			if rr.Code != test.expectedStatus {
				t.Errorf("Expected status %d, got %d", test.expectedStatus, rr.Code)
			}

			if test.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response["success"] != true {
					t.Error("Expected success to be true")
				}

				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Error("Expected data field in response")
				}

				if data["id"] == nil {
					t.Error("Expected session id in data")
				}

				if data["total_chunks"] == nil {
					t.Error("Expected total_chunks in data")
				}
			}
		})
	}
}

func TestUploadChunkHandler(t *testing.T) {
	tempDir := t.TempDir()
	mediaDir := t.TempDir()
	handler := NewUploadHandlers(tempDir, mediaDir)

	// Create a session first
	startReq := &models.StartUploadRequest{
		FileName:  "test.jpg",
		FileSize:  1024,
		ChunkSize: 256,
	}

	body, _ := json.Marshal(startReq)
	req := httptest.NewRequest("POST", "/api/upload/start", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.StartUploadHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Failed to create session: %d", rr.Code)
	}

	var startResponse map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &startResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal start response: %v", err)
	}

	data, ok := startResponse["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected data field in start response")
	}

	sessionID := data["id"].(string)

	tests := []struct {
		name           string
		sessionID      string
		chunkNumber    string
		chunkData      []byte
		expectedStatus int
	}{
		{
			name:           "Valid chunk upload",
			sessionID:      sessionID,
			chunkNumber:    "0",
			chunkData:      []byte("test chunk data"),
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid session ID",
			sessionID:      "invalid",
			chunkNumber:    "0",
			chunkData:      []byte("test chunk data"),
			expectedStatus: http.StatusInternalServerError, // Manager will error
		},
		{
			name:           "Invalid chunk number",
			sessionID:      sessionID,
			chunkNumber:    "invalid",
			chunkData:      []byte("test chunk data"),
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty chunk data",
			sessionID:      sessionID,
			chunkNumber:    "1",
			chunkData:      []byte{},
			expectedStatus: http.StatusOK, // Empty chunks are allowed
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// Create multipart form data
			var body bytes.Buffer
			writer := multipart.NewWriter(&body)

			// Add form fields
			writer.WriteField("sessionId", test.sessionID)
			writer.WriteField("chunk_number", test.chunkNumber)

			// Add chunk file
			part, err := writer.CreateFormFile("chunk", "chunk.dat")
			if err != nil {
				t.Fatalf("Failed to create form file: %v", err)
			}
			part.Write(test.chunkData)
			writer.Close()

			req := httptest.NewRequest("POST", "/api/upload/chunk", &body)
			req.Header.Set("Content-Type", writer.FormDataContentType())

			rr := httptest.NewRecorder()
			handler.UploadChunkHandler(rr, req)

			if rr.Code != test.expectedStatus {
				t.Errorf("Expected status %d, got %d", test.expectedStatus, rr.Code)
			}

			if test.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response["success"] != true {
					t.Error("Expected success to be true")
				}
			}
		})
	}
}

func TestGetProgressHandler(t *testing.T) {
	tempDir := t.TempDir()
	mediaDir := t.TempDir()
	handler := NewUploadHandlers(tempDir, mediaDir)

	// Create a session
	startReq := &models.StartUploadRequest{
		FileName:  "test.jpg",
		FileSize:  1024,
		ChunkSize: 256,
	}

	body, _ := json.Marshal(startReq)
	req := httptest.NewRequest("POST", "/api/upload/start", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	handler.StartUploadHandler(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("Failed to create session: %d", rr.Code)
	}

	var startResponse map[string]interface{}
	err := json.Unmarshal(rr.Body.Bytes(), &startResponse)
	if err != nil {
		t.Fatalf("Failed to unmarshal start response: %v", err)
	}

	data, ok := startResponse["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("Expected data field in start response")
	}

	sessionID := data["id"].(string)

	tests := []struct {
		name           string
		sessionID      string
		expectedStatus int
	}{
		{
			name:           "Valid session ID",
			sessionID:      sessionID,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Invalid session ID",
			sessionID:      "invalid",
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			url := "/api/upload/progress?sessionId=" + test.sessionID
			req := httptest.NewRequest("GET", url, nil)

			rr := httptest.NewRecorder()
			handler.GetProgressHandler(rr, req)

			if rr.Code != test.expectedStatus {
				t.Errorf("Expected status %d, got %d", test.expectedStatus, rr.Code)
			}

			if test.expectedStatus == http.StatusOK {
				var response map[string]interface{}
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				if err != nil {
					t.Fatalf("Failed to unmarshal response: %v", err)
				}

				if response["success"] != true {
					t.Error("Expected success to be true")
				}

				data, ok := response["data"].(map[string]interface{})
				if !ok {
					t.Error("Expected data field in response")
				}

				if data["total_chunks"] == nil {
					t.Error("Expected total_chunks in data")
				}
			}
		})
	}
}

func TestInvalidJSONRequest(t *testing.T) {
	tempDir := t.TempDir()
	mediaDir := t.TempDir()
	handler := NewUploadHandlers(tempDir, mediaDir)

	req := httptest.NewRequest("POST", "/api/upload/start", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.StartUploadHandler(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
}
