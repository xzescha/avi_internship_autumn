package domain

// AssignmentStats содержит количество назначений ревьюверу.
type AssignmentStats struct {
	ReviewerID string
	Count      int64
}

// PullRequestAssignmentStats — количество назначений по PR.
type PullRequestAssignmentStats struct {
	PullRequestID string
	Count         int64
}
