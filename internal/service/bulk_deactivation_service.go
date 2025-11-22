package service

import (
	"context"
	"time"

	"github.com/pavel/avitotech_previewer/internal/domain"
	"github.com/pavel/avitotech_previewer/internal/repository"
)

type BulkDeactivationService struct {
	userRepo  *repository.UserRepository
	prRepo    *repository.PullRequestRepository
	prService *PullRequestService
}

func NewBulkDeactivationService(userRepo *repository.UserRepository, prRepo *repository.PullRequestRepository, prService *PullRequestService) *BulkDeactivationService {
	return &BulkDeactivationService{
		userRepo:  userRepo,
		prRepo:    prRepo,
		prService: prService,
	}
}

type BulkDeactivationResult struct {
	DeactivatedCount int64          `json:"deactivated_count"`
	ReassignedPRs    []ReassignedPR `json:"reassigned_prs"`
	Timestamp        time.Time      `json:"timestamp"`
}

type ReassignedPR struct {
	PRID        string `json:"pr_id"`
	OldReviewer string `json:"old_reviewer"`
	NewReviewer string `json:"new_reviewer"`
}

func (s *BulkDeactivationService) BulkDeactivateTeam(ctx context.Context, teamName string, excludeUserIDs []string) (*BulkDeactivationResult, error) {
	// Получаем пользователей команды до деактивации
	teamUsers, err := s.userRepo.GetTeamUsers(ctx, teamName)
	if err != nil {
		return nil, err
	}

	if len(teamUsers) == 0 {
		return nil, &domain.Error{Code: "NOT_FOUND", Message: "team not found"}
	}

	// Деактивируем пользователей
	deactivatedCount, err := s.userRepo.BulkDeactivateUsers(ctx, teamName, excludeUserIDs)
	if err != nil {
		return nil, err
	}

	// Находим пользователей, которых нужно деактивировать
	usersToDeactivate := s.getUsersToDeactivate(teamUsers, excludeUserIDs)

	// Переназначаем открытые PR для деактивированных пользователей
	reassignedPRs, err := s.reassignPRsForDeactivatedUsers(ctx, usersToDeactivate)
	if err != nil {
		return nil, err
	}

	result := &BulkDeactivationResult{
		DeactivatedCount: deactivatedCount,
		ReassignedPRs:    reassignedPRs,
		Timestamp:        time.Now(),
	}

	return result, nil
}

func (s *BulkDeactivationService) getUsersToDeactivate(teamUsers []domain.User, excludeUserIDs []string) []domain.User {
	excludeSet := make(map[string]bool)
	for _, id := range excludeUserIDs {
		excludeSet[id] = true
	}

	var usersToDeactivate []domain.User
	for _, user := range teamUsers {
		if !excludeSet[user.UserID] && user.IsActive {
			usersToDeactivate = append(usersToDeactivate, user)
		}
	}

	return usersToDeactivate
}

func (s *BulkDeactivationService) reassignPRsForDeactivatedUsers(ctx context.Context, usersToDeactivate []domain.User) ([]ReassignedPR, error) {
	var reassignedPRs []ReassignedPR

	for _, user := range usersToDeactivate {
		// Находим открытые PR, где пользователь является ревьюером
		prIDs, err := s.prRepo.GetOpenPRsWithReviewer(ctx, user.UserID)
		if err != nil {
			return nil, err
		}

		for _, prID := range prIDs {
			// Пытаемся переназначить ревьюера
			newReviewer, err := s.prService.ReassignReviewer(ctx, prID, user.UserID)
			if err != nil {
				// Если не удалось переназначить, пропускаем (можно добавить логирование)
				continue
			}

			reassignedPRs = append(reassignedPRs, ReassignedPR{
				PRID:        prID,
				OldReviewer: user.UserID,
				NewReviewer: newReviewer,
			})
		}
	}

	return reassignedPRs, nil
}
