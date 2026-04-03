# URL Reducing Service

## Запуск сервиса (с выбором хранилища)

### 1. Локальный запуск в терминале

Для запуска inmemory хранилища:
```bash
STORAGE=memory go run ./cmd/server/main.go
```

Для запуска с PostgreSQL:
1. В начале поднимите контейнер с БД: `docker compose up -d postgres`
2. И запустите сервис:
```bash
STORAGE=postgres go run ./cmd/server/main.go
```
*(При запуске с Postgres сервис автоматически применит миграции базы данных)*.

### 2. Запуск через Docker Compose (решение для продакшена на PostgreSQL)

Для удобства проверки сервис полностью обернут в Docker Compose. По умолчанию собирается и запускается полная связка с PostgreSQL.

```bash
docker compose up --build -d
```
Сервис будет доступен по адресу: `http://localhost:8080`

## Использование API

### Создать короткую ссылку
```bash
curl -X POST http://localhost:8080/api/v1/links/reduce \
  -H "Content-Type: application/json" \
  -d '{"url":"https://ozon.ru/test_request"}'
```

### Перейти по короткой ссылке
```bash
curl -X GET http://localhost:8080/api/v1/links/короткий_код
```
*(Короткий код, полученный в результате POST-ручки должен передаваться напрямую в ссылке)*.