package handler

import (
	"net/http"

	"github.com/pavel/avitotech_previewer/internal/repository"
)

type StatsHandler struct {
	*BaseHandler
	statsRepo *repository.StatsRepository
}

func NewStatsHandler(statsRepo *repository.StatsRepository) *StatsHandler {
	return &StatsHandler{
		BaseHandler: &BaseHandler{},
		statsRepo:   statsRepo,
	}
}

func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	stats, err := h.statsRepo.GetStats(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]interface{}{
		"stats": stats,
	})
}
