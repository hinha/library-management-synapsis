# Stage 1: Base
FROM golang:1.24-alpine AS base

WORKDIR /app
RUN apk add --no-cache git

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Stage 2: Build user-service
FROM base AS user-builder
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/user-service ./cmd/user-service

# Stage 3: Build book-service
FROM base AS book-builder
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/book-service ./cmd/book-service

# Stage 4: Build transaction-service
FROM base AS transaction-builder
RUN CGO_ENABLED=0 GOOS=linux go build -o bin/transaction-service ./cmd/transaction-service

# Stage 5: Final image
FROM alpine:latest

WORKDIR /app

COPY --from=user-builder /app/bin/user-service .
COPY --from=book-builder /app/bin/book-service .
COPY --from=transaction-builder /app/bin/transaction-service .

EXPOSE 50051 6081 50052 6082 50053 6083

CMD ["sh", "-c", "./user-service & ./book-service & ./transaction-service && wait"]