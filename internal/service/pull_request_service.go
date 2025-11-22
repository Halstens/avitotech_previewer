package service

import (
	"context"
	"math/rand"

	"github.com/pavel/avitotech_previewer/internal/domain"
	"github.com/pavel/avitotech_previewer/internal/repository"
)

type PullRequestService struct {
	prRepo   *repository.PullRequestRepository
	userRepo *repository.UserRepository
}

func NewPullRequestService(prRepo *repository.PullRequestRepository, userRepo *repository.UserRepository) *PullRequestService {
	return &PullRequestService{
		prRepo:   prRepo,
		userRepo: userRepo,
	}
}

func (s *PullRequestService) GetPR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	return s.prRepo.GetPR(ctx, prID)
}

// CreatePR создает PR и назначает до 2 случайных активных ревьюеров из команды автора
func (s *PullRequestService) CreatePR(ctx context.Context, pr *domain.PullRequest) (*domain.PullRequest, error) {
	// Устанавливаем статус
	pr.Status = "OPEN"

	// Получаем команду автора
	teamName, err := s.prRepo.GetUserTeam(ctx, pr.AuthorID)
	if err != nil {
		return nil, err
	}

	// Получаем активных пользователей команды (исключая автора)
	reviewerCandidates, err := s.prRepo.GetTeamActiveUsers(ctx, teamName, pr.AuthorID)
	if err != nil {
		return nil, err
	}

	// Выбираем до 2 случайных ревьюеров
	selectedReviewers := selectRandomReviewers(reviewerCandidates, 2)

	// Создаем PR с назначенными ревьюерами
	if err := s.prRepo.CreatePR(ctx, pr, selectedReviewers); err != nil {
		return nil, err
	}

	pr.AssignedReviewers = selectedReviewers
	return pr, nil
}

// MergePR помечает PR как мердженный
func (s *PullRequestService) MergePR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	return s.prRepo.MergePR(ctx, prID)
}

// ReassignReviewer переназначает ревьюера
func (s *PullRequestService) ReassignReviewer(ctx context.Context, prID string, oldReviewerID string) (string, error) {
	// Получаем PR
	pr, err := s.prRepo.GetPR(ctx, prID)
	if err != nil {
		return "", err
	}

	// Проверяем, что PR не мерджен
	if pr.Status == "MERGED" {
		return "", &domain.Error{Code: "PR_MERGED", Message: "cannot reassign on merged PR"}
	}

	// Проверяем, что старый ревьюер назначен на PR
	if !contains(pr.AssignedReviewers, oldReviewerID) {
		return "", &domain.Error{Code: "NOT_ASSIGNED", Message: "reviewer is not assigned to this PR"}
	}

	// Получаем команду старого ревьюера
	teamName, err := s.prRepo.GetUserTeam(ctx, oldReviewerID)
	if err != nil {
		return "", err
	}

	// Получаем активных пользователей команды (исключая старого ревьюера и автора PR)
	reviewerCandidates, err := s.prRepo.GetTeamActiveUsers(ctx, teamName, oldReviewerID)
	if err != nil {
		return "", err
	}

	// Исключаем автора PR из кандидатов
	reviewerCandidates = excludeUser(reviewerCandidates, pr.AuthorID)

	// Исключаем уже назначенных ревьюеров (кроме старого, которого заменяем)
	reviewerCandidates = excludeUsers(reviewerCandidates, pr.AssignedReviewers)

	if len(reviewerCandidates) == 0 {
		return "", &domain.Error{Code: "NO_CANDIDATE", Message: "no active replacement candidate in team"}
	}

	// Выбираем случайного кандидата
	newReviewerID := reviewerCandidates[rand.Intn(len(reviewerCandidates))]

	// Заменяем ревьюера в списке
	newReviewers := replaceUser(pr.AssignedReviewers, oldReviewerID, newReviewerID)

	// Обновляем PR
	if err := s.prRepo.UpdatePRReviewers(ctx, prID, newReviewers); err != nil {
		return "", err
	}

	return newReviewerID, nil
}

// GetUserReviewPRs возвращает PR, где пользователь назначен ревьюером
func (s *PullRequestService) GetUserReviewPRs(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	return s.prRepo.GetUserReviewPRs(ctx, userID)
}

// Вспомогательные функции

func selectRandomReviewers(candidates []string, max int) []string {
	if len(candidates) == 0 {
		return nil
	}

	if len(candidates) <= max {
		// Перемешиваем кандидатов и возвращаем всех
		shuffled := make([]string, len(candidates))
		copy(shuffled, candidates)
		rand.Shuffle(len(shuffled), func(i, j int) {
			shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
		})
		return shuffled
	}

	// Выбираем max случайных кандидатов
	selected := make([]string, max)
	indices := rand.Perm(len(candidates))
	for i := 0; i < max; i++ {
		selected[i] = candidates[indices[i]]
	}
	return selected
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

func excludeUser(slice []string, userID string) []string {
	var result []string
	for _, s := range slice {
		if s != userID {
			result = append(result, s)
		}
	}
	return result
}

func excludeUsers(slice []string, exclude []string) []string {
	var result []string
	for _, s := range slice {
		if !contains(exclude, s) {
			result = append(result, s)
		}
	}
	return result
}

func replaceUser(slice []string, old, new string) []string {
	result := make([]string, len(slice))
	for i, s := range slice {
		if s == old {
			result[i] = new
		} else {
			result[i] = s
		}
	}
	return result
}
