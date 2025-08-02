# Makefile for WordPress Go Next Backend

.PHONY: run test swag docker clean up down build logs restart status server server-bat server-ps help

# ===== Development Commands =====
run:
	go run main.go

server:
	go run main.go server

test:
	go test ./...

swag:
	swag init -g main.go -o docs

# ===== Docker Compose Commands =====
up:
	@echo "=== Starting Backend with PostgreSQL ==="
	docker-compose up -d
	@echo "Backend and PostgreSQL started!"
	@echo "Health check: curl http://localhost:8080/health"
	@echo "PostgreSQL: localhost:5432"

down:
	@echo "=== Stopping Backend and PostgreSQL ==="
	docker-compose down
	@echo "Services stopped!"

build:
	@echo "=== Building Backend Service ==="
	docker-compose build
	@echo "Backend service built!"

logs:
	@echo "=== Showing Logs ==="
	docker-compose logs -f

restart:
	@echo "=== Restarting Services ==="
	docker-compose down
	docker-compose up -d
	@echo "Services restarted!"

status:
	@echo "=== Service Status ==="
	docker-compose ps

# ===== Docker Commands (Legacy) =====
docker:
	docker-compose up --build

# ===== Cleanup Commands =====
clean:
	docker-compose down -v
	rm -rf docs/swagger.json docs/swagger.yaml docs/docs.go

# ===== Windows Batch File Equivalent =====
server-bat:
	@echo "Starting Go Backend Server..."
	@echo "Environment: Development"
	@echo "Port: 8080"
	@echo ""
	GIN_MODE=debug APP_ENV=development go run main.go server

# ===== PowerShell Script Equivalent =====
server-ps:
	@echo "Starting Go Backend Server..."
	@echo "Environment: Development"
	@echo "Port: 8080"
	@echo ""
	@echo "Running: go run main.go server"
	GIN_MODE=debug APP_ENV=development go run main.go server

# ===== Help =====
help:
	@echo "Available commands:"
	@echo ""
	@echo "Development:"
	@echo "  make run          - Run backend (go run main.go)"
	@echo "  make server       - Run backend server (go run main.go server)"
	@echo "  make test         - Run tests"
	@echo "  make swag         - Generate Swagger docs"
	@echo ""
	@echo "Docker Compose:"
	@echo "  make up           - Start backend with PostgreSQL"
	@echo "  make down         - Stop services"
	@echo "  make build        - Build backend service"
	@echo "  make logs         - Show logs"
	@echo "  make restart      - Restart services"
	@echo "  make status       - Show service status"
	@echo ""
	@echo "Legacy:"
	@echo "  make docker       - Start with docker-compose up --build"
	@echo "  make server-bat   - Run server (batch file equivalent)"
	@echo "  make server-ps    - Run server (PowerShell equivalent)"
	@echo ""
	@echo "Cleanup:"
	@echo "  make clean        - Clean up Docker and docs"
	@echo "  make help         - Show this help" 