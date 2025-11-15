package pg

import (
	"avi_internship_autumn/internal/domain"
	"avi_internship_autumn/internal/repository"
	"context"
	"database/sql"
	"errors"
)

type teamRepo struct {
	db *sql.DB
}

// NewTeamRepository возвращает postgres-реализацию TeamRepository.
func NewTeamRepository(db *sql.DB) repository.TeamRepository {
	return &teamRepo{db: db}
}

// Create вставляет новую команду в таблицу teams.
func (r *teamRepo) Create(ctx context.Context, teamName string) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO teams (team_name)
        VALUES ($1)
    `, teamName)
	return err
}

// Exists проверяет, есть ли команда с таким именем.
func (r *teamRepo) Exists(ctx context.Context, teamName string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM teams WHERE team_name = $1
        )
    `, teamName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Get возвращает команду и всех её участников.
// Если команды нет — domain.ErrNotFound.
func (r *teamRepo) Get(ctx context.Context, teamName string) (domain.Team, error) {
	// Сначала убеждаемся, что команда существует
	var name string
	err := r.db.QueryRowContext(ctx, `
        SELECT team_name
        FROM teams
        WHERE team_name = $1
    `, teamName).Scan(&name)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Team{}, domain.ErrNotFound
		}
		return domain.Team{}, err
	}

	// Забираем всех юзеров этой команды
	rows, err := r.db.QueryContext(ctx, `
        SELECT user_id, username, is_active
        FROM users
        WHERE team_name = $1
        ORDER BY user_id
    `, teamName)
	if err != nil {
		return domain.Team{}, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			return
		}
	}(rows)

	members := make([]domain.User, 0)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.IsActive); err != nil {
			return domain.Team{}, err
		}
		u.TeamName = teamName
		members = append(members, u)
	}
	if err := rows.Err(); err != nil {
		return domain.Team{}, err
	}

	return domain.Team{
		Name:    teamName,
		Members: members,
	}, nil
}
