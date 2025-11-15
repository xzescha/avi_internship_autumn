// Package service сервисный слой с логикой
package service

import (
	"context"
	"errors"
	"math/rand"
	"time"

	"avi_internship_autumn/internal/app"
	"avi_internship_autumn/internal/domain"
	"avi_internship_autumn/internal/repository"
)

type prService struct {
	prs   repository.PRRepository
	users repository.UserRepository
}

// NewPRService создаёт сервис для работы с pull requestами.
func NewPRService(
	prs repository.PRRepository,
	users repository.UserRepository,
) app.PRService {
	return &prService{
		prs:   prs,
		users: users,
	}
}

// CreatePR создает PR и назначает до 2 ревьюверов из команды автора.
func (s *prService) CreatePR(ctx context.Context, id, name, authorID string) (domain.PullRequest, error) {
	exists, err := s.prs.Exists(ctx, id)
	if err != nil {
		return domain.PullRequest{}, err
	}
	if exists {
		return domain.PullRequest{}, domain.ErrPRExists
	}

	author, err := s.users.GetByID(ctx, authorID)
	if err != nil {
		// ожидается domain.ErrNotFound, который наверху превратится в 404
		return domain.PullRequest{}, err
	}

	members, err := s.listTeamMembers(ctx, author.TeamName)
	if err != nil {
		return domain.PullRequest{}, err
	}

	candidates := filterActiveExcept(members, author.ID)

	reviewerIDs := pickRandomUserIDs(candidates, 2)

	pr := domain.PullRequest{
		ID:                id,
		Name:              name,
		AuthorID:          authorID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: reviewerIDs,
	}

	if err := s.prs.Create(ctx, pr); err != nil {
		return domain.PullRequest{}, err
	}

	for _, rid := range reviewerIDs {
		if err := s.prs.AddReviewer(ctx, id, rid); err != nil {
			return domain.PullRequest{}, err
		}
	}

	return pr, nil
}

// MergePR делает merge PR.
func (s *prService) MergePR(ctx context.Context, id string) (domain.PullRequest, error) {
	pr, err := s.prs.GetForUpdate(ctx, id)
	if err != nil {
		// ожидается domain.ErrNotFound, который наверху превратится в 404
		return domain.PullRequest{}, err
	}

	if pr.IsMerged() {
		reviewers, err := s.prs.GetReviewers(ctx, id)
		if err != nil {
			return domain.PullRequest{}, err
		}
		pr.AssignedReviewers = reviewers
		return pr, nil
	}

	if err := s.prs.UpdateStatusMerged(ctx, id); err != nil {
		return domain.PullRequest{}, err
	}

	pr, err = s.prs.GetForUpdate(ctx, id)
	if err != nil {
		return domain.PullRequest{}, err
	}

	reviewers, err := s.prs.GetReviewers(ctx, id)
	if err != nil {
		return domain.PullRequest{}, err
	}
	pr.AssignedReviewers = reviewers

	return pr, nil
}

// ReassignReviewer переназначает одного ревьювера на другого из его команды.
func (s *prService) ReassignReviewer(ctx context.Context, prID, oldReviewerID string) (domain.PullRequest, string, error) {
	pr, err := s.prs.GetForUpdate(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, "", err // может быть domain.ErrNotFound
	}

	if err := pr.CanBeReassigned(); err != nil {
		return domain.PullRequest{}, "", err // domain.ErrPRMerged
	}

	currentReviewers, err := s.prs.GetReviewers(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, "", err
	}
	if !contains(currentReviewers, oldReviewerID) {
		return domain.PullRequest{}, "", domain.ErrNotAssigned
	}

	oldReviewer, err := s.users.GetByID(ctx, oldReviewerID)
	if err != nil {
		return domain.PullRequest{}, "", err // может быть domain.ErrNotFound
	}

	members, err := s.listTeamMembers(ctx, oldReviewer.TeamName)
	if err != nil {
		return domain.PullRequest{}, "", err
	}

	candidates := make([]domain.User, 0, len(members))
	for _, u := range members {
		if !u.IsActive {
			continue
		}
		if u.ID == pr.AuthorID {
			continue
		}
		if contains(currentReviewers, u.ID) {
			continue
		}
		candidates = append(candidates, u)
	}

	newReviewerID, err := pickSingleRandomUserID(candidates)
	if err != nil {
		if errors.Is(err, domain.ErrNoCandidate) {
			return domain.PullRequest{}, "", err
		}
		return domain.PullRequest{}, "", err
	}

	if err := s.prs.RemoveReviewer(ctx, prID, oldReviewerID); err != nil {
		return domain.PullRequest{}, "", err
	}
	if err := s.prs.AddReviewer(ctx, prID, newReviewerID); err != nil {
		return domain.PullRequest{}, "", err
	}

	reviewers, err := s.prs.GetReviewers(ctx, prID)
	if err != nil {
		return domain.PullRequest{}, "", err
	}
	pr.AssignedReviewers = reviewers

	return pr, newReviewerID, nil
}

// listTeamMembers — обертка вокруг UserRepository.ListByTeam.
// Для этого UserRepository должен реализовывать метод ListByTeam.
func (s *prService) listTeamMembers(ctx context.Context, teamName string) ([]domain.User, error) {
	type teamLister interface {
		ListByTeam(ctx context.Context, teamName string) ([]domain.User, error)
	}

	lister, ok := s.users.(teamLister)
	if !ok {
		// если почему-то не реализовано — это чисто программерская ошибка
		return nil, errors.New("UserRepository does not implement ListByTeam")
	}

	return lister.ListByTeam(ctx, teamName)
}

func filterActiveExcept(users []domain.User, excludeID string) []domain.User {
	res := make([]domain.User, 0, len(users))
	for _, u := range users {
		if !u.IsActive {
			continue
		}
		if u.ID == excludeID {
			continue
		}
		res = append(res, u)
	}
	return res
}

// GetAssignmentStatsByReviewer возвращает статистику назначений по ревьюверам.
func (s *prService) GetAssignmentStatsByReviewer(ctx context.Context) ([]domain.AssignmentStats, error) {
	return s.prs.GetAssignmentStatsByReviewer(ctx)
}

// GetAssignmentStatsByPR статистика по PR
func (s *prService) GetAssignmentStatsByPR(ctx context.Context) ([]domain.PullRequestAssignmentStats, error) {
	return s.prs.GetAssignmentStatsByPR(ctx)
}

// pickRandomUserIDs выбирает до предела случайных user.ID.
func pickRandomUserIDs(users []domain.User, limit int) []string {
	if limit <= 0 || len(users) == 0 {
		return nil
	}
	if len(users) <= limit {
		out := make([]string, 0, len(users))
		for _, u := range users {
			out = append(out, u.ID)
		}
		return out
	}

	// локальный rand.Rand, чтобы не возникало гонок, каждый раз создаю новый источник
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	idx := r.Perm(len(users))[:limit]

	out := make([]string, 0, limit)
	for _, i := range idx {
		out = append(out, users[i].ID)
	}
	return out
}

// pickSingleRandomUserID возвращает ID одного случайного пользователя или ErrNoCandidate.
func pickSingleRandomUserID(users []domain.User) (string, error) {
	if len(users) == 0 {
		return "", domain.ErrNoCandidate
	}
	if len(users) == 1 {
		return users[0].ID, nil
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	i := r.Intn(len(users))
	return users[i].ID, nil
}

func contains(ids []string, target string) bool {
	for _, id := range ids {
		if id == target {
			return true
		}
	}
	return false
}
