package handler

import (
	"encoding/json"
	"net/http"

	"github.com/pavel/avitotech_previewer/internal/domain"
	"github.com/pavel/avitotech_previewer/internal/repository"
	"github.com/pavel/avitotech_previewer/internal/service"
)

type UserHandler struct {
	*BaseHandler
	userRepo  *repository.UserRepository
	prService *service.PullRequestService
}

func NewUserHandler(userRepo *repository.UserRepository, prService *service.PullRequestService) *UserHandler {
	return &UserHandler{
		BaseHandler: &BaseHandler{},
		userRepo:    userRepo,
		prService:   prService,
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

func (h *UserHandler) GetUserReviews(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		h.writeError(w, http.StatusBadRequest, "user_id parameter is required", "MISSING_PARAMETER")
		return
	}

	// Проверяем существование пользователя
	_, err := h.userRepo.GetUserByID(r.Context(), userID)
	if err != nil {
		if domain.IsDomainError(err, "NOT_FOUND") {
			h.writeError(w, http.StatusNotFound, "user not found", "NOT_FOUND")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR")
		return
	}

	// Получаем PR пользователя как ревьюера
	prs, err := h.prService.GetUserReviewPRs(r.Context(), userID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	})
}
