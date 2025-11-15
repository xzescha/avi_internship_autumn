// Package app пакет инициализации приложения
package app

import (
	"net/http"
)

// App структура собранного приложения. Хранит интерфейсы сервисов и корневой HTTP-хендлер.
type App struct {
	Handler http.Handler

	TeamService TeamService
	UserService UserService
	PRService   PRService
}

// NewApp обертка в красивую структуру
func NewApp(
	handler http.Handler,
	teamSvc TeamService,
	userSvc UserService,
	prSvc PRService,
) *App {
	return &App{
		Handler:     handler,
		TeamService: teamSvc,
		UserService: userSvc,
		PRService:   prSvc,
	}
}
