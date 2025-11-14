package domain

// ActiveMembers возвращает всех активных участников команды.
func (t Team) ActiveMembers() []User {
	res := make([]User, 0, len(t.Members))
	for _, u := range t.Members {
		if u.IsActive {
			res = append(res, u)
		}
	}
	return res
}

// ActiveMembersExcept возвращает активных участников команды,
// исключая пользователей с указанными ID (например, автора PR или текущих ревьюверов).
func (t Team) ActiveMembersExcept(excludedIDs ...string) []User {
	if len(excludedIDs) == 0 {
		return t.ActiveMembers()
	}

	excluded := make(map[string]struct{}, len(excludedIDs))
	for _, id := range excludedIDs {
		excluded[id] = struct{}{}
	}

	res := make([]User, 0, len(t.Members))
	for _, u := range t.Members {
		if !u.IsActive {
			continue
		}
		if _, skip := excluded[u.ID]; skip {
			continue
		}
		res = append(res, u)
	}
	return res
}

// IsMerged показывает, что PR уже в статусе MERGED.
func (pr PullRequest) IsMerged() bool {
	return pr.Status == PRStatusMerged
}

// CanBeReassigned проверяет, можно ли менять ревьюверов для этого PR.
// Если нельзя — возвращает доменную ошибку (например, ErrPRMerged).
func (pr PullRequest) CanBeReassigned() error {
	if pr.IsMerged() {
		return ErrPRMerged
	}
	return nil
}
