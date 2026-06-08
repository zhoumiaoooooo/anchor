package router

import (
	"net/http"

	"anchor-server/internal/handler"
	"anchor-server/internal/model"
)

func New(h *handler.Handler) http.Handler {
	mux := http.NewServeMux()

	// CORS middleware wrapper
	handler := corsMiddleware(mux)

	// Subjects
	mux.HandleFunc("POST /api/v1/subjects", h.CreateSubject)
	mux.HandleFunc("GET /api/v1/subjects", h.ListSubjects)
	mux.HandleFunc("GET /api/v1/subjects/{id}", func(w http.ResponseWriter, r *http.Request) {
		h.GetSubject(w, r, r.PathValue("id"))
	})

	// Interview
	mux.HandleFunc("GET /api/v1/subjects/{id}/interviews/progress", func(w http.ResponseWriter, r *http.Request) {
		h.GetInterviewProgress(w, r, r.PathValue("id"))
	})
	mux.HandleFunc("POST /api/v1/subjects/{id}/interviews/{chapter}/start", func(w http.ResponseWriter, r *http.Request) {
		chapter := model.Chapter(r.PathValue("chapter"))
		h.StartChapter(w, r, r.PathValue("id"), chapter)
	})
	mux.HandleFunc("POST /api/v1/subjects/{id}/interviews/respond", func(w http.ResponseWriter, r *http.Request) {
		h.RespondChapter(w, r, r.PathValue("id"))
	})
	mux.HandleFunc("POST /api/v1/subjects/{id}/interviews/complete", func(w http.ResponseWriter, r *http.Request) {
		h.CompleteChapter(w, r, r.PathValue("id"))
	})
	mux.HandleFunc("GET /api/v1/interviews/{sessionId}/messages", func(w http.ResponseWriter, r *http.Request) {
		h.GetSessionMessages(w, r, r.PathValue("sessionId"))
	})

	// Media
	mux.HandleFunc("POST /api/v1/subjects/{id}/media/upload", func(w http.ResponseWriter, r *http.Request) {
		h.UploadMedia(w, r, r.PathValue("id"))
	})

	// Memories
	mux.HandleFunc("GET /api/v1/subjects/{id}/memories", func(w http.ResponseWriter, r *http.Request) {
		h.ListMemories(w, r, r.PathValue("id"))
	})

	// Health check
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	// Static files and SPA
	fs := http.FileServer(http.Dir("./web"))
	mux.Handle("GET /", fs)

	return handler
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
