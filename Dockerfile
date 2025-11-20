FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o zeno-auth cmd/auth/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl wget

# Install golang-migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /home/appuser

# Copy the binary from builder stage
COPY --from=builder /app/zeno-auth .
COPY --chown=appuser:appuser migrations ./migrations
COPY --chown=appuser:appuser scripts/entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

# Switch to non-root user
USER appuser

# Expose port
EXPOSE 8080

# Run migrations and start app
CMD ["./entrypoint.sh"]