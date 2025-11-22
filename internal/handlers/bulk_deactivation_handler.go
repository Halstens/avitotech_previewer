package handler

import (
	"encoding/json"
	"net/http"

	"github.com/pavel/avitotech_previewer/internal/service"
)

type BulkDeactivationHandler struct {
	*BaseHandler
	bulkService *service.BulkDeactivationService
}

func NewBulkDeactivationHandler(bulkService *service.BulkDeactivationService) *BulkDeactivationHandler {
	return &BulkDeactivationHandler{
		BaseHandler: &BaseHandler{},
		bulkService: bulkService,
	}
}

func (h *BulkDeactivationHandler) BulkDeactivateTeam(w http.ResponseWriter, r *http.Request) {
	var request struct {
		TeamName       string   `json:"team_name"`
		ExcludeUserIDs []string `json:"exclude_user_ids,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	if request.TeamName == "" {
		h.writeError(w, http.StatusBadRequest, "team_name is required", "MISSING_PARAMETER")
		return
	}

	result, err := h.bulkService.BulkDeactivateTeam(r.Context(), request.TeamName, request.ExcludeUserIDs)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"result": result,
	})
}
