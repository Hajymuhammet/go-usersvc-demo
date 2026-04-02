# Makefile for go-usersvc-demo

APP_NAME=usersvc-demo
GO=go
GOTEST=$(GO) test
GOBUILD=$(GO) build
GOLINT=golangci-lint
DOCKER_COMPOSE=docker-compose

.PHONY: all fmt vet lint test build run clean deps db-up db-down db-reset proto

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

# Protobuf generate (gRPC, if tooling is set up)
proto:
	protoc --go_out=. --go-grpc_out=. proto/user.proto
