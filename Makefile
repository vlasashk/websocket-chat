test_server: up_env
	go test -cover ./internal/integration_test -coverpkg=./internal/server/...,./internal/storage/... -coverprofile cover.out && go tool cover -html=cover.out -o cover.html
	docker compose -f docker-compose.yaml down

up_env:
	docker compose -f docker-compose.yaml up --scale server=0 --scale storage=0 -d

run_server:
	docker compose -f docker-compose.yaml up -d

run_client:
	go run cmd/client/main.go