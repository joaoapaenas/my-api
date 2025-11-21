# Variables
DB_URL=sqlite3://dev.db
MIGRATE=migrate -path sql/schema -database "$(DB_URL)"

.PHONY: build run migrate-up migrate-down clean generate test swagger

# Build the binary
build:
	go build -o bin/api.exe cmd/api/main.go

# Run with Air (Live Reload)
run:
	air

# Run standard Go run
run-std:
	go run cmd/api/main.go

# Database Migrations
migrate-up:
	@if not exist dev.db type nul > dev.db
	$(MIGRATE) up

migrate-down:
	$(MIGRATE) down

# Generate SQLC and Swagger code
generate:
	sqlc generate
	swag init -g cmd/api/main.go

# Testing
test:
	go test -v ./...

# Clean artifacts
clean:
	del /q bin\api.exe dev.db
