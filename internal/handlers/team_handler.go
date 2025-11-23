package handler

import (
	"encoding/json"
	"net/http"

	"github.com/pavel/avitotech_previewer/internal/domain"
	"github.com/pavel/avitotech_previewer/internal/repository"
)

type TeamHandler struct {
	*BaseHandler
	teamRepo *repository.TeamRepository
}

func NewTeamHandler(teamRepo *repository.TeamRepository) *TeamHandler {
	return &TeamHandler{
		BaseHandler: &BaseHandler{},
		teamRepo:    teamRepo,
	}
}

func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	var team domain.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		h.writeError(w, http.StatusBadRequest, "Invalid request body", "INVALID_REQUEST")
		return
	}

	if team.TeamName == "" || len(team.Members) == 0 {
		h.writeError(w, http.StatusBadRequest, "Team name and members are required", "INVALID_REQUEST")
		return
	}

	if err := h.teamRepo.CreateTeam(r.Context(), &team); err != nil {
		if domain.IsDomainError(err, "TEAM_EXISTS") {
			h.writeError(w, http.StatusBadRequest, "team_name already exists", "TEAM_EXISTS")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR")
		return
	}

	h.writeJSON(w, http.StatusCreated, map[string]interface{}{
		"team": team,
	})
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		h.writeError(w, http.StatusBadRequest, "team_name parameter is required", "MISSING_PARAMETER")
		return
	}

	team, err := h.teamRepo.GetTeam(r.Context(), teamName)
	if err != nil {
		if domain.IsDomainError(err, "NOT_FOUND") {
			h.writeError(w, http.StatusNotFound, "team not found", "NOT_FOUND")
			return
		}
		h.writeError(w, http.StatusInternalServerError, "Internal server error", "INTERNAL_ERROR")
		return
	}

	h.writeJSON(w, http.StatusOK, team)
}
