package http

import (
	"avi_internship_autumn/internal/app"
	"net/http"
)

// NewRouter собирает http.Handler со всеми эндпоинтами сервиса.
// На вход сервисы, внутри создаются хендлеры.
func NewRouter(
	teamSvc app.TeamService,
	userSvc app.UserService,
	prSvc app.PRService,
) http.Handler {
	mux := http.NewServeMux()

	teamHandler := NewTeamHandler(teamSvc)
	userHandler := NewUserHandler(userSvc)
	prHandler := NewPRHandler(prSvc)

	// Teams
	mux.HandleFunc("/team/add", teamHandler.AddTeam)
	mux.HandleFunc("/team/get", teamHandler.GetTeam)

	// Users
	mux.HandleFunc("/users/setIsActive", userHandler.SetIsActive)
	mux.HandleFunc("/users/getReview", userHandler.GetReview)

	// PullRequests
	mux.HandleFunc("/pullRequest/create", prHandler.Create)
	mux.HandleFunc("/pullRequest/merge", prHandler.Merge)
	mux.HandleFunc("/pullRequest/reassign", prHandler.Reassign)

	// Statistics
	mux.HandleFunc("/stats/assignments", prHandler.StatsAssignments)

	return mux
}
