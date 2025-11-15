package service

import (
	"avi_internship_autumn/internal/app"
	"avi_internship_autumn/internal/domain"
	"avi_internship_autumn/internal/repository"
	"context"
)

type teamService struct {
	teams repository.TeamRepository
	users repository.UserRepository
}

// NewTeamService создаёт сервис для работы с командами.
func NewTeamService(teams repository.TeamRepository, users repository.UserRepository) app.TeamService {
	return &teamService{
		teams: teams,
		users: users,
	}
}

// CreateTeam создает команду и апсертит всех участников.
// Если команда уже существует — возвращает domain.ErrTeamExists.
func (s *teamService) CreateTeam(ctx context.Context, team domain.Team) (domain.Team, error) {
	exists, err := s.teams.Exists(ctx, team.Name)
	if err != nil {
		return domain.Team{}, err
	}
	if exists {
		return domain.Team{}, domain.ErrTeamExists
	}

	if err := s.teams.Create(ctx, team.Name); err != nil {
		return domain.Team{}, err
	}

	for _, m := range team.Members {
		u := m
		u.TeamName = team.Name
		if err := s.users.Upsert(ctx, u); err != nil {
			return domain.Team{}, err
		}
	}

	created, err := s.teams.Get(ctx, team.Name)
	if err != nil {
		return domain.Team{}, err
	}

	return created, nil
}

// GetTeam возвращает команду по имени или domain.ErrNotFound.
func (s *teamService) GetTeam(ctx context.Context, teamName string) (domain.Team, error) {
	team, err := s.teams.Get(ctx, teamName)
	if err != nil {
		return domain.Team{}, err
	}
	return team, nil
}
