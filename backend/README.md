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

## Prerequisites
- Docker & Docker Compose
- PowerShell (for Windows automation scripts)
- Go 1.24+ (for local development outside Docker)

## Environment Setup

This project uses environment-specific `.env` files for configuration:
- `backend/.env.development`
- `backend/.env.staging`
- `backend/.env.production`

You can generate these files automatically using the provided PowerShell script:

```powershell
# In the backend directory
./setup-env.ps1
```

This will create all three `.env` files with example values if they do not already exist.

## Switching Environments

To switch between development, staging, and production environments, use the `switch-env.ps1` script:

```powershell
# Usage:
./switch-env.ps1 -envType development   # or staging, production
```

This sets the appropriate `ENV` and `BUILD_TARGET` environment variables for Docker Compose.

## Running the Backend with Docker Compose

### Development (default)
```powershell
# Ensure ENV and BUILD_TARGET are set (default is development)
docker compose up --build
```

### Staging
```powershell
./switch-env.ps1 -envType staging
docker compose up --build
```

### Production
```powershell
./switch-env.ps1 -envType production
docker compose up --build
```

## Manual Environment Switching
If you prefer, you can set the environment variables manually:
```powershell
$env:ENV="production"
$env:BUILD_TARGET="prod"
docker compose up --build
```

## Healthchecks & Logging
- The backend service includes a healthcheck on `/health`.
- Logging is configured with rotation (max 10MB per file, 3 files).

## Volumes
- In development, the source code is mounted for hot reload.
- In staging/production, remove or set the volume to read-only for security.

## Troubleshooting
- Ensure the correct `.env` file exists for your environment.
- If you change environment variables, rebuild the containers: `docker compose up --build`
- For port conflicts, adjust the `ports` section in `docker-compose.yml`.

## Scripts
- `setup-env.ps1`: Generates all required `.env` files with example values.
- `switch-env.ps1`: Sets environment variables for Docker Compose based on the target environment.

### Getting Started

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