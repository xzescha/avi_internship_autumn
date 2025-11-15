//go:build e2e

package e2e

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	_ "github.com/lib/pq"
	tc "github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"

	"avi_internship_autumn/internal/app"
	apihttp "avi_internship_autumn/internal/http"
	"avi_internship_autumn/internal/service"
)

func startPostgres(t *testing.T, ctx context.Context) (*sql.DB, func()) {
	t.Helper()

	// 1. Если есть E2E_DSN — используем его (режим docker-compose)
	if dsn := os.Getenv("E2E_DSN"); dsn != "" {
		db, err := sql.Open("postgres", dsn)
		if err != nil {
			t.Fatalf("failed to open db with E2E_DSN: %v", err)
		}

		ctxPing, cancel := context.WithTimeout(ctx, 10*time.Second)
		defer cancel()
		if err := db.PingContext(ctxPing); err != nil {
			t.Fatalf("failed to ping db with E2E_DSN: %v", err)
		}

		teardown := func() {
			_ = db.Close()
		}
		return db, teardown
	}

	// 2. Иначе — локальный режим: поднимаем Postgres через testcontainers
	req := tc.ContainerRequest{
		Image:        "postgres:16-alpine",
		ExposedPorts: []string{"5432/tcp"},
		Env: map[string]string{
			"POSTGRES_USER":     "test",
			"POSTGRES_PASSWORD": "test",
			"POSTGRES_DB":       "testdb",
		},
		WaitingFor: wait.ForListeningPort("5432/tcp").
			WithStartupTimeout(30 * time.Second),
	}

	container, err := tc.GenericContainer(ctx, tc.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("failed to start postgres container: %v", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		t.Fatalf("failed to get container host: %v", err)
	}
	mappedPort, err := container.MappedPort(ctx, "5432")
	if err != nil {
		t.Fatalf("failed to get mapped port: %v", err)
	}

	dsn := "postgres://test:test@" + host + ":" + mappedPort.Port() + "/testdb?sslmode=disable"

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}

	ctxPing, cancelPing := context.WithTimeout(ctx, 10*time.Second)
	defer cancelPing()
	if err := db.PingContext(ctxPing); err != nil {
		t.Fatalf("failed to ping db: %v", err)
	}

	teardown := func() {
		_ = db.Close()
		_ = container.Terminate(context.Background())
	}

	return db, teardown
}

func applyMigrations(t *testing.T, db *sql.DB) {
	t.Helper()

	dir := "internal/db/migrations"
	entries, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		t.Fatalf("failed to list migrations: %v", err)
	}
	sort.Strings(entries)

	for _, path := range entries {
		content, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read migration %s: %v", path, err)
		}
		if _, err := db.Exec(string(content)); err != nil {
			t.Fatalf("failed to exec migration %s: %v", path, err)
		}
	}
}

func TestE2E_FullFlow(t *testing.T) {
	ctx := context.Background()

	db, teardown := startPostgres(t, ctx)
	defer teardown()

	// Если E2E_DSN не задан, значит мы сами подняли тестовый Postgres через testcontainers
	// и должны применить миграции. В docker-compose база уже проинициализирована entrypoint'ом.
	if os.Getenv("E2E_DSN") == "" {
		applyMigrations(t, db)
	}

	// репозитории
	repos := app.NewRepositories(db)

	// сервисы
	teamSvc := service.NewTeamService(repos.Teams, repos.Users)
	userSvc := service.NewUserService(repos.Users, repos.PRs)
	prSvc := service.NewPRService(repos.PRs, repos.Users)

	// HTTP
	handler := apihttp.NewRouter(teamSvc, userSvc, prSvc)
	server := httptest.NewServer(handler)
	defer server.Close()

	client := server.Client()

	// 1) создаём команду
	createTeamReq := `{
	  "team_name": "payments_e2e",
	  "members": [
	    { "user_id": "u1", "username": "Alice", "is_active": true },
	    { "user_id": "u2", "username": "Bob",   "is_active": true },
	    { "user_id": "u3", "username": "Charlie", "is_active": true }
	  ]
	}`

	resp, err := client.Post(server.URL+"/team/add", "application/json", strings.NewReader(createTeamReq))
	if err != nil {
		t.Fatalf("team/add request failed: %v", err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusBadRequest {
		body, _ := io.ReadAll(resp.Body)
		t.Fatalf("unexpected status %d, body: %s", resp.StatusCode, string(body))
	}

	// 2) создаём PR
	createPRReq := `{
	  "pull_request_id": "pr-e2e-1",
	  "pull_request_name": "Add payments endpoint",
	  "author_id": "u1"
	}`

	resp, err = client.Post(server.URL+"/pullRequest/create", "application/json", strings.NewReader(createPRReq))
	if err != nil {
		t.Fatalf("pullRequest/create request failed: %v", err)
	}
	bodyBytes, _ := io.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		return
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		t.Fatalf("unexpected status %d for create PR, body: %s", resp.StatusCode, string(bodyBytes))
	}

	type prResponse struct {
		PR struct {
			ID                string   `json:"pull_request_id"`
			AssignedReviewers []string `json:"assigned_reviewers"`
		} `json:"pr"`
	}

	var prResp prResponse
	_ = json.Unmarshal(bodyBytes, &prResp)

	// 3) bulkDeactivate u2, u3
	bulkReq := `{
	  "team_name": "payments_e2e",
	  "user_ids": ["u2", "u3"]
	}`
	resp, err = client.Post(server.URL+"/users/bulkDeactivate", "application/json", strings.NewReader(bulkReq))
	if err != nil {
		t.Fatalf("bulkDeactivate request failed: %v", err)
	}
	bodyBytes, _ = io.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d for bulkDeactivate, body: %s", resp.StatusCode, string(bodyBytes))
	}

	// 4) проверяем статистику
	resp, err = client.Get(server.URL + "/stats/assignments")
	if err != nil {
		t.Fatalf("stats/assignments request failed: %v", err)
	}
	bodyBytes, _ = io.ReadAll(resp.Body)
	err = resp.Body.Close()
	if err != nil {
		return
	}
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("unexpected status %d for stats, body: %s", resp.StatusCode, string(bodyBytes))
	}
}
