package domain

import (
	"time"
)

// User представляет пользователя системы.
type User struct {
	ID       string
	Username string
	TeamName string
	IsActive bool
}

// Team представляет команду пользователей.
type Team struct {
	Name    string
	Members []User
}

// PRStatus описывает статус pull requestа.
type PRStatus string

const (
	// PRStatusOpen означает, что pull request открыт.
	PRStatusOpen PRStatus = "OPEN"
	// PRStatusMerged означает, что pull request замержен.
	PRStatusMerged PRStatus = "MERGED"
)

// PullRequest представляет pull request в репозитории.
type PullRequest struct {
	ID                string
	Name              string
	AuthorID          string
	Status            PRStatus
	AssignedReviewers []string
	CreatedAt         time.Time
	MergedAt          *time.Time
}
