# Makefile for go-usersvc-demo

APP_NAME=usersvc-demo
GO=go
GOTEST=$(GO) test
GOBUILD=$(GO) build
GOLINT=golangci-lint
DOCKER_COMPOSE=docker-compose

# Load environment variables from .env
ifneq (,$(wildcard .env))
	include .env
	export $(shell sed 's/=.*//' .env)
endif

.PHONY: all fmt vet lint test build run clean deps db-up db-down db-reset migrate-up migrate-down migrate-status proto init-swagger

all: fmt vet lint test build

fmt:
	$(GO) fmt ./...

vet:
	$(GO) vet ./...

lint:
	$(GOLINT) run ./...

test:
	$(GOTEST) ./...

build:
	$(GOBUILD) -o bin/$(APP_NAME) ./cmd/server

run:
	$(GO) run ./cmd/server

clean:
	rm -rf bin

deps:
	$(GO) mod tidy
	$(GO) mod download

# Database helpers (docker-compose)
db-up:
	$(DOCKER_COMPOSE) up -d

db-down:
	$(DOCKER_COMPOSE) down

db-reset: db-down db-up

# Database migrations
migrate-up:
	migrate -path migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" up

migrate-down:
	migrate -path migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" down

migrate-status:
	migrate -path migrations -database "postgresql://$(DB_USER):$(DB_PASSWORD)@$(DB_HOST):$(DB_PORT)/$(DB_NAME)?sslmode=$(DB_SSLMODE)" version

# Protobuf generate (gRPC, if tooling is set up)
proto:
	protoc --go_out=. --go-grpc_out=. proto/user.proto

init-swagger:
	swag init -g cmd/server/main.go -o ./docs