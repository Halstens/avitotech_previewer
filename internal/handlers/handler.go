// internal/handler/handler.go
package handler

import (
	"net/http"

	"github.com/pavel/avitotech_previewer/internal/database"
	"github.com/pavel/avitotech_previewer/internal/repository"
	"github.com/pavel/avitotech_previewer/internal/service"
)

type Handler struct {
	*BaseHandler
	db                      *database.DB
	mux                     *http.ServeMux
	teamHandler             *TeamHandler
	userHandler             *UserHandler
	prHandler               *PullRequestHandler
	statsHandler            *StatsHandler
	bulkDeactivationHandler *BulkDeactivationHandler
}

func New(db *database.DB) *Handler {
	teamRepo := repository.NewTeamRepository(db.DB)
	userRepo := repository.NewUserRepository(db.DB)
	prRepo := repository.NewPullRequestRepository(db.DB)
	prService := service.NewPullRequestService(prRepo, userRepo)
	statsRepo := repository.NewStatsRepository(db.DB)
	bulkService := service.NewBulkDeactivationService(userRepo, prRepo, prService)

	h := &Handler{
		BaseHandler:             &BaseHandler{},
		db:                      db,
		mux:                     http.NewServeMux(),
		teamHandler:             NewTeamHandler(teamRepo),
		userHandler:             NewUserHandler(userRepo, prService),
		prHandler:               NewPullRequestHandler(prService),
		statsHandler:            NewStatsHandler(statsRepo),
		bulkDeactivationHandler: NewBulkDeactivationHandler(bulkService),
	}

	h.registerRoutes()
	return h
}

func (h *Handler) registerRoutes() {

	h.mux.HandleFunc("GET /health", h.healthCheck)

	h.mux.HandleFunc("POST /team/add", h.teamHandler.AddTeam)
	h.mux.HandleFunc("GET /team/get", h.teamHandler.GetTeam)

	h.mux.HandleFunc("POST /users/setIsActive", h.userHandler.SetUserActive)
	h.mux.HandleFunc("GET /users/getReview", h.userHandler.GetUserReviews)

	h.mux.HandleFunc("POST /pullRequest/create", h.prHandler.CreatePR)
	h.mux.HandleFunc("POST /pullRequest/merge", h.prHandler.MergePR)
	h.mux.HandleFunc("POST /pullRequest/reassign", h.prHandler.ReassignPR)

	h.mux.HandleFunc("GET /stats", h.statsHandler.GetStats)

	h.mux.HandleFunc("POST /team/bulkDeactivate", h.bulkDeactivationHandler.BulkDeactivateTeam)
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.mux.ServeHTTP(w, r)
}

func (h *Handler) healthCheck(w http.ResponseWriter, r *http.Request) {
	if err := h.db.HealthCheck(); err != nil {
		h.writeError(w, http.StatusServiceUnavailable, "Service Unavailable", "DATABASE_ERROR")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}
