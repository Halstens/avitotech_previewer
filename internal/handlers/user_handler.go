// internal/handler/user_handler.go
package handler

import (
	"encoding/json"
	"net/http"

	"github.com/pavel/avitotech_previewer/internal/domain"
	"github.com/pavel/avitotech_previewer/internal/repository"
)

type UserHandler struct {
	*BaseHandler
	userRepo *repository.UserRepository
}

func NewUserHandler(userRepo *repository.UserRepository) *UserHandler {
	return &UserHandler{
		BaseHandler: &BaseHandler{},
		userRepo:    userRepo,
	}
}

func (h *UserHandler) SetUserActive(w http.ResponseWriter, r *http.Request) {
	var request struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	user, err := h.userRepo.UpdateUserActive(r.Context(), request.UserID, request.IsActive)
	if err != nil {
		if domain.IsDomainError(err, "NOT_FOUND") {
			h.writeError(w, http.StatusNotFound, "user not found", "NOT_FOUND")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"user": user,
	})
}
