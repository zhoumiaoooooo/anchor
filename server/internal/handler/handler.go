package handler

import (
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"os"

	"anchor-server/internal/config"
	"anchor-server/internal/model"
	"anchor-server/internal/service/interview"

	"github.com/google/uuid"
)

type Handler struct {
	cfg    *config.Config
	db     *sql.DB
	engine *interview.Engine
}

func New(cfg *config.Config, db *sql.DB, engine *interview.Engine) *Handler {
	return &Handler{cfg: cfg, db: db, engine: engine}
}

// --- Subjects ---

func (h *Handler) CreateSubject(w http.ResponseWriter, r *http.Request) {
	var req model.CreateSubjectRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "无效的请求", http.StatusBadRequest)
		return
	}
	if req.Name == "" || req.Relationship == "" {
		jsonError(w, "名字和关系不能为空", http.StatusBadRequest)
		return
	}

	id := uuid.New().String()
	_, err := h.db.Exec(
		`INSERT INTO subjects (id, name, relationship, birth_year, hometown) VALUES (?, ?, ?, ?, ?)`,
		id, req.Name, req.Relationship, req.BirthYear, req.Hometown,
	)
	if err != nil {
		jsonError(w, "创建失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, http.StatusCreated, map[string]string{"id": id})
}

func (h *Handler) ListSubjects(w http.ResponseWriter, r *http.Request) {
	rows, err := h.db.Query(
		`SELECT id, name, relationship, birth_year, hometown, avatar_path, created_at, updated_at FROM subjects ORDER BY created_at DESC`,
	)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	subjects := []model.Subject{}
	for rows.Next() {
		var s model.Subject
		rows.Scan(&s.ID, &s.Name, &s.Relationship, &s.BirthYear, &s.Hometown, &s.AvatarPath, &s.CreatedAt, &s.UpdatedAt)
		subjects = append(subjects, s)
	}
	jsonOK(w, http.StatusOK, subjects)
}

func (h *Handler) GetSubject(w http.ResponseWriter, r *http.Request, id string) {
	var s model.Subject
	err := h.db.QueryRow(
		`SELECT id, name, relationship, birth_year, hometown, avatar_path, created_at, updated_at FROM subjects WHERE id = ?`, id,
	).Scan(&s.ID, &s.Name, &s.Relationship, &s.BirthYear, &s.Hometown, &s.AvatarPath, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		jsonError(w, "未找到", http.StatusNotFound)
		return
	}
	jsonOK(w, http.StatusOK, s)
}

// --- Interview ---

func (h *Handler) GetInterviewProgress(w http.ResponseWriter, r *http.Request, subjectID string) {
	progress, err := h.engine.GetProgress(subjectID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, http.StatusOK, progress)
}

func (h *Handler) StartChapter(w http.ResponseWriter, r *http.Request, subjectID string, chapter model.Chapter) {
	session, resp, err := h.engine.StartChapter(subjectID, chapter)
	if err != nil {
		jsonError(w, err.Error(), http.StatusBadRequest)
		return
	}
	jsonOK(w, http.StatusOK, map[string]interface{}{
		"session_id": session.ID,
		"message":    resp.Message,
		"chapter":    string(chapter),
	})
}

func (h *Handler) RespondChapter(w http.ResponseWriter, r *http.Request, subjectID string) {
	var req model.RespondRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "无效的请求", http.StatusBadRequest)
		return
	}
	if req.SessionID == "" || req.Content == "" {
		jsonError(w, "session_id 和 content 不能为空", http.StatusBadRequest)
		return
	}
	if !h.sessionBelongsToSubject(w, req.SessionID, subjectID) {
		return
	}

	resp, err := h.engine.ProcessResponse(req.SessionID, req.Content, 0)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, http.StatusOK, resp)
}

