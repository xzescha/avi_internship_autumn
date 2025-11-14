package http

import (
	"avi_internship_autumn/internal/app"
	"avi_internship_autumn/internal/domain"
	"encoding/json"
	"net/http"
	"time"
)

type teamMemberDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type teamDTO struct {
	TeamName string          `json:"team_name"`
	Members  []teamMemberDTO `json:"members"`
}

type userDTO struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type pullRequestDTO struct {
	PullRequestID     string     `json:"pull_request_id"`
	PullRequestName   string     `json:"pull_request_name"`
	AuthorID          string     `json:"author_id"`
	Status            string     `json:"status"`
	AssignedReviewers []string   `json:"assigned_reviewers"`
	CreatedAt         *time.Time `json:"createdAt,omitempty"`
	MergedAt          *time.Time `json:"mergedAt,omitempty"`
}

type pullRequestShortDTO struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
}

func teamToDTO(t domain.Team) teamDTO {
	members := make([]teamMemberDTO, 0, len(t.Members))
	for _, m := range t.Members {
		members = append(members, teamMemberDTO{
			UserID:   m.ID,
			Username: m.Username,
			IsActive: m.IsActive,
		})
	}
	return teamDTO{
		TeamName: t.Name,
		Members:  members,
	}
}

func pullRequestToDTO(pr domain.PullRequest) pullRequestDTO {
	dto := pullRequestDTO{
		PullRequestID:     pr.ID,
		PullRequestName:   pr.Name,
		AuthorID:          pr.AuthorID,
		Status:            string(pr.Status),
		AssignedReviewers: pr.AssignedReviewers,
	}

	if !pr.CreatedAt.IsZero() {
		t := pr.CreatedAt
		dto.CreatedAt = &t
	}
	if pr.MergedAt != nil && !pr.MergedAt.IsZero() {
		dto.MergedAt = pr.MergedAt
	}

	return dto
}

func pullRequestToShortDTO(pr domain.PullRequest) pullRequestShortDTO {
	return pullRequestShortDTO{
		PullRequestID:   pr.ID,
		PullRequestName: pr.Name,
		AuthorID:        pr.AuthorID,
		Status:          string(pr.Status),
	}
}

type TeamHandler struct {
	svc app.TeamService
}

func NewTeamHandler(svc app.TeamService) *TeamHandler {
	return &TeamHandler{svc: svc}
}

// AddTeam POST /team/add
func (h *TeamHandler) AddTeam(w http.ResponseWriter, r *http.Request) {
	var req struct {
		TeamName string          `json:"team_name"`
		Members  []teamMemberDTO `json:"members"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	team := domain.Team{
		Name:    req.TeamName,
		Members: make([]domain.User, 0, len(req.Members)),
	}

	for _, m := range req.Members {
		team.Members = append(team.Members, domain.User{
			ID:       m.UserID,
			Username: m.Username,
			TeamName: req.TeamName,
			IsActive: m.IsActive,
		})
	}

	created, err := h.svc.CreateTeam(r.Context(), team)
	if err != nil {
		WriteError(w, err)
		return
	}

	resp := struct {
		Team teamDTO `json:"team"`
	}{
		Team: teamToDTO(created),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// GetTeam GET /team/get?team_name=...
func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	team, err := h.svc.GetTeam(r.Context(), teamName)
	if err != nil {
		WriteError(w, err)
		return
	}

	resp := teamToDTO(team)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

type UserHandler struct {
	svc app.UserService
}

func NewUserHandler(svc app.UserService) *UserHandler {
	return &UserHandler{svc: svc}
}

// SetIsActive POST /users/setIsActive
func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req struct {
		UserID   string `json:"user_id"`
		IsActive bool   `json:"is_active"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	user, err := h.svc.SetIsActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		WriteError(w, err)
		return
	}

	resp := struct {
		User userDTO `json:"user"`
	}{
		User: userDTO{
			UserID:   user.ID,
			Username: user.Username,
			TeamName: user.TeamName,
			IsActive: user.IsActive,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// GetReview GET /users/getReview?user_id=...
func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	prs, err := h.svc.GetReviewPRs(r.Context(), userID)
	if err != nil {
		WriteError(w, err)
		return
	}

	resp := struct {
		UserID       string                `json:"user_id"`
		PullRequests []pullRequestShortDTO `json:"pull_requests"`
	}{
		UserID:       userID,
		PullRequests: make([]pullRequestShortDTO, 0, len(prs)),
	}

	for _, pr := range prs {
		resp.PullRequests = append(resp.PullRequests, pullRequestToShortDTO(pr))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// ===== PullRequests =====

type PRHandler struct {
	svc app.PRService
}

func NewPRHandler(svc app.PRService) *PRHandler {
	return &PRHandler{svc: svc}
}

// Create POST /pullRequest/create
func (h *PRHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID   string `json:"pull_request_id"`
		PullRequestName string `json:"pull_request_name"`
		AuthorID        string `json:"author_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pr, err := h.svc.CreatePR(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		WriteError(w, err)
		return
	}

	resp := struct {
		PR pullRequestDTO `json:"pr"`
	}{
		PR: pullRequestToDTO(pr),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(resp)
}

// Merge POST /pullRequest/merge
func (h *PRHandler) Merge(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pr, err := h.svc.MergePR(r.Context(), req.PullRequestID)
	if err != nil {
		WriteError(w, err)
		return
	}

	resp := struct {
		PR pullRequestDTO `json:"pr"`
	}{
		PR: pullRequestToDTO(pr),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}

// Reassign POST /pullRequest/reassign
func (h *PRHandler) Reassign(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PullRequestID string `json:"pull_request_id"`
		OldUserID     string `json:"old_user_id"`
		// на всякий случай поддержим пример из спецификации
		OldReviewerID string `json:"old_reviewer_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	oldID := req.OldUserID
	if oldID == "" {
		oldID = req.OldReviewerID
	}
	if req.PullRequestID == "" || oldID == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	pr, replacedBy, err := h.svc.ReassignReviewer(r.Context(), req.PullRequestID, oldID)
	if err != nil {
		WriteError(w, err)
		return
	}

	resp := struct {
		PR         pullRequestDTO `json:"pr"`
		ReplacedBy string         `json:"replaced_by"`
	}{
		PR:         pullRequestToDTO(pr),
		ReplacedBy: replacedBy,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)
}
