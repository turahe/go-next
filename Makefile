# Go Next - Backend Optimization Makefile

.PHONY: help build dev prod test clean logs monitor backup restore

# Default target
help:
	@echo "Go Next - Backend Optimization Commands"
	@echo "================================================"
	@echo "Development:"
	@echo "  make dev          - Start development environment"
	@echo "  make dev-build    - Build development containers"
	@echo "  make dev-logs     - View development logs"
	@echo ""
	@echo "Production:"
	@echo "  make prod         - Start production environment"
	@echo "  make prod-build   - Build production containers"
	@echo "  make prod-logs    - View production logs"
	@echo ""
	@echo "Testing:"
	@echo "  make test         - Run backend tests"
	@echo "  make test-coverage - Run tests with coverage"
	@echo ""
	@echo "Maintenance:"
	@echo "  make clean        - Clean up containers and volumes"
	@echo "  make logs         - View all logs"
	@echo "  make monitor      - Open monitoring dashboards"
	@echo "  make backup       - Backup database"
	@echo "  make restore      - Restore database from backup"
	@echo ""
	@echo "Optimization:"
	@echo "  make optimize     - Run performance optimizations"
	@echo "  make benchmark    - Run performance benchmarks"
	@echo "  make health       - Check system health"

# Development commands
dev:
	@echo "Starting development environment..."
	docker-compose -f docker-compose.dev.yml up -d
	@echo "Development environment started!"
	@echo "Frontend: http://localhost:3000"
	@echo "Backend API: http://localhost:8080"
	@echo "Swagger Docs: http://localhost:8080/swagger/index.html"

dev-build:
	@echo "Building development containers..."
	docker-compose -f docker-compose.dev.yml build --no-cache
	@echo "Development containers built!"

dev-logs:
	@echo "Viewing development logs..."
	docker-compose -f docker-compose.dev.yml logs -f

# Production commands
prod:
	@echo "Starting production environment..."
	docker-compose -f docker-compose.prod.yml up -d
	@echo "Production environment started!"
	@echo "Application: https://localhost"
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana: http://localhost:3001 (admin/admin123)"
	@echo "Kibana: http://localhost:5601"

prod-build:
	@echo "Building production containers..."
	docker-compose -f docker-compose.prod.yml build --no-cache
	@echo "Production containers built!"

prod-logs:
	@echo "Viewing production logs..."
	docker-compose -f docker-compose.prod.yml logs -f

# Testing commands
test:
	@echo "Running backend tests..."
	cd backend && go test -v ./...

test-coverage:
	@echo "Running tests with coverage..."
	cd backend && go test -v -coverprofile=coverage.out ./...
	cd backend && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: backend/coverage.html"

# Maintenance commands
clean:
	@echo "Cleaning up containers and volumes..."
	docker-compose -f docker-compose.dev.yml down -v
	docker-compose -f docker-compose.prod.yml down -v
	docker system prune -f
	@echo "Cleanup completed!"

logs:
	@echo "Viewing all logs..."
	docker-compose -f docker-compose.prod.yml logs -f --tail=100

monitor:
	@echo "Opening monitoring dashboards..."
	@echo "Prometheus: http://localhost:9090"
	@echo "Grafana: http://localhost:3001 (admin/admin123)"
	@echo "Kibana: http://localhost:5601"
	@echo "RabbitMQ Management: http://localhost:15672 (admin/admin123)"

backup:
	@echo "Creating database backup..."
	@mkdir -p backups
	docker exec go-next-postgres pg_dump -U wordpress_user wordpress_go_next > backups/backup_$(shell date +%Y%m%d_%H%M%S).sql
	@echo "Backup created in backups/ directory"

restore:
	@echo "Restoring database from backup..."
	@if [ -z "$(BACKUP_FILE)" ]; then \
		echo "Usage: make restore BACKUP_FILE=backups/backup_YYYYMMDD_HHMMSS.sql"; \
		exit 1; \
	fi
	docker exec -i go-next-postgres psql -U wordpress_user wordpress_go_next < $(BACKUP_FILE)
	@echo "Database restored from $(BACKUP_FILE)"

