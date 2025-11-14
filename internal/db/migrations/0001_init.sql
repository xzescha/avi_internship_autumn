-- Команды
CREATE TABLE teams (
                       team_name TEXT PRIMARY KEY,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- Пользователи
CREATE TABLE users (
                       user_id   TEXT PRIMARY KEY,
                       username  TEXT NOT NULL,
                       team_name TEXT NOT NULL REFERENCES teams(team_name) ON DELETE RESTRICT,
                       is_active BOOLEAN NOT NULL DEFAULT TRUE,
                       created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
                       updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_users_team ON users(team_name);
CREATE INDEX idx_users_team_active ON users(team_name, is_active);

-- Pull Requests
CREATE TABLE pull_requests (
                               pull_request_id   TEXT PRIMARY KEY,
                               pull_request_name TEXT NOT NULL,
                               author_id         TEXT NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,
                               status            TEXT NOT NULL CHECK (status IN ('OPEN', 'MERGED')),
                               created_at        TIMESTAMPTZ NOT NULL DEFAULT now(),
                               merged_at         TIMESTAMPTZ
);

-- Назначенные ревьюверы
CREATE TABLE pr_reviewers (
                              pull_request_id TEXT NOT NULL REFERENCES pull_requests(pull_request_id) ON DELETE CASCADE,
                              reviewer_id     TEXT NOT NULL REFERENCES users(user_id) ON DELETE RESTRICT,
                              PRIMARY KEY (pull_request_id, reviewer_id)
);

CREATE INDEX idx_pr_reviewers_reviewer ON pr_reviewers(reviewer_id);
