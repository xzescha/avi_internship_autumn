package app

import (
	"avi_internship_autumn/internal/repository"
	"avi_internship_autumn/internal/repository/pg"
	"database/sql"
)

// Repositories обертка над репозиториями, чтобы иметь возможность передавать единым скопом
type Repositories struct {
	Teams repository.TeamRepository
	Users repository.UserRepository
	PRs   repository.PRRepository
}

// NewRepositories создаёт postgres-реализации всех репозиториев.
func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		Teams: pg.NewTeamRepository(db),
		Users: pg.NewUserRepository(db),
		PRs:   pg.NewPRRepository(db),
	}
}
