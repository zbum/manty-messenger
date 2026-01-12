# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build dependencies including vips
RUN apk add --no-cache git gcc musl-dev vips-dev

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build binary with CGO enabled for vips
RUN CGO_ENABLED=1 GOOS=linux go build -o main ./cmd/server

# Runtime stage
FROM alpine:3.19

WORKDIR /app

# Install runtime dependencies
RUN apk --no-cache add ca-certificates tzdata vips

# Copy binary from builder
COPY --from=builder /app/main .
COPY --from=builder /app/internal/database/migrations ./internal/database/migrations

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

EXPOSE 8080

CMD ["./main"]
