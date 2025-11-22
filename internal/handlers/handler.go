// internal/handler/handler.go
package handler

import (
	"net/http"

	"github.com/pavel/avitotech_previewer/internal/database"
	"github.com/pavel/avitotech_previewer/internal/repository"
)

type Handler struct {
	*BaseHandler
	db          *database.DB
	mux         *http.ServeMux
	teamHandler *TeamHandler
	userHandler *UserHandler
}

func New(db *database.DB) *Handler {
	teamRepo := repository.NewTeamRepository(db.DB)
	userRepo := repository.NewUserRepository(db.DB)

	h := &Handler{
		BaseHandler: &BaseHandler{},
		db:          db,
		mux:         http.NewServeMux(),
		teamHandler: NewTeamHandler(teamRepo),
		userHandler: NewUserHandler(userRepo),
	}

	h.registerRoutes()
	return h
}

func (h *Handler) registerRoutes() {
	// Health check
	h.mux.HandleFunc("GET /health", h.healthCheck)

	// Teams
	h.mux.HandleFunc("POST /team/add", h.teamHandler.AddTeam)
	h.mux.HandleFunc("GET /team/get", h.teamHandler.GetTeam)

	// Users
	h.mux.HandleFunc("POST /users/setIsActive", h.userHandler.SetUserActive)
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

// Заглушки для обработчиков (будут реализованы в следующих этапах)
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
