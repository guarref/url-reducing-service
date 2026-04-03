FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /url-reducing-service ./cmd/server

FROM alpine:3.20

RUN apk add --no-cache ca-certificates

WORKDIR /

COPY --from=builder /url-reducing-service /url-reducing-service
COPY --from=builder /app/migrations /migrations

ENV SERVICE_PORT=8080
ENV STORAGE=postgres
ENV MIGRATE_FOLDER=/migrations
ENV BASE_URL=http://localhost:8080
ENV SERVICE_ENV=development

EXPOSE 8080

ENTRYPOINT ["/url-reducing-service"]