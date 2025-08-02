# Backend with PostgreSQL Docker Compose

This setup provides a Docker Compose configuration for the Go backend with PostgreSQL database.

## 📁 Files

- `docker-compose.yml` - Backend and PostgreSQL Docker Compose configuration
- `Makefile` - Management commands for all operations
- `.env` - Environment variables (create this file)

## 🚀 Quick Start

### Using the Makefile

```bash
# Start backend with PostgreSQL
make up

# Check status
make status

# View logs
make logs

# Stop services
make down

# Restart services
make restart

# Show all available commands
make help
```

### Manual Commands

```bash
# Start services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Check status
docker-compose ps
```

## 🔧 Services

### PostgreSQL Database
- **Image**: `postgres:15-alpine`
- **Port**: `5432`
- **Database**: `go_next`
- **User**: `postgres`
- **Password**: `postgres`
- **Health Check**: `pg_isready`

### Backend Service
- **Port**: `8080`
- **Database**: PostgreSQL
- **Environment**: Development
- **Health Check**: `http://localhost:8080/health`

## 🔧 Configuration

### Environment Variables

Create a `.env` file in the backend directory:

```env
# Application Settings
APP_NAME=go-next-backend
APP_ENV=development
APP_DEBUG=true
APP_URL=http://localhost:3000
API_URL=http://localhost:8080

# Backend Settings
BACKEND_PORT=8080
GIN_MODE=debug
CORS_ORIGIN=http://localhost:3000

# Database Settings (PostgreSQL)
DB_TYPE=postgres
DB_HOST=postgres
DB_PORT=5432
DB_NAME=go_next
DB_USER=postgres
DB_PASSWORD=postgres
DB_SSLMODE=disable

# JWT Settings
JWT_SECRET=your-super-secret-jwt-key-here-change-in-production
JWT_EXPIRATION=24h
```

### Database Configuration

- **PostgreSQL**: Production-ready database
- **Auto-migration**: Enabled for all models
- **Connection Pool**: Optimized settings
- **SSL**: Disabled for development

### Volumes

- `.:/app` - Source code mount
- `backend_data:/app/data` - Persistent data
- `./log:/app/log` - Log files
- `./storage:/app/storage` - File storage
- `./data:/app/data` - Database directory
- `postgres_data:/var/lib/postgresql/data` - PostgreSQL data

## 🏥 Health Checks

### Backend Health Check
```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0"
}
```

### PostgreSQL Health Check
```bash
docker-compose exec postgres pg_isready -U postgres -d go_next
```

## 🔍 API Endpoints

- `GET /health` - Health check
- `GET /swagger/*` - API documentation (Swagger)
- `GET /api/*` - API endpoints

## 🛠️ Development

### Available Makefile Commands

```bash
# Development
make run          # Run backend (go run main.go)
make server       # Run backend server (go run main.go server)
make test         # Run tests
make swag         # Generate Swagger docs

# Docker Compose
make up           # Start backend with PostgreSQL
make down         # Stop services
make build        # Build backend service
make logs         # Show logs
make restart      # Restart services
make status       # Show service status

# Legacy
make docker       # Start with docker-compose up --build
make server-bat   # Run server (batch file equivalent)
make server-ps    # Run server (PowerShell equivalent)

# Cleanup
make clean        # Clean up Docker and docs
make help         # Show all commands
```

### Hot Reload

The backend service uses `go run` for development with hot reload capabilities.

### Database Access

Connect to PostgreSQL:
```bash
# Using docker-compose
docker-compose exec postgres psql -U postgres -d go_next

# Using external client
psql -h localhost -p 5432 -U postgres -d go_next
```

### Logs

View real-time logs:
```bash
# All services
make logs

# Backend only
docker-compose logs -f backend

# PostgreSQL only
docker-compose logs -f postgres
```

## 🔒 Security Notes

1. **Database Password**: Change default password in production
2. **JWT Secret**: Use strong secret in production
3. **SSL**: Enable SSL for production database
4. **Environment**: Set `APP_ENV=production` for production

## 📊 Resource Limits

### Backend
- **Memory**: 2GB limit, 1GB reservation
- **CPU**: 1.0 limit, 0.5 reservation

### PostgreSQL
- **Memory**: 1GB limit, 512MB reservation
- **CPU**: 0.5 limit, 0.25 reservation

## 🚨 Troubleshooting

### Service Won't Start
1. Check logs: `make logs`
2. Verify Docker is running
3. Check port availability: `netstat -an | findstr 8080`

### Database Issues
1. Check PostgreSQL logs: `docker-compose logs postgres`
2. Verify database connection: `docker-compose exec postgres pg_isready`
3. Check environment variables

### Build Issues
1. Check Docker build context
2. Verify Go dependencies
3. Check memory limits

## 📝 Example Usage

```bash
# Start development environment
make up

# Test the API
curl http://localhost:8080/health

# Check database
docker-compose exec postgres psql -U postgres -d go_next -c "SELECT version();"

# View logs
make logs

# Stop when done
make down
```

## 🎯 Benefits

- **Production Ready**: PostgreSQL for robust data storage
- **Development Friendly**: Hot reload and debugging
- **Isolated**: Custom network and volumes
- **Scalable**: Easy to extend with additional services
- **Reliable**: Health checks and restart policies
- **Unified Commands**: All operations through Makefile 