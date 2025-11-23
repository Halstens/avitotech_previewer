package domain

import (
	"time"
)

type Team struct {
	TeamName string       `json:"team_name" db:"team_name"`
	Members  []TeamMember `json:"members"`
}

type TeamMember struct {
	UserID   string `json:"user_id" db:"user_id"`
	Username string `json:"username" db:"username"`
	IsActive bool   `json:"is_active" db:"is_active"`
}

type User struct {
	UserID   string `json:"user_id" db:"user_id"`
	Username string `json:"username" db:"username"`
	TeamName string `json:"team_name" db:"team_name"`
	IsActive bool   `json:"is_active" db:"is_active"`
}

type PullRequest struct {
	PullRequestID     string     `json:"pull_request_id" db:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name" db:"pull_request_name"`
	AuthorID          string     `json:"author_id" db:"author_id"`
	Status            string     `json:"status" db:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers" db:"-"`
	CreatedAt         *time.Time `json:"createdAt,omitempty" db:"created_at"`
	MergedAt          *time.Time `json:"mergedAt,omitempty" db:"merged_at"`
}

type PullRequestShort struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

type TeamDB struct {
	TeamName  string    `db:"team_name"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type UserDB struct {
	ID        int       `db:"id"`
	UserID    string    `db:"user_id"`
	Username  string    `db:"username"`
	TeamName  string    `db:"team_name"`
	IsActive  bool      `db:"is_active"`
	CreatedAt time.Time `db:"created_at"`
	UpdatedAt time.Time `db:"updated_at"`
}

type PullRequestDB struct {
	ID              int        `db:"id"`
	PullRequestID   string     `db:"pull_request_id"`
	PullRequestName string     `db:"pull_request_name"`
	AuthorID        string     `db:"author_id"`
	Status          string     `db:"status"`
	CreatedAt       *time.Time `db:"created_at"`
	MergedAt        *time.Time `db:"merged_at"`
	UpdatedAt       time.Time  `db:"updated_at"`
}

type PullRequestReviewerDB struct {
	ID            int       `db:"id"`
	PullRequestID string    `db:"pull_request_id"`
	ReviewerID    string    `db:"reviewer_id"`
	AssignedAt    time.Time `db:"assigned_at"`
}
