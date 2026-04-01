# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/server ./cmd/server

# Final stage
FROM alpine:3.19

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

COPY --from=builder /app/bin/server ./bin/server
COPY --from=builder /app/migrations  ./migrations

EXPOSE 8080 50051

CMD ["./bin/server"]
