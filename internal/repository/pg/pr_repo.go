// Package pg содержит реализации репозиториев, работающих с PostgreSQL.
package pg

import (
	"avi_internship_autumn/internal/domain"
	"avi_internship_autumn/internal/repository"
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
)

type prRepo struct {
	db *sql.DB
}

type prRowScanner interface {
	Scan(dest ...any) error
}

func scanPullRequest(s prRowScanner) (domain.PullRequest, error) {
	var pr domain.PullRequest
	var statusStr string
	var createdAt sql.NullTime
	var mergedAt sql.NullTime

	if err := s.Scan(
		&pr.ID,
		&pr.Name,
		&pr.AuthorID,
		&statusStr,
		&createdAt,
		&mergedAt,
	); err != nil {
		return domain.PullRequest{}, err
	}

	pr.Status = domain.PRStatus(statusStr)
	if createdAt.Valid {
		pr.CreatedAt = createdAt.Time
	}
	if mergedAt.Valid {
		t := mergedAt.Time
		pr.MergedAt = &t
	}

	return pr, nil
}

// Вспомогательная функция для сканирования статистики вида (id, count)
func scanStats[T any](rows *sql.Rows, mapper func(id string, count int64) T) ([]T, error) {
	defer func() {
		_ = rows.Close()
	}()

	var result []T

	for rows.Next() {
		var id string
		var cnt int64

		if err := rows.Scan(&id, &cnt); err != nil {
			return nil, err
		}

		result = append(result, mapper(id, cnt))
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return result, nil
}

// NewPRRepository создаёт репозиторий pull requestов на базе PostgreSQL.
func NewPRRepository(db *sql.DB) repository.PRRepository {
	return &prRepo{db: db}
}

// Exists проверяет, есть ли PR с таким id.
func (r *prRepo) Exists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := r.db.QueryRowContext(ctx, `
        SELECT EXISTS (
            SELECT 1 FROM pull_requests WHERE pull_request_id = $1
        )
    `, id).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}

// Create создаёт запись о PR (без ревьюверов — они добавляются отдельно через AddReviewer).
func (r *prRepo) Create(ctx context.Context, pr domain.PullRequest) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status)
        VALUES ($1, $2, $3, $4)
    `, pr.ID, pr.Name, pr.AuthorID, string(pr.Status))
	return err
}

// GetForUpdate возвращает PR по id.
// Имя намекает на блокировку, но блокировка будет работать только,
// если вызывающий оборачивает это в транзакцию.
func (r *prRepo) GetForUpdate(ctx context.Context, id string) (domain.PullRequest, error) {
	row := r.db.QueryRowContext(ctx, `
        SELECT pull_request_id,
               pull_request_name,
               author_id,
               status,
               created_at,
               merged_at
        FROM pull_requests
        WHERE pull_request_id = $1
    `, id)

	pr, err := scanPullRequest(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.PullRequest{}, domain.ErrNotFound
		}
		return domain.PullRequest{}, err
	}

	return pr, nil
}

// UpdateStatusMerged ставит PR в статус MERGED и проставляет merged_at (если ещё не стоял).
func (r *prRepo) UpdateStatusMerged(ctx context.Context, id string) error {
	res, err := r.db.ExecContext(ctx, `
        UPDATE pull_requests
        SET status   = 'MERGED',
            merged_at = COALESCE(merged_at, now())
        WHERE pull_request_id = $1
    `, id)
	if err != nil {
		return err
	}

	affected, err := res.RowsAffected()
	if err == nil && affected == 0 {
		return domain.ErrNotFound
	}
	return err
}

// ListReviewerPRs возвращает список PR, где пользователь назначен ревьювером.
func (r *prRepo) ListReviewerPRs(ctx context.Context, reviewerID string) ([]domain.PullRequest, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT p.pull_request_id,
               p.pull_request_name,
               p.author_id,
               p.status,
               p.created_at,
               p.merged_at
        FROM pull_requests p
        JOIN pr_reviewers r ON r.pull_request_id = p.pull_request_id
        WHERE r.reviewer_id = $1
        ORDER BY p.created_at DESC, p.pull_request_id
    `, reviewerID)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			return
		}
	}(rows)

	var prs []domain.PullRequest

	for rows.Next() {
		var pr domain.PullRequest
		var statusStr string
		var createdAt sql.NullTime
		var mergedAt sql.NullTime

		if err := rows.Scan(
			&pr.ID,
			&pr.Name,
			&pr.AuthorID,
			&statusStr,
			&createdAt,
			&mergedAt,
		); err != nil {
			return nil, err
		}

		pr.Status = domain.PRStatus(statusStr)
		if createdAt.Valid {
			pr.CreatedAt = createdAt.Time
		}
		if mergedAt.Valid {
			t := mergedAt.Time
			pr.MergedAt = &t
		}
		// AssignedReviewers тут не подтягиваем — для /users/getReview это не нужно
		prs = append(prs, pr)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prs, nil
}

