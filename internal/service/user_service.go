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
