package repository

import (
	"context"
	"database/sql"
	"fmt"
)

type StatsRepository struct {
	db *sql.DB
}

func NewStatsRepository(db *sql.DB) *StatsRepository {
	return &StatsRepository{db: db}
}

// GetStats возвращает статистику по PR и назначениям
func (r *StatsRepository) GetStats(ctx context.Context) (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Общее количество PR по статусам
	var openCount, mergedCount int
	err := r.db.QueryRowContext(ctx, `
		SELECT 
			COUNT(*) FILTER (WHERE status = 'OPEN') as open_count,
			COUNT(*) FILTER (WHERE status = 'MERGED') as merged_count
		FROM pull_requests
	`).Scan(&openCount, &mergedCount)
	if err != nil {
		return nil, fmt.Errorf("failed to get PR counts: %w", err)
	}

	stats["pull_requests"] = map[string]int{
		"open":   openCount,
		"merged": mergedCount,
		"total":  openCount + mergedCount,
	}

	// Количество назначений по пользователям
	rows, err := r.db.QueryContext(ctx, `
		SELECT u.user_id, u.username, COUNT(prr.pull_request_id) as assignment_count
		FROM users u
		LEFT JOIN pull_request_reviewers prr ON u.user_id = prr.reviewer_id
		GROUP BY u.user_id, u.username
		ORDER BY assignment_count DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get assignment stats: %w", err)
	}
	defer rows.Close()

	assignments := make([]map[string]interface{}, 0)
	for rows.Next() {
		var userID, username string
		var count int
		if err := rows.Scan(&userID, &username, &count); err != nil {
			return nil, fmt.Errorf("failed to scan assignment: %w", err)
		}
		assignments = append(assignments, map[string]interface{}{
			"user_id":          userID,
			"username":         username,
			"assignment_count": count,
		})
	}
	stats["assignments"] = assignments

	// Статистика по командам
	teamStats, err := r.getTeamStats(ctx)
	if err != nil {
		return nil, err
	}
	stats["teams"] = teamStats

	return stats, nil
}

// getTeamStats возвращает статистику по командам
func (r *StatsRepository) getTeamStats(ctx context.Context) (map[string]interface{}, error) {
	// Количество пользователей по командам
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			team_name,
			COUNT(*) as total_users,
			COUNT(*) FILTER (WHERE is_active = true) as active_users,
			COUNT(*) FILTER (WHERE is_active = false) as inactive_users
		FROM users
		GROUP BY team_name
		ORDER BY team_name
	`)
	if err != nil {
		return nil, fmt.Errorf("failed to get team stats: %w", err)
	}
	defer rows.Close()

	teamStats := make(map[string]interface{})
	teams := make([]map[string]interface{}, 0)

	for rows.Next() {
		var teamName string
		var total, active, inactive int
		if err := rows.Scan(&teamName, &total, &active, &inactive); err != nil {
			return nil, fmt.Errorf("failed to scan team stats: %w", err)
		}

		teamData := map[string]interface{}{
			"team_name":      teamName,
			"total_users":    total,
			"active_users":   active,
			"inactive_users": inactive,
		}
		teams = append(teams, teamData)
	}

	teamStats["summary"] = teams

	// Общая статистика по командам
	var totalTeams, totalUsers int
	err = r.db.QueryRowContext(ctx, `
		SELECT 
			COUNT(DISTINCT team_name) as total_teams,
			COUNT(*) as total_users
		FROM users
	`).Scan(&totalTeams, &totalUsers)
	if err != nil {
		return nil, fmt.Errorf("failed to get overall team stats: %w", err)
	}

	teamStats["overall"] = map[string]int{
		"total_teams": totalTeams,
		"total_users": totalUsers,
	}

	return teamStats, nil
}
