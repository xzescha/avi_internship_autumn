package app

import (
	"avi_internship_autumn/internal/domain"
	"context"
)

type TeamService interface {
	CreateTeam(ctx context.Context, team domain.Team) (domain.Team, error)
	GetTeam(ctx context.Context, teamName string) (domain.Team, error)
}

type UserService interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error)
	GetReviewPRs(ctx context.Context, userID string) ([]domain.PullRequest, error)
}

type PRService interface {
	CreatePR(ctx context.Context, id, name, authorID string) (domain.PullRequest, error)
	MergePR(ctx context.Context, id string) (domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (domain.PullRequest, string, error)
}
