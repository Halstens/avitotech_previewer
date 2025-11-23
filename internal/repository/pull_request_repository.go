package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pavel/avitotech_previewer/internal/domain"
)

type PullRequestRepository struct {
	db *sql.DB
}

func NewPullRequestRepository(db *sql.DB) *PullRequestRepository {
	return &PullRequestRepository{db: db}
}

func (r *PullRequestRepository) GetUserReviewPRs(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		JOIN pull_request_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.reviewer_id = $1
		ORDER BY pr.created_at DESC`,
		userID)
	if err != nil {
		return nil, fmt.Errorf("failed to query user review PRs: %w", err)
	}
	defer rows.Close()

	var prs []domain.PullRequestShort
	for rows.Next() {
		var pr domain.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("failed to scan PR: %w", err)
		}
		prs = append(prs, pr)
	}

	return prs, nil
}

func (r *PullRequestRepository) CreatePR(ctx context.Context, pr *domain.PullRequest, reviewerIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)",
		pr.PullRequestID).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check PR existence: %w", err)
	}
	if exists {
		return &domain.Error{Code: "PR_EXISTS", Message: "PR id already exists"}
	}

	var authorExists bool
	err = tx.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM users WHERE user_id = $1 AND is_active = true)",
		pr.AuthorID).Scan(&authorExists)
	if err != nil {
		return fmt.Errorf("failed to check author existence: %w", err)
	}
	if !authorExists {
		return &domain.Error{Code: "NOT_FOUND", Message: "author not found or inactive"}
	}

	now := time.Now()
	_, err = tx.ExecContext(ctx, `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
		VALUES ($1, $2, $3, $4, $5)`,
		pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status, now)
	if err != nil {
		return fmt.Errorf("failed to insert PR: %w", err)
	}

	for _, reviewerID := range reviewerIDs {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
			VALUES ($1, $2)`,
			pr.PullRequestID, reviewerID)
		if err != nil {
			return fmt.Errorf("failed to assign reviewer %s: %w", reviewerID, err)
		}
	}

	return tx.Commit()
}

func (r *PullRequestRepository) GetPR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	var pr domain.PullRequest
	err := r.db.QueryRowContext(ctx, `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests 
		WHERE pull_request_id = $1`,
		prID).Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err == sql.ErrNoRows {
		return nil, &domain.Error{Code: "NOT_FOUND", Message: "PR not found"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get PR: %w", err)
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT reviewer_id 
		FROM pull_request_reviewers 
		WHERE pull_request_id = $1`,
		prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	return &pr, nil
}

func (r *PullRequestRepository) MergePR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var pr domain.PullRequest
	now := time.Now()
	err = tx.QueryRowContext(ctx, `
		UPDATE pull_requests 
		SET status = 'MERGED', merged_at = $1, updated_at = CURRENT_TIMESTAMP
		WHERE pull_request_id = $2
		RETURNING pull_request_id, pull_request_name, author_id, status, created_at, merged_at`,
		&now, prID).Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err == sql.ErrNoRows {
		return nil, &domain.Error{Code: "NOT_FOUND", Message: "PR not found"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to merge PR: %w", err)
	}

	rows, err := tx.QueryContext(ctx, `
		SELECT reviewer_id 
		FROM pull_request_reviewers 
		WHERE pull_request_id = $1`,
		prID)
	if err != nil {
		return nil, fmt.Errorf("failed to get reviewers: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, fmt.Errorf("failed to scan reviewer: %w", err)
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return &pr, nil
}

func (r *PullRequestRepository) UpdatePRReviewers(ctx context.Context, prID string, reviewerIDs []string) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		DELETE FROM pull_request_reviewers 
		WHERE pull_request_id = $1`,
		prID)
	if err != nil {
		return fmt.Errorf("failed to delete old reviewers: %w", err)
	}

	for _, reviewerID := range reviewerIDs {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO pull_request_reviewers (pull_request_id, reviewer_id)
			VALUES ($1, $2)`,
			prID, reviewerID)
		if err != nil {
			return fmt.Errorf("failed to assign reviewer %s: %w", reviewerID, err)
		}
	}

	return tx.Commit()
}

func (r *PullRequestRepository) GetTeamActiveUsers(ctx context.Context, teamName string, excludeUserID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id 
		FROM users 
		WHERE team_name = $1 AND is_active = true AND user_id != $2
		ORDER BY user_id`,
		teamName, excludeUserID)
	if err != nil {
		return nil, fmt.Errorf("failed to query team users: %w", err)
	}
	defer rows.Close()

	var userIDs []string
	for rows.Next() {
		var userID string
		if err := rows.Scan(&userID); err != nil {
			return nil, fmt.Errorf("failed to scan user ID: %w", err)
		}
		userIDs = append(userIDs, userID)
	}

	return userIDs, nil
}

func (r *PullRequestRepository) GetUserTeam(ctx context.Context, userID string) (string, error) {
	var teamName string
	err := r.db.QueryRowContext(ctx, `
		SELECT team_name 
		FROM users 
		WHERE user_id = $1`,
		userID).Scan(&teamName)
	if err == sql.ErrNoRows {
		return "", &domain.Error{Code: "NOT_FOUND", Message: "user not found"}
	}
	if err != nil {
		return "", fmt.Errorf("failed to get user team: %w", err)
	}
	return teamName, nil
}

func (r *PullRequestRepository) GetOpenPRsWithReviewer(ctx context.Context, reviewerID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT DISTINCT pr.pull_request_id
		FROM pull_requests pr
		JOIN pull_request_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.reviewer_id = $1 AND pr.status = 'OPEN'
	`, reviewerID)
	if err != nil {
		return nil, fmt.Errorf("failed to query PRs with reviewer: %w", err)
	}
	defer rows.Close()

	var prIDs []string
	for rows.Next() {
		var prID string
		if err := rows.Scan(&prID); err != nil {
			return nil, fmt.Errorf("failed to scan PR ID: %w", err)
		}
		prIDs = append(prIDs, prID)
	}

	return prIDs, nil
}