// GetReviewers возвращает список reviewer_id для PR.
func (r *prRepo) GetReviewers(ctx context.Context, prID string) ([]string, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT reviewer_id
        FROM pr_reviewers
        WHERE pull_request_id = $1
        ORDER BY reviewer_id
    `, prID)
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			return
		}
	}(rows)

	var reviewers []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, id)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reviewers, nil
}

// AddReviewer добавляет связку PR–reviewer.
func (r *prRepo) AddReviewer(ctx context.Context, prID, reviewerID string) error {
	_, err := r.db.ExecContext(ctx, `
        INSERT INTO pr_reviewers (pull_request_id, reviewer_id)
        VALUES ($1, $2)
        ON CONFLICT DO NOTHING
    `, prID, reviewerID)
	return err
}

// RemoveReviewer удаляет ревьювера у PR.
func (r *prRepo) RemoveReviewer(ctx context.Context, prID, reviewerID string) error {
	res, err := r.db.ExecContext(ctx, `
        DELETE FROM pr_reviewers
        WHERE pull_request_id = $1 AND reviewer_id = $2
    `, prID, reviewerID)
	if err != nil {
		return err
	}

	// Если вдруг ни одной строки не затронули — формально можно вернуть ErrNotFound,
	// но сервис перед этим уже проверяет назначение, так что это скорее аномалия.
	_, _ = res.RowsAffected()
	return nil
}

// GetAssignmentStatsByReviewer возвращает число назначений по каждому ревьюверу.
func (r *prRepo) GetAssignmentStatsByReviewer(ctx context.Context) ([]domain.AssignmentStats, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT reviewer_id, COUNT(*) AS cnt
        FROM pr_reviewers
        GROUP BY reviewer_id
        ORDER BY cnt DESC, reviewer_id
    `)
	if err != nil {
		return nil, err
	}
	return scanStats(rows, func(id string, count int64) domain.AssignmentStats {
		return domain.AssignmentStats{
			ReviewerID: id,
			Count:      count,
		}
	})
}

// GetAssignmentStatsByPR возвращает число назначений по каждому PR.
func (r *prRepo) GetAssignmentStatsByPR(ctx context.Context) ([]domain.PullRequestAssignmentStats, error) {
	rows, err := r.db.QueryContext(ctx, `
        SELECT pull_request_id, COUNT(*) AS cnt
        FROM pr_reviewers
        GROUP BY pull_request_id
        ORDER BY cnt DESC, pull_request_id
    `)
	if err != nil {
		return nil, err
	}

	return scanStats(rows, func(id string, count int64) domain.PullRequestAssignmentStats {
		return domain.PullRequestAssignmentStats{
			PullRequestID: id,
			Count:         count,
		}
	})
}

// ListOpenPRsByReviewers возвращает все открытые PR, где назначен кто-то из reviewerIDs.
func (r *prRepo) ListOpenPRsByReviewers(ctx context.Context, reviewerIDs []string) ([]domain.PullRequest, error) {
	if len(reviewerIDs) == 0 {
		return nil, nil
	}

	rows, err := r.db.QueryContext(ctx, `
        SELECT DISTINCT
               p.pull_request_id,
               p.pull_request_name,
               p.author_id,
               p.status,
               p.created_at,
               p.merged_at
        FROM pull_requests p
        JOIN pr_reviewers r ON r.pull_request_id = p.pull_request_id
        WHERE p.status = 'OPEN'
          AND r.reviewer_id = ANY($1)
        ORDER BY p.created_at DESC, p.pull_request_id
    `, pq.Array(reviewerIDs))
	if err != nil {
		return nil, err
	}
	defer func(rows *sql.Rows) {
		err := rows.Close()
		if err != nil {
			return
		}
	}(rows)

	var prs []domain.PullRequest
	for rows.Next() {
		pr, err := scanPullRequest(rows)
		if err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return prs, nil
}
