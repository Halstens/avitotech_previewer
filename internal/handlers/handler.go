package handler

import (
	"encoding/json"
	"net/http"

	"github.com/pavel/avitotech_previewer/internal/database"
)

type Handler struct {
	db  *database.DB
	mux *http.ServeMux
}

func New(db *database.DB) *Handler {
	h := &Handler{
		db:  db,
		mux: http.NewServeMux(),
	}

	h.registerRoutes()
	return h
}

func (h *Handler) registerRoutes() {
	// Health check
	h.mux.HandleFunc("GET /health", h.healthCheck)

	// Teams
	h.mux.HandleFunc("POST /team/add", h.addTeam)
	h.mux.HandleFunc("GET /team/get", h.getTeam)

	// Users
	h.mux.HandleFunc("POST /users/setIsActive", h.setUserActive)
	h.mux.HandleFunc("GET /users/getReview", h.getUserReviews)

	// Pull Requests
	h.mux.HandleFunc("POST /pullRequest/create", h.createPR)
	h.mux.HandleFunc("POST /pullRequest/merge", h.mergePR)
	h.mux.HandleFunc("POST /pullRequest/reassign", h.reassignPR)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

// HealthCheck проверяет доступность сервиса и базы данных
func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.db.HealthCheck(); err != nil {
		h.writeError(w, http.StatusServiceUnavailable, "Service Unavailable", "DATABASE_ERROR")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

// Вспомогательные методы для ответов
func (h *Handler) writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (h *Handler) writeError(w http.ResponseWriter, status int, message, code string) {
	h.writeJSON(w, status, map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}

// Заглушки для обработчиков (будут реализованы в следующих этапах)
func (h *Handler) addTeam(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "Not implemented", "NOT_IMPLEMENTED")
}

func (h *Handler) getTeam(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "Not implemented", "NOT_IMPLEMENTED")
}

func (h *Handler) setUserActive(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "Not implemented", "NOT_IMPLEMENTED")
}

func (h *Handler) getUserReviews(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "Not implemented", "NOT_IMPLEMENTED")
}

func (h *Handler) createPR(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "Not implemented", "NOT_IMPLEMENTED")
}

func (h *Handler) mergePR(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "Not implemented", "NOT_IMPLEMENTED")
}

func (h *Handler) reassignPR(w http.ResponseWriter, r *http.Request) {
	h.writeError(w, http.StatusNotImplemented, "Not implemented", "NOT_IMPLEMENTED")
}
