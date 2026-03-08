.PHONY: proto build test up down lint

proto:
	protoc \
		--proto_path=proto \
		--go_out=internal/identity/port/grpc/pb --go_opt=paths=source_relative \
		--go-grpc_out=internal/identity/port/grpc/pb --go-grpc_opt=paths=source_relative \
		auth/v1/auth.proto

build:
	go build ./...

test:
	go test -v -count=1 ./...

up:
	docker compose -f infra/docker-compose.yml up --build -d

down:
	docker compose -f infra/docker-compose.yml down -v

logs:
	docker compose -f infra/docker-compose.yml logs -f auth-service

lint:
	golangci-lint run ./...
