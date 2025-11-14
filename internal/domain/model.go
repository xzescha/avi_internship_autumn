package domain

import (
	"time"
)

type User struct {
	ID       string
	Username string
	TeamName string
	IsActive bool
}

type Team struct {
	Name    string
	Members []User
}

type PRStatus string

const (
	PRStatusOpen   PRStatus = "OPEN"
	PRStatusMerged PRStatus = "MERGED"
)

type PullRequest struct {
	ID                string
	Name              string
	AuthorID          string
	Status            PRStatus
	AssignedReviewers []string
	CreatedAt         time.Time
	MergedAt          *time.Time
}
