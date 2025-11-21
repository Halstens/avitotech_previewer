# Makefile
.PHONY: build run test migrate-up migrate-down docker-up docker-down clean

build:
	go build -o bin/server ./cmd/server

run: build
	./bin/server

test:
	go test ./... -v

migrate-up:
	migrate -path migrations -database "postgres://postgres:password@localhost:5432/pr_reviewer?sslmode=disable" up

migrate-down:
	migrate -path migrations -database "postgres://postgres:password@localhost:5432/pr_reviewer?sslmode=disable" down

docker-up:
	docker-compose up --build

docker-down:
	docker-compose down

clean:
	rm -rf bin/
	docker-compose down -v

dev: docker-up