package pg

import (
	"avi_internship_autumn/internal/domain"
	"avi_internship_autumn/internal/repository"
	"context"
	"database/sql"
	"errors"
)

type userRepo struct {
	db *sql.DB
}

// NewUserRepository создаёт репозиторий пользователей на базе PostgreSQL.
func NewUserRepository(db *sql.DB) repository.UserRepository {
	return &userRepo{db: db}
}

// Upsert создаёт или обновляет пользователя.
// Если user_id уже есть — обновляем username, team_name и is_active.
func (r *userRepo) Upsert(ctx context.Context, u domain.User) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO users (user_id, username, team_name, is_active)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (user_id) DO UPDATE
        SET username = EXCLUDED.username,
            team_name = EXCLUDED.team_name,
            is_active = EXCLUDED.is_active,
            updated_at = now()
    `, u.ID, u.Username, u.TeamName, u.IsActive)
	return err
}

// GetByID возвращает пользователя по id или domain.ErrNotFound.
func (r *userRepo) GetByID(ctx context.Context, id string) (domain.User, error) {
	var u domain.User
	err := r.db.QueryRowContext(ctx, `
        SELECT user_id, username, team_name, is_active
        FROM users
        WHERE user_id = $1
    `, id).Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	return u, nil
}

// ListByTeam возвращает всех пользователей команды.
func (r *userRepo) ListByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT user_id, username, team_name, is_active
        FROM users
        WHERE team_name = $1
        ORDER BY user_id
    `, teamName)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			return
		}
	}(rows)

	users := make([]domain.User, 0)
	for rows.Next() {
		var u domain.User
		if err := rows.Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive); err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

// UpdateIsActive обновляет флаг активности и возвращает обновлённого пользователя.
// Если user_id нет — domain.ErrNotFound.
func (r *userRepo) UpdateIsActive(ctx context.Context, id string, isActive bool) (domain.User, error) {
	var u domain.User
	err := r.db.QueryRowContext(ctx, `
        UPDATE users
        SET is_active = $2,
            updated_at = now()
        WHERE user_id = $1
        RETURNING user_id, username, team_name, is_active
    `, id, isActive).Scan(&u.ID, &u.Username, &u.TeamName, &u.IsActive)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.User{}, domain.ErrNotFound
		}
		return domain.User{}, err
	}
	return u, nil
}
