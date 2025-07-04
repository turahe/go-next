# WordPress Go Next Backend

A modular, production-ready Golang backend for a WordPress-like web application. Features JWT authentication, Casbin RBAC, PostgreSQL, S3-compatible media uploads, and a clean service/controller architecture.

## Features
- Modular Go structure (Gin, GORM, PostgreSQL)
- JWT authentication & Casbin RBAC (admin, editor, moderator, user, guest)
- User, Post, Comment, Category, Role, Media models
- Media uploads to S3/MinIO (via casdoor/oss)
- Polymorphic many-to-many media association
- Email/phone verification & password reset
- Docker Compose for local dev (PostgreSQL, MinIO)
- Swagger (OpenAPI) API docs
- Service interfaces for business logic
- Test scaffolding for endpoints

## Getting Started

### Prerequisites
- Go 1.20+
- Docker & Docker Compose
- (Optional) swag CLI for Swagger docs: `go install github.com/swaggo/swag/cmd/swag@latest`

### Setup
1. Clone the repo:
   ```sh
   git clone <repo-url>
   cd wordpress-go-next/backend
   ```
2. Copy and edit environment variables:
   ```sh
   cp .env.example .env
   # Edit .env as needed
   ```
3. Start services with Docker Compose:
   ```sh
   docker-compose up --build
   ```
   - PostgreSQL: `localhost:5432`
   - MinIO: `localhost:9000` (UI: `localhost:9001`)
   - Backend: `localhost:8080`

4. Create the `media` bucket in MinIO (via UI or `mc` CLI).

### Running Locally
- Start backend only:
  ```sh
  go run main.go
  ```
- Run tests:
  ```sh
  go test ./...
  ```

### API Documentation
- Swagger UI: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
- Regenerate docs after code changes:
  ```sh
  swag init -g main.go -o docs
  ```

### Environment Variables
See `.env.example` for all required variables, including:
- Database (PostgreSQL)
- JWT secret
- S3/MinIO credentials
- Email/SMS provider settings

### Project Structure
- `cmd/` - Entrypoints
- `internal/` - Controllers, services, models, middleware, routers
- `pkg/` - Config, database, logger, storage, utils
- `docs/` - Swagger docs
- `tests/` - API tests

### Useful Commands
- `make run` - Run backend
- `make test` - Run tests
- `make swag` - Regenerate Swagger docs
- `make docker` - Build and run with Docker Compose

---

## License
MIT 