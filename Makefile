.PHONY: build run test clean docker-build migrate-up migrate-down dev

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
	migrate -path migrations -database "$(DATABASE_URL)" up

# Run migrations down
migrate-down:
	migrate -path migrations -database "$(DATABASE_URL)" down

# Install dependencies
deps:
	go mod tidy
	go mod download

# Generate JWT private key for development
gen-key:
	@openssl genrsa -out jwt-private.pem 2048
	@echo "JWT private key generated: jwt-private.pem"