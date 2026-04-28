.PHONY: proto build run-gateway run-catalog run-artist run-progress run-prefs test test-int up down tidy

proto:
	bash scripts/proto-gen.sh

build:
	go build ./...

run-gateway:
	go run ./cmd/gateway

run-catalog:
	go run ./cmd/catalog-svc

run-artist:
	go run ./cmd/artist-svc

run-progress:
	go run ./cmd/progress-svc

run-prefs:
	go run ./cmd/prefs-svc

test:
	go test ./...

test-int:
	go test -tags=integration ./test/...

up:
	docker compose -f deploy/docker-compose.yml up --build

down:
	docker compose -f deploy/docker-compose.yml down -v

tidy:
	go mod tidy
