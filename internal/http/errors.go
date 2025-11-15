// Package http пакет для работы с http запросами к серверу
package http

import (
	"avi_internship_autumn/internal/domain"
	"encoding/json"
	"errors"
	"net/http"
)

// ErrorCode коды из OpenAPI ErrorResponse.error.code
type ErrorCode string

const (
	// CodeTeamExists — команда уже существует.
	CodeTeamExists ErrorCode = "TEAM_EXISTS"
	// CodePRExists - Pull Request уже существует
	CodePRExists ErrorCode = "PR_EXISTS"
	// CodePRMerged - Pull Request уже смержен
	CodePRMerged ErrorCode = "PR_MERGED"
	// CodeNotAssigned - Ревьюер не назначен
	CodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	// CodeNoCandidate - Нет доступного активного кандидата
	CodeNoCandidate ErrorCode = "NO_CANDIDATE"
	// CodeNotFound - Нет такого ресурса
	CodeNotFound ErrorCode = "NOT_FOUND"
)

// структура под ErrorResponse из openapi.yml
type errorBody struct {
	Code    ErrorCode `json:"code"`
	Message string    `json:"message"`
}

// ErrorResponse структура сообщения об ошибке
type ErrorResponse struct {
	Error errorBody `json:"error"`
}

// ErrorHTTP структура сообщения об ошибке с кодом HTTP
type ErrorHTTP struct {
	Status int
	Body   *ErrorResponse
}

// FromDomainError из ошибки домена генерируем ответ
func FromDomainError(err error) *ErrorHTTP {
	switch {
	case errors.Is(err, domain.ErrTeamExists):
		return &ErrorHTTP{
			Status: http.StatusBadRequest, // 400
			Body: &ErrorResponse{
				Error: errorBody{
					Code:    CodeTeamExists,
					Message: "team_name already exists",
				},
			},
		}
	case errors.Is(err, domain.ErrPRExists):
		return &ErrorHTTP{
			Status: http.StatusConflict, // 409
			Body: &ErrorResponse{
				Error: errorBody{
					Code:    CodePRExists,
					Message: "pull_request_id already exists",
				},
			},
		}
	case errors.Is(err, domain.ErrPRMerged):
		return &ErrorHTTP{
			Status: http.StatusConflict, // 409
			Body: &ErrorResponse{
				Error: errorBody{
					Code:    CodePRMerged,
					Message: "cannot reassign on merged PR",
				},
			},
		}
	case errors.Is(err, domain.ErrNotAssigned):
		return &ErrorHTTP{
			Status: http.StatusConflict, // 409
			Body: &ErrorResponse{
				Error: errorBody{
					Code:    CodeNotAssigned,
					Message: "reviewer is not assigned to this PR",
				},
			},
		}
	case errors.Is(err, domain.ErrNoCandidate):
		return &ErrorHTTP{
			Status: http.StatusConflict, // 409
			Body: &ErrorResponse{
				Error: errorBody{
					Code:    CodeNoCandidate,
					Message: "no active replacement candidate in team",
				},
			},
		}
	case errors.Is(err, domain.ErrNotFound):
		return &ErrorHTTP{
			Status: http.StatusNotFound, // 404
			Body: &ErrorResponse{
				Error: errorBody{
					Code:    CodeNotFound,
					Message: "resource not found",
				},
			},
		}
	default:
		// Неописанная ошибка будет возвращать 500 без тела
		return &ErrorHTTP{
			Status: http.StatusInternalServerError,
			Body:   nil,
		}
	}
}

// WriteError утилита для хендлеров
func WriteError(w http.ResponseWriter, err error) {
	httpErr := FromDomainError(err)

	if httpErr.Body == nil {
		w.WriteHeader(httpErr.Status)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpErr.Status)
	_ = json.NewEncoder(w).Encode(httpErr.Body)
}
