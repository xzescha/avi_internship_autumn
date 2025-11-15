package service

import (
	"avi_internship_autumn/internal/app"
	"avi_internship_autumn/internal/domain"
	"avi_internship_autumn/internal/repository"
	"context"
)

type userService struct {
	users repository.UserRepository
	prs   repository.PRRepository
}

// NewUserService создаёт сервис для работы с пользователями и их PR.
func NewUserService(
	users repository.UserRepository,
	prs repository.PRRepository,
) app.UserService {
	return &userService{
		users: users,
		prs:   prs,
	}
}

// SetIsActive обновляет флаг активности пользователя.
// Если user_id нет — возвращает domain.ErrNotFound.
func (s *userService) SetIsActive(ctx context.Context, userID string, isActive bool) (domain.User, error) {
	u, err := s.users.UpdateIsActive(ctx, userID, isActive)
	if err != nil {
		return domain.User{}, err
	}
	return u, nil
}

// GetReviewPRs возвращает список PR, где пользователь назначен ревьювером.
// Если юзера нет — domain.ErrNotFound.
func (s *userService) GetReviewPRs(ctx context.Context, userID string) ([]domain.PullRequest, error) {
	_, err := s.users.GetByID(ctx, userID)
	if err != nil {
		return nil, err // тут может быть domain.ErrNotFound
	}

	prs, err := s.prs.ListReviewerPRs(ctx, userID)
	if err != nil {
		return nil, err
	}

	return prs, nil
}

func (s *userService) BulkDeactivateTeam(ctx context.Context, teamName string, userIDs []string) (domain.BulkDeactivateResult, error) {
	result := domain.BulkDeactivateResult{
		TeamName: teamName,
	}

	if len(userIDs) == 0 {
		return result, nil
	}

	members, err := s.users.ListByTeam(ctx, teamName)
	if err != nil {
		return result, err
	}

	deactSetInput := make(map[string]struct{}, len(userIDs))
	for _, id := range userIDs {
		deactSetInput[id] = struct{}{}
	}

	deactivatedInTeam := make([]string, 0, len(userIDs))
	candidatePool := make([]domain.User, 0, len(members))

	for _, u := range members {
		_, toDeactivate := deactSetInput[u.ID]
		if toDeactivate {
			deactivatedInTeam = append(deactivatedInTeam, u.ID)
			continue
		}
		if u.IsActive {
			candidatePool = append(candidatePool, u)
		}
	}

	if len(deactivatedInTeam) == 0 {
		return result, nil
	}

	affectedUsers, err := s.users.BulkDeactivateInTeam(ctx, teamName, deactivatedInTeam)
	if err != nil {
		return result, err
	}
	result.DeactivatedUsers = affectedUsers

	prs, err := s.prs.ListOpenPRsByReviewers(ctx, deactivatedInTeam)
	if err != nil {
		return result, err
	}
	if len(prs) == 0 {
		return result, nil
	}

	deactSet := make(map[string]struct{}, len(deactivatedInTeam))
	for _, id := range deactivatedInTeam {
		deactSet[id] = struct{}{}
	}

	candidateIDs := make([]string, 0, len(candidatePool))
	for _, u := range candidatePool {
		candidateIDs = append(candidateIDs, u.ID)
	}

	for _, pr := range prs {
		current, err := s.prs.GetReviewers(ctx, pr.ID)
		if err != nil {
			return result, err
		}

		if len(current) == 0 {
			continue
		}

		changed := false
		reviewersSet := make(map[string]struct{}, len(current))
		for _, id := range current {
			reviewersSet[id] = struct{}{}
		}

		for _, rid := range current {
			if _, isDeactivated := deactSet[rid]; !isDeactivated {
				continue
			}

			if err := s.prs.RemoveReviewer(ctx, pr.ID, rid); err != nil {
				return result, err
			}
			delete(reviewersSet, rid)
			changed = true

			replacement := pickReplacementForPR(pr.AuthorID, reviewersSet, candidateIDs)
			if replacement == "" {
				continue
			}

			if err := s.prs.AddReviewer(ctx, pr.ID, replacement); err != nil {
				return result, err
			}
			reviewersSet[replacement] = struct{}{}
		}

		if changed {
			result.AffectedPRs++
		}
	}

	return result, nil
}

// pickReplacementForPR выбирает первого подходящего кандидата:
// - не автор PR
// - ещё не в списке ревьюверов
func pickReplacementForPR(authorID string, currentReviewers map[string]struct{}, candidates []string) string {
	for _, id := range candidates {
		if id == authorID {
			continue
		}
		if _, exists := currentReviewers[id]; exists {
			continue
		}
		return id
	}
	return ""
}
