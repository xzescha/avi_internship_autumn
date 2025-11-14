package domain

import "errors"

var (
	ErrTeamExists  = errors.New("team exists")
	ErrPRExists    = errors.New("pr exists")
	ErrPRMerged    = errors.New("pr is merged")
	ErrNotAssigned = errors.New("reviewer not assigned")
	ErrNoCandidate = errors.New("no candidate")
	ErrNotFound    = errors.New("not found")
)
