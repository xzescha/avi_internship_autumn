// Package repository содержит интерфейсы доступа к хранилищам данных.
package repository

import (
	"avi_internship_autumn/internal/domain"
	"context"
)

// TeamRepository определяет операции над хранилищем команд.
type TeamRepository interface {
	Create(ctx context.Context, teamName string) error
	Exists(ctx context.Context, teamName string) (bool, error)
	Get(ctx context.Context, teamName string) (domain.Team, error)
}

// UserRepository определяет операции над хранилищем пользователей.
type UserRepository interface {
	Upsert(ctx context.Context, u domain.User) error
	GetByID(ctx context.Context, id string) (domain.User, error)
	ListByTeam(ctx context.Context, teamName string) ([]domain.User, error)
	UpdateIsActive(ctx context.Context, id string, isActive bool) (domain.User, error)
	BulkDeactivateInTeam(ctx context.Context, teamName string, userIDs []string) (int64, error)
}

// PRRepository определяет операции над хранилищем pull requestов.
type PRRepository interface {
	Exists(ctx context.Context, id string) (bool, error)
	Create(ctx context.Context, pr domain.PullRequest) error
	GetForUpdate(ctx context.Context, id string) (domain.PullRequest, error)
	UpdateStatusMerged(ctx context.Context, id string) error
	ListReviewerPRs(ctx context.Context, reviewerID string) ([]domain.PullRequest, error)

	GetReviewers(ctx context.Context, prID string) ([]string, error)
	AddReviewer(ctx context.Context, prID, reviewerID string) error
	RemoveReviewer(ctx context.Context, prID, reviewerID string) error

	GetAssignmentStatsByReviewer(ctx context.Context) ([]domain.AssignmentStats, error)
	GetAssignmentStatsByPR(ctx context.Context) ([]domain.PullRequestAssignmentStats, error)
	ListOpenPRsByReviewers(ctx context.Context, reviewerIDs []string) ([]domain.PullRequest, error)
}
