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

func (r *UserRepository) BulkDeactivateUsers(ctx context.Context, teamName string, excludeUserIDs []string) (int64, error) {
	query := "UPDATE users SET is_active = false, updated_at = CURRENT_TIMESTAMP WHERE team_name = $1"
	args := []interface{}{teamName}

	if len(excludeUserIDs) > 0 {
		query += " AND user_id NOT IN ("
		for i, id := range excludeUserIDs {
			if i > 0 {
				query += ","
			}
			query += fmt.Sprintf("$%d", i+2)
			args = append(args, id)
		}
		query += ")"
	}

	result, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return 0, fmt.Errorf("failed to deactivate users: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}

func (r *UserRepository) GetTeamUsers(ctx context.Context, teamName string) ([]domain.User, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id, username, team_name, is_active
		FROM users 
		WHERE team_name = $1
		ORDER BY user_id`,
		teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to query team users: %w", err)
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan user: %w", err)
		}
		users = append(users, user)
	}

	return users, nil
}
