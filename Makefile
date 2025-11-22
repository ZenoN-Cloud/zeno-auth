.PHONY: build run test clean docker-build migrate-up migrate-down dev local-up local-down local-logs local-restart

# Build the application
build:
	go build -o zeno-auth cmd/auth/main.go

# Run the application
run:
	go run cmd/auth/main.go

# Run in development mode with .env
dev:
	@if [ ! -f .env ]; then cp .env.example .env; fi
	go run cmd/auth/main.go

# ───────────────────────────────────────────────────────
# Локальная разработка через Docker Compose
# ───────────────────────────────────────────────────────

# Запустить все сервисы локально
local-up:
	@echo "🚀 Запуск Zeno Auth локально..."
	docker-compose up -d
	@echo "✅ Сервисы запущены!"
	@echo "📍 API:      http://localhost:8080"
	@echo "📍 Health:   http://localhost:8080/health"
	@echo "📍 JWKS:     http://localhost:8080/.well-known/jwks.json"
	@echo "📍 pgAdmin:  http://localhost:5050 (admin@zeno.local / admin)"

# Остановить все сервисы
local-down:
	@echo "🛑 Остановка сервисов..."
	docker-compose down

# Остановить и удалить все данные (volumes)
local-clean:
	@echo "🧹 Очистка всех данных..."
	docker-compose down -v
	rm -rf logs/*

# Показать логи
local-logs:
	docker-compose logs -f

# Показать логи только auth сервиса
local-logs-auth:
	docker-compose logs -f zeno-auth

# Перезапустить сервисы
local-restart:
	@echo "🔄 Перезапуск сервисов..."
	docker-compose restart

# Пересобрать и запустить
local-rebuild:
	@echo "🔨 Пересборка и запуск..."
	docker-compose up -d --build

# Статус сервисов
local-status:
	docker-compose ps

# Тестирование API локально
local-test:
	@echo "🧪 Запуск тестов API..."
	@bash scripts/test-local.sh

# Очистить базу данных
local-db-clean:
	@echo "🧹 Очистка базы данных..."
	docker exec zeno-auth-postgres psql -U zeno_auth -d zeno_auth -c "TRUNCATE TABLE refresh_tokens, org_memberships, organizations, users CASCADE;"
	@echo "✅ База данных очищена!"

# Run tests
test:
	go test -v ./...

# Run unit tests only
test-unit:
	go test -v -short ./...

# Run integration tests
test-integration:
	docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
	docker-compose -f docker-compose.test.yml down -v

# Run E2E tests
test-e2e:
	@if [ -z "$(E2E_BASE_URL)" ]; then echo "E2E_BASE_URL not set"; exit 1; fi
	E2E_BASE_URL=$(E2E_BASE_URL) go test -v ./test/e2e_test.go

# Clean build artifacts
clean:
	rm -f zeno-auth

# Build Docker image
docker-build:
	docker build -t zeno-auth .

# Run migrations up
migrate-up:
	@echo "⬆️  Running migrations up..."
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "❌ DATABASE_URL not set"; \
		exit 1; \
	fi
	migrate -path migrations -database "$(DATABASE_URL)" up
	@echo "✅ Migrations applied!"

# Run migrations down
migrate-down:
	@echo "⬇️  Rolling back migrations..."
	@if [ -z "$(DATABASE_URL)" ]; then \
		echo "❌ DATABASE_URL not set"; \
		exit 1; \
	fi
	migrate -path migrations -database "$(DATABASE_URL)" down
	@echo "✅ Migrations rolled back!"

# Reset migrations (down + up)
migrate-reset:
	@echo "🔄 Resetting database..."
	@$(MAKE) migrate-down
	@$(MAKE) migrate-up
	@echo "✅ Database reset complete!"

# Create new migration
migrate-create:
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Usage: make migrate-create NAME=migration_name"; \
		exit 1; \
	fi
	@echo "🆕 Creating migration: $(NAME)"
	migrate create -ext sql -dir migrations -seq $(NAME)
	@echo "✅ Migration files created!"

# Install dependencies
deps:
	@echo "📦 Installing dependencies..."
	go mod tidy
	go mod download
	@echo "✅ Dependencies installed!"

# Install dev tools
install-tools:
	@echo "🔧 Installing development tools..."
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	@go install mvdan.cc/gofumpt@latest
	@go install honnef.co/go/tools/cmd/staticcheck@latest
	@echo "✅ Tools installed!"

# Generate JWT key pair for development
generate-keys:
	@echo "🔑 Generating JWT key pair..."
	@openssl genrsa -out jwt-private.pem 2048
	@openssl rsa -in jwt-private.pem -pubout -out jwt-public.pem
	@echo "✅ Keys generated:"
	@echo "   - jwt-private.pem (keep secret!)"
	@echo "   - jwt-public.pem"
	@echo ""
	@echo "📝 Next steps:"
	@echo "   1. Copy .env.example to .env.local"
	@echo "   2. Paste keys into .env.local"
	@echo "   3. Never commit .env.local!"

gen-key: generate-keys

# Lint code
lint:
	@echo "🔍 Running linters..."
	go vet ./...
	@echo "✅ Linting passed!"

# Format code
fmt:
	@echo "🎨 Formatting code..."
	go fmt ./...
	@echo "✅ Code formatted!"

# Format with gofumpt (if installed)
fmt-strict:
	@echo "🎨 Formatting code (strict)..."
	@command -v gofumpt >/dev/null 2>&1 && gofumpt -l -w . || go fmt ./...
	@echo "✅ Code formatted!"

# Vet code
vet:
	@echo "🔍 Vetting code..."
	go vet ./...
	@echo "✅ Vet passed!"

# Run staticcheck (if installed)
staticcheck:
	@echo "🔍 Running staticcheck..."
	@command -v staticcheck >/dev/null 2>&1 && staticcheck ./... || echo "⚠️  staticcheck not installed"

# Test coverage
cover:
	@echo "📊 Running tests with coverage..."
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"

# Run all checks
check: fmt vet lint test
	@echo "✅ All checks passed!"