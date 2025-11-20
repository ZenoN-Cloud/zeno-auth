FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy source code first
COPY . .

# Update dependencies and download
RUN go mod tidy && go mod download

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o zeno-auth cmd/auth/main.go

# Final stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates curl wget
WORKDIR /root/

# Install golang-migrate
RUN curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz && \
    mv migrate /usr/local/bin/migrate && \
    chmod +x /usr/local/bin/migrate

# Copy the binary from builder stage
COPY --from=builder /app/zeno-auth .
COPY migrations ./migrations
COPY scripts/entrypoint.sh ./entrypoint.sh
RUN chmod +x ./entrypoint.sh

# Expose port
EXPOSE 8080

# Run migrations and start app
CMD ["./entrypoint.sh"]