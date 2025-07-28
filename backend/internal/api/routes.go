package api

import "net/http"

func (s *Server) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	// Apply middleware
	var handler http.Handler = mux
	handler = CORS(s.config.CORSOrigins)(handler)
	handler = Logging(handler)
	handler = Recovery(handler)

	// Register routes
	mux.HandleFunc("/", s.RootHandler)
	mux.HandleFunc("/api/health", s.HealthHandler)

	// Upload routes
	mux.HandleFunc("/api/upload/start", s.uploadHandler.StartUploadHandler)
	mux.HandleFunc("/api/upload/chunk", s.uploadHandler.UploadChunkHandler)
	mux.HandleFunc("/api/upload/complete", s.uploadHandler.CompleteUploadHandler)
	mux.HandleFunc("/api/upload/progress", s.uploadHandler.GetProgressHandler)
	mux.HandleFunc("/api/upload/pause", s.uploadHandler.PauseUploadHandler)
	mux.HandleFunc("/api/upload/resume", s.uploadHandler.ResumeUploadHandler)
	mux.HandleFunc("/api/upload/cancel", s.uploadHandler.CancelUploadHandler)

	// Media browsing routes
	mux.HandleFunc("/api/media/browse", s.mediaHandler.BrowseHandler)
	mux.HandleFunc("/api/media/files", s.mediaHandler.ListFilesHandler)
	mux.HandleFunc("/api/media/metadata", s.mediaHandler.MetadataHandler)
	mux.HandleFunc("/api/media/user-date", s.mediaHandler.UserDateHandler)

	// Static file serving for media files
	mediaFileServer := http.FileServer(http.Dir(s.config.MediaPath))
	mux.Handle("/media/", http.StripPrefix("/media/", mediaFileServer))

	// Catch-all for undefined routes
	mux.HandleFunc("/api/", s.NotFoundHandler)

	return handler
}
