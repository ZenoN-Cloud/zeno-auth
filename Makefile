.PHONY: build run test clean docker-build migrate-up migrate-down

# Build the application
build:
	go build -o zeno-auth cmd/auth/main.go

# Run the application
run:
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