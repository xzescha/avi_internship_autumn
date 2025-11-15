// Package domain содержит бизнес-модели и доменные ошибки приложения.
package domain

import "errors"

var (
	// ErrTeamExists команда существует
	ErrTeamExists = errors.New("team exists")
	// ErrPRExists Pull Request существует
	ErrPRExists = errors.New("pr exists")
	// ErrPRMerged Pull Request смержили
	ErrPRMerged = errors.New("pr is merged")
	// ErrNotAssigned ревьюер не назначен
	ErrNotAssigned = errors.New("reviewer not assigned")
	// ErrNoCandidate нет свободных кандидатов в ревьюеры
	ErrNoCandidate = errors.New("no candidate")
	// ErrNotFound ресурс не найден (общая ошибка относительно)
	ErrNotFound = errors.New("not found")
)
