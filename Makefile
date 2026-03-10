APP=booking-api

.PHONY: run test race up down migrate-up migrate-down build

run:
	go run ./cmd/api

build:
	go build -o bin/$(APP) ./cmd/api

test:
	go test ./...

race:
	go test -race ./...

up:
	docker compose up -d postgres zookeeper kafka

down:
	docker compose down -v

migrate-up:
	docker run --rm --network host -v $(PWD)/migrations:/migrations migrate/migrate:4 \
		-path=/migrations -database "postgres://booking:booking@localhost:5432/booking?sslmode=disable" up

migrate-down:
	docker run --rm --network host -v $(PWD)/migrations:/migrations migrate/migrate:4 \
		-path=/migrations -database "postgres://booking:booking@localhost:5432/booking?sslmode=disable" down 1
