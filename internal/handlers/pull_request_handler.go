package handler

import (
	"encoding/json"
	"net/http"

	"github.com/pavel/avitotech_previewer/internal/domain"
	"github.com/pavel/avitotech_previewer/internal/service"
)

type PullRequestHandler struct {
	*BaseHandler
	prService *service.PullRequestService
}

func NewPullRequestHandler(prService *service.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{
		BaseHandler: &BaseHandler{},
		prService:   prService,
	}
}

func (h *PullRequestHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	pr := &domain.PullRequest{
		PullRequestID:   request.PullRequestID,
		PullRequestName: request.PullRequestName,
		AuthorID:        request.AuthorID,
	}

	createdPR, err := h.prService.CreatePR(r.Context(), pr)
	if err != nil {
		switch {
		case domain.IsDomainError(err, "PR_EXISTS"):
			h.writeError(w, http.StatusConflict, "PR id already exists", "PR_EXISTS")
		case domain.IsDomainError(err, "NOT_FOUND"):
			h.writeError(w, http.StatusNotFound, "author/team not found", "NOT_FOUND")
		default:
			h.writeError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR")
		}
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"pr": createdPR,
	})
}

func (h *PullRequestHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PullRequestID string `json:"pull_request_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	pr, err := h.prService.MergePR(r.Context(), request.PullRequestID)
	if err != nil {
		if domain.IsDomainError(err, "NOT_FOUND") {
			h.writeError(w, http.StatusNotFound, "PR not found", "NOT_FOUND")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"pr": pr,
	})
}

func (h *PullRequestHandler) ReassignPR(w http.ResponseWriter, r *http.Request) {
	var request struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_reviewer_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	newReviewerID, err := h.prService.ReassignReviewer(r.Context(), request.PullRequestID, request.OldUserID)
	if err != nil {
		switch {
		case domain.IsDomainError(err, "NOT_FOUND"):
			h.writeError(w, http.StatusNotFound, "PR or user not found", "NOT_FOUND")
		case domain.IsDomainError(err, "PR_MERGED"):
			h.writeError(w, http.StatusConflict, "cannot reassign on merged PR", "PR_MERGED")
		case domain.IsDomainError(err, "NOT_ASSIGNED"):
			h.writeError(w, http.StatusConflict, "reviewer is not assigned to this PR", "NOT_ASSIGNED")
		case domain.IsDomainError(err, "NO_CANDIDATE"):
			h.writeError(w, http.StatusConflict, "no active replacement candidate in team", "NO_CANDIDATE")
		default:
			h.writeError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR")
		}
		return
	}

	pr, err := h.prService.GetPR(r.Context(), request.PullRequestID)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"pr":          pr,
		"replaced_by": newReviewerID,
	})
}