func (h *Handler) CompleteChapter(w http.ResponseWriter, r *http.Request, subjectID string) {
	var req struct {
		SessionID string `json:"session_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, "无效的请求", http.StatusBadRequest)
		return
	}
	if req.SessionID == "" {
		jsonError(w, "session_id 不能为空", http.StatusBadRequest)
		return
	}
	if !h.sessionBelongsToSubject(w, req.SessionID, subjectID) {
		return
	}
	count, err := h.engine.CompleteChapter(req.SessionID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	jsonOK(w, http.StatusOK, map[string]interface{}{"fragments_extracted": count})
}

func (h *Handler) GetSessionMessages(w http.ResponseWriter, r *http.Request, sessionID string) {
	messages, err := h.engine.GetSessionMessages(sessionID)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	if messages == nil {
		messages = []model.InterviewMessage{}
	}
	jsonOK(w, http.StatusOK, messages)
}

// --- Media Upload ---

func (h *Handler) UploadMedia(w http.ResponseWriter, r *http.Request, subjectID string) {
	r.ParseMultipartForm(32 << 20)
	file, header, err := r.FormFile("file")
	if err != nil {
		jsonError(w, "需要上传文件", http.StatusBadRequest)
		return
	}
	defer file.Close()

	mediaType := r.FormValue("type")
	if mediaType == "" {
		mediaType = "photo"
	}

	id := uuid.New().String()
	filename := id + "_" + header.Filename
	os.MkdirAll(h.cfg.UploadDir, 0755)
	filePath := h.cfg.UploadDir + "/" + filename

	dst, err := os.Create(filePath)
	if err != nil {
		jsonError(w, "创建文件失败", http.StatusInternalServerError)
		return
	}
	defer dst.Close()
	io.Copy(dst, file)

	h.db.Exec(
		`INSERT INTO media_assets (id, subject_id, type, file_path, original_filename, mime_type, file_size_bytes)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, subjectID, mediaType, filePath, header.Filename, header.Header.Get("Content-Type"), header.Size,
	)
	jsonOK(w, http.StatusCreated, map[string]string{"id": id, "file_path": filePath})
}

// --- Anchor packs (MVP stub) ---

func (h *Handler) ListMemories(w http.ResponseWriter, r *http.Request, subjectID string) {
	rows, err := h.db.Query(
		`SELECT id, subject_id, chapter, raw_text, anchor_potential, has_music, is_procedural, times_used, created_at
		 FROM memory_fragments WHERE subject_id = ? ORDER BY anchor_potential DESC`, subjectID,
	)
	if err != nil {
		jsonError(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	type MemoryItem struct {
		ID              string `json:"id"`
		SubjectID       string `json:"subject_id"`
		Chapter         string `json:"chapter"`
		RawText         string `json:"raw_text"`
		AnchorPotential int    `json:"anchor_potential"`
		HasMusic        bool   `json:"has_music"`
		IsProcedural    bool   `json:"is_procedural"`
		TimesUsed       int    `json:"times_used"`
		CreatedAt       string `json:"created_at"`
	}

	memories := []MemoryItem{}
	for rows.Next() {
		var m MemoryItem
		var hasM, isP int
		rows.Scan(&m.ID, &m.SubjectID, &m.Chapter, &m.RawText, &m.AnchorPotential, &hasM, &isP, &m.TimesUsed, &m.CreatedAt)
		m.HasMusic = hasM == 1
		m.IsProcedural = isP == 1
		memories = append(memories, m)
	}
	jsonOK(w, http.StatusOK, memories)
}

// --- Helpers ---

func jsonOK(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func jsonError(w http.ResponseWriter, msg string, status int) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": msg})
}

func (h *Handler) sessionBelongsToSubject(w http.ResponseWriter, sessionID string, subjectID string) bool {
	ok, err := h.engine.SessionBelongsToSubject(sessionID, subjectID)
	if err != nil {
		jsonError(w, "会话校验失败", http.StatusInternalServerError)
		return false
	}
	if !ok {
		jsonError(w, "会话不属于当前人物", http.StatusForbidden)
		return false
	}
	return true
}
