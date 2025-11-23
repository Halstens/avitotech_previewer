package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pavel/avitotech_previewer/internal/domain"
)

type TeamRepository struct {
	db *sql.DB
}

func NewTeamRepository(db *sql.DB) *TeamRepository {
	return &TeamRepository{db: db}
}

func (r *TeamRepository) CreateTeam(ctx context.Context, team *domain.Team) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	var exists bool
	err = tx.QueryRowContext(ctx,
		"SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)",
		team.TeamName).Scan(&exists)
	if err != nil {
		return fmt.Errorf("failed to check team existence: %w", err)
	}
	if exists {
		return &domain.Error{Code: "TEAM_EXISTS", Message: "team already exists"}
	}

	_, err = tx.ExecContext(ctx,
		"INSERT INTO teams (team_name) VALUES ($1)",
		team.TeamName)
	if err != nil {
		return fmt.Errorf("failed to insert team: %w", err)
	}

	for _, member := range team.Members {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO users (user_id, username, team_name, is_active) 
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id) 
			DO UPDATE SET username = $2, team_name = $3, is_active = $4, updated_at = CURRENT_TIMESTAMP`,
			member.UserID, member.Username, team.TeamName, member.IsActive)
		if err != nil {
			return fmt.Errorf("failed to upsert user %s: %w", member.UserID, err)
		}
	}

	return tx.Commit()
}

func (r *TeamRepository) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	var team domain.Team
	team.TeamName = teamName

	rows, err := r.db.QueryContext(ctx, `
		SELECT user_id, username, is_active 
		FROM users 
		WHERE team_name = $1 
		ORDER BY user_id`,
		teamName)
	if err != nil {
		return nil, fmt.Errorf("failed to query team members: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var member domain.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, fmt.Errorf("failed to scan team member: %w", err)
		}
		team.Members = append(team.Members, member)
	}

	if len(team.Members) == 0 {
		return nil, &domain.Error{Code: "NOT_FOUND", Message: "team not found"}
	}

	return &team, nil
}
