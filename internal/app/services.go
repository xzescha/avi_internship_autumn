package app

import (
	"avi_internship_autumn/internal/domain"
	"context"
)

// TeamService описывает операции над командами.
type TeamService interface {
	CreateTeam(ctx context.Context, team domain.Team) (domain.Team, error)
	GetTeam(ctx context.Context, teamName string) (domain.Team, error)
}

// UserService описывает операции над пользователями.
type UserService interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error)
	GetReviewPRs(ctx context.Context, userID string) ([]domain.PullRequest, error)
	BulkDeactivateTeam(ctx context.Context, teamName string, userIDs []string) (domain.BulkDeactivateResult, error)
}

// PRService описывает операции над pull requestами.
type PRService interface {
	CreatePR(ctx context.Context, id, name, authorID string) (domain.PullRequest, error)
	MergePR(ctx context.Context, id string) (domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (domain.PullRequest, string, error)

	GetAssignmentStatsByReviewer(ctx context.Context) ([]domain.AssignmentStats, error)
	GetAssignmentStatsByPR(ctx context.Context) ([]domain.PullRequestAssignmentStats, error)
}