# Optimization commands
optimize:
	@echo "Running performance optimizations..."
	@echo "1. Optimizing database indexes..."
	docker exec go-next-postgres psql -U wordpress_user wordpress_go_next -c "REINDEX DATABASE wordpress_go_next;"
	@echo "2. Analyzing database statistics..."
	docker exec go-next-postgres psql -U wordpress_user wordpress_go_next -c "ANALYZE;"
	@echo "3. Clearing Redis cache..."
	docker exec go-next-redis redis-cli FLUSHALL
	@echo "4. Restarting services..."
	docker-compose -f docker-compose.prod.yml restart backend
	@echo "Optimization completed!"

benchmark:
	@echo "Running performance benchmarks..."
	@if ! command -v wrk &> /dev/null; then \
		echo "Installing wrk benchmark tool..."; \
		if command -v apt-get &> /dev/null; then \
			sudo apt-get update && sudo apt-get install -y wrk; \
		elif command -v yum &> /dev/null; then \
			sudo yum install -y wrk; \
		elif command -v brew &> /dev/null; then \
			brew install wrk; \
		else \
			echo "Please install wrk manually: https://github.com/wg/wrk"; \
			exit 1; \
		fi; \
	fi
	@echo "Benchmarking API endpoints..."
	wrk -t12 -c400 -d30s http://localhost/api/v1/posts
	@echo "Benchmark completed!"

health:
	@echo "Checking system health..."
	@echo "1. Checking container status..."
	docker-compose -f docker-compose.prod.yml ps
	@echo ""
	@echo "2. Checking backend health..."
	curl -f http://localhost/health || echo "Backend health check failed"
	@echo ""
	@echo "3. Checking database connection..."
	docker exec go-next-postgres pg_isready -U wordpress_user -d wordpress_go_next
	@echo ""
	@echo "4. Checking Redis connection..."
	docker exec go-next-redis redis-cli ping
	@echo ""
	@echo "5. Checking disk usage..."
	docker system df
	@echo ""
	@echo "Health check completed!"

# SSL certificate management
ssl-generate:
	@echo "Generating self-signed SSL certificates..."
	@mkdir -p nginx/ssl
	openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
		-keyout nginx/ssl/key.pem \
		-out nginx/ssl/cert.pem \
		-subj "/C=US/ST=State/L=City/O=Organization/CN=localhost"
	@echo "SSL certificates generated in nginx/ssl/"

# Database management
db-migrate:
	@echo "Running database migrations..."
	docker exec go-next-backend ./main migrate

db-seed:
	@echo "Seeding database with sample data..."
	docker exec go-next-backend ./main seed

# Performance monitoring
perf-monitor:
	@echo "Starting performance monitoring..."
	@echo "Monitoring CPU and memory usage..."
	watch -n 1 'docker stats --no-stream'

# Security scan
security-scan:
	@echo "Running security scan..."
	docker run --rm -v /var/run/docker.sock:/var/run/docker.sock \
		-v $(PWD):/app \
		-aquasec/trivy fs /app
	@echo "Security scan completed!"

# Update dependencies
update-deps:
	@echo "Updating Go dependencies..."
	cd backend && go mod tidy && go mod download
	@echo "Dependencies updated!"

# Generate documentation
docs:
	@echo "Generating API documentation..."
	cd backend && swag init -g main.go -o docs
	@echo "Documentation generated!"

# Quick start for development
quick-start:
	@echo "Quick start for development..."
	@echo "1. Building containers..."
	make dev-build
	@echo "2. Starting services..."
	make dev
	@echo "3. Running tests..."
	make test
	@echo "4. Opening monitoring..."
	make monitor
	@echo "Quick start completed!" 