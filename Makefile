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