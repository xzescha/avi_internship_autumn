APP_NAME := avi_internship_autumn
DOCKER_COMPOSE := docker compose

.PHONY: build run lint test test-e2e-local up up-tests down k6-only

## Сборка бинарника (локально, без Docker)
build:
	go build -o bin/server ./cmd/server

## Запуск сервера локально (без Docker)
run: build
	HTTP_PORT=8080 \
	DB_HOST=localhost \
	DB_PORT=5432 \
	DB_USER=admin_PR \
	DB_PASSWORD=appPass_QWERTY \
	DB_NAME=PR_serv \
	DB_SSLMODE=disable \
	./bin/server

## Линтер
lint:
	docker run --rm -v $(PWD):/app -w /app golangci/golangci-lint:v2.4.0 \
    		golangci-lint run ./...

## E2E локально (использует testcontainers)
test-e2e-local:
	go test -tags=e2e ./test/e2e -v

## Поднять сервисы (postgres + app) через Docker Compose
up:
	$(DOCKER_COMPOSE) up --build

## Пайплайн для ревью с тестами: postgres + app + e2e + k6 (PROFILE tests)
up-tests:
	$(DOCKER_COMPOSE) --profile tests up --build

## Только k6 (нагрузочное тестирование)
k6-only:
	$(DOCKER_COMPOSE) run --rm k6

## Остановить и удалить образы
down:
	$(DOCKER_COMPOSE) down -v
