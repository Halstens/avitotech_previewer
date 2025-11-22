package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pavel/avitotech_previewer/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) UpdateUserActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	var user domain.User

	err := r.db.QueryRowContext(ctx, `
		UPDATE users 
		SET is_active = $1, updated_at = CURRENT_TIMESTAMP 
		WHERE user_id = $2 
		RETURNING user_id, username, team_name, is_active`,
		isActive, userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)

	if err == sql.ErrNoRows {
		return nil, &domain.Error{Code: "NOT_FOUND", Message: "user not found"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return &user, nil
}

func (r *UserRepository) GetUserByID(ctx context.Context, userID string) (*domain.User, error) {
	var user domain.User

	err := r.db.QueryRowContext(ctx, `
		SELECT user_id, username, team_name, is_active 
		FROM users 
		WHERE user_id = $1`,
		userID).Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive)

	if err == sql.ErrNoRows {
		return nil, &domain.Error{Code: "NOT_FOUND", Message: "user not found"}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return &user, nil
}
