# ============================
#       BUILDER STAGE
# ============================
FROM golang:1.25-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build static binary
RUN CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64 \
    go build -ldflags="-w -s" -o zeno-auth ./cmd/auth

# ============================
#        FINAL STAGE
# ============================
FROM alpine:latest

# Required packages
RUN apk --no-cache add \
    ca-certificates \
    curl \
    wget \
    tzdata \
    libc6-compat

# Install golang-migrate (pinned version)
ENV MIGRATE_VERSION=v4.17.0
RUN wget -qO migrate.tgz "https://github.com/golang-migrate/migrate/releases/download/${MIGRATE_VERSION}/migrate.linux-amd64.tar.gz" \
    && tar -xzf migrate.tgz \
    && mv migrate /usr/local/bin/migrate \
    && chmod +x /usr/local/bin/migrate \
    && rm migrate.tgz

# Create non-root user
RUN addgroup -g 1000 appuser && \
    adduser -D -u 1000 -G appuser appuser

WORKDIR /app

# Copy binary
COPY --from=builder /app/zeno-auth zeno-auth

# Copy migrations + entrypoint
COPY --chown=appuser:appuser migrations ./migrations
COPY --chown=appuser:appuser scripts/entrypoint.sh /usr/local/bin/entrypoint.sh
RUN chmod +x /usr/local/bin/entrypoint.sh

USER appuser

EXPOSE 8080

CMD ["entrypoint.sh"]
