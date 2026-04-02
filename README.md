# Go User Service Demo

A production-ready user management service built with Go. Supports both REST and gRPC with PostgreSQL persistence, Redis caching, and Docker Compose for containerized deployment.

## 🚀 Features

- REST API (HTTP)
- gRPC API (Protocol Buffers)
- PostgreSQL persistence
- Redis caching for user reads
- Docker Compose for dev stack orchestration
- Clean architecture: domain, service, infrastructure, transport
- Validations using `github.com/go-playground/validator/v10`
- Consistent error handling and API responses

## 🧩 Prerequisites

- Go 1.25.0
- Docker
- Docker Compose

## 📁 Project Structure

```
go-usersvc-demo/
├── cmd/
│   ├── api/          # REST API entrypoint
│   ├── grpc/         # gRPC API entrypoint
│   └── server/       # shared initialization
├── internal/
│   ├── config/       # environment + config parsing
│   ├── domain/       # entity definitions
│   ├── infrastructure/
│   │   ├── postgres/ # Postgres repository
│   │   └── redis/    # Redis cache layer
│   ├── service/      # business logic
│   └── transport/    # HTTP and gRPC handlers
├── pkg/              # generated protobufs + shared types
│   └── pb/
├── proto/            # .proto definitions
├── migrations/       # DB migration files
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── README.md
```

## 🏁 Getting Started

### 1) Environment config

Set optional env vars in `.env`, or rely on defaults from `docker-compose.yml`:

- `DB_HOST` (default: `db`)
- `DB_PORT` (default: `5432`)
- `DB_USER` (default: `postgres`)
- `DB_PASSWORD` (default: `postgres`)
- `DB_NAME` (default: `usersdb`)
- `REDIS_ADDR` (default: `redis:6379`)
- `HTTP_PORT` (default: `8080`)
- `GRPC_PORT` (default: `50051`)

### 2) Start dependencies

```bash
docker-compose up -d
```

Wait for the database and redis to be ready.

### 3) (Optional) Apply migrations manually

```bash
go run github.com/golang-migrate/migrate/v4/cmd/migrate \
  -path migrations -database "postgres://postgres:postgres@localhost:5432/usersdb?sslmode=disable" up
```

### 4) Install Go modules

```bash
go mod tidy
```

### 5) Run REST server

```bash
go run cmd/api/main.go
```

### 6) Run gRPC server

```bash
go run cmd/grpc/main.go
```

## 🛠️ API Reference

### REST

- `POST /users` - create user
- `GET /users/:id` - get user by ID
- `GET /users?page=<n>&limit=<n>` - list users
- `PUT /users/:id` - update user
- `DELETE /users/:id` - delete user

Example create payload:

```json
{
  "name": "Jane Doe",
  "email": "jane@example.com",
  "password": "secret123"
}
```

Example response:

```json
{
  "id": "1",
  "name": "Jane Doe",
  "email": "jane@example.com",
  "created_at": "2026-03-31T12:00:00Z",
  "updated_at": "2026-03-31T12:00:00Z"
}
```

### gRPC

Generated protobufs (in `pkg/pb`):

- `CreateUser`, `GetUserByID`, `ListUsers`, `UpdateUser`, `DeleteUser`

Protofile is at `proto/user.proto`.

## 🧪 Testing

Unit tests:

```bash
go test ./internal/... ./... -v
```

## 🪄 Build

Docker image build:

```bash
docker build -t go-usersvc-demo .
```

## 🛑 Stop and clean

```bash
docker-compose down
```

With volumes:

```bash
docker-compose down -v
```

## 📦 Feedback

If you want CLI scripts (`make`, additional integration tests, auth), open a GitHub issue or PR.

## 📜 License

MIT License

