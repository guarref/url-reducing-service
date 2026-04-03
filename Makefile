.PHONY: run-memory run-postgres test docker-up docker-down

# запуск с хранилищем в памяти приложения
run-memory:
	STORAGE=memory go run cmd/server/main.go

# запуск с хранилищем в PostgreSQL
run-postgres:
	STORAGE=postgres go run cmd/server/main.go

# запуск сервера с хранилищем в PostgreSQL через docker
docker-up:
	docker compose up --build -d

# остановка контейнеров
docker-down: 
	docker compose down

# запуск unit-тестов
test:
	go test -v ./...