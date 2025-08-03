# 🚀 Quick Reference Guide

This guide provides quick access to the most commonly used features and configurations in the Go-Next Admin Panel.

## 🏃‍♂️ Quick Start

### 1. Start Development Environment
```bash
# Start all services
docker-compose up -d

# Check service health
docker-compose ps
```

### 2. Start Backend
```bash
cd backend
go mod download
go run main.go
```

### 3. Start Frontend
```bash
cd admin-frontend
npm install
npm run dev
```

### 4. Access Services
- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080
- **API Docs**: http://localhost:8080/swagger/index.html
- **RabbitMQ**: http://localhost:15672 (admin/admin123)
- **Meilisearch**: http://localhost:7700
- **Email Testing**: http://localhost:8025

## 🔧 Environment Configuration

### Frontend Environment
```bash
# Copy environment template
cp admin-frontend/env.example admin-frontend/.env

# Or minimal config
cp admin-frontend/env.minimal admin-frontend/.env
```

### Backend Environment
```env
# Database
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=postgres
DB_NAME=go_next

# JWT
JWT_SECRET=your-secret-key

# Redis
REDIS_URL=redis://localhost:6379
REDIS_PASSWORD=redis_password

# RabbitMQ
RABBITMQ_USER=admin
RABBITMQ_PASSWORD=admin123

# Meilisearch
MEILI_MASTER_KEY=your-super-secret-master-key
```

## 🔌 API Endpoints

### Authentication
```bash
POST /api/login          # User login
POST /api/register       # User registration
POST /api/v1/auth/refresh # Refresh JWT token
```

### Users
```bash
GET    /api/v1/users     # List users (paginated)
GET    /api/v1/users/:id # Get user profile
POST   /api/v1/users     # Create user
PUT    /api/v1/users/:id # Update user
DELETE /api/v1/users/:id # Delete user
```

### Content
```bash
GET    /api/v1/posts     # List posts
POST   /api/v1/posts     # Create post
PUT    /api/v1/posts/:id # Update post
DELETE /api/v1/posts/:id # Delete post
```

### Search
```bash
GET /api/v1/search           # Search across content
GET /api/v1/search/suggestions # Get search suggestions
```

### WebSocket
```bash
WS /ws # Real-time notifications and updates
```

## 🆕 New Features (Q4 2024)

### WebSocket Integration
- **Real-time notifications** - Live updates via WebSocket
- **Connection management** - Automatic reconnection
- **User-specific channels** - Targeted notifications

### RabbitMQ Queue System
- **Asynchronous processing** - Background job handling
- **Email queue** - Queue-based email delivery
- **WhatsApp integration** - OTP delivery via WhatsApp
- **Queue monitoring** - Health monitoring and alerts

### Meilisearch Search
- **Typo-tolerant search** - Handles spelling mistakes
- **Instant search** - Real-time search results
- **Advanced filtering** - Multiple filter options
- **Search analytics** - Performance tracking

### Enhanced RBAC
- **Organization-based permissions** - Multi-org support
- **Dynamic role management** - Flexible role creation
- **Menu management** - Dynamic navigation
- **Granular permissions** - Fine-grained access control

### Export Functionality
- **CSV export** - Data export capabilities
- **Excel export** - Spreadsheet export
- **Bulk operations** - Mass data actions
- **Filtered exports** - Export filtered data

## 🛠️ Development Commands

### Backend
```bash
# Run development server
go run main.go

# Run tests
go test ./...

# Build binary
go build -o main cmd/main.go

# Database migrations
go run cmd/migrate.go
```

### Frontend
```bash
# Development server
npm run dev

# Build for production
npm run build

# Lint code
npm run lint

# Preview build
npm run preview
```

### Docker
```bash
# Start all services
docker-compose up -d

# View logs
docker-compose logs -f

# Stop services
docker-compose down

# Rebuild services
docker-compose up -d --build
```

## 📊 Service Health Checks

### Database (PostgreSQL)
```bash
# Check connection
pg_isready -h localhost -p 5432 -U postgres

# Check database
psql -h localhost -U postgres -d go_next -c "SELECT 1;"
```

### Redis
```bash
# Check Redis
redis-cli -h localhost -p 6379 ping

# Check with password
redis-cli -h localhost -p 6379 -a redis_password ping
```

### RabbitMQ
```bash
# Check RabbitMQ
curl -u admin:admin123 http://localhost:15672/api/overview
```

### Meilisearch
```bash
# Check Meilisearch
curl http://localhost:7700/health
```

## 🔍 Troubleshooting

### Common Issues

#### Backend won't start
```bash
# Check database connection
go run cmd/check-db.go

# Check environment variables
go run cmd/check-env.go
```

#### Frontend build fails
```bash
# Clear node modules
rm -rf node_modules package-lock.json
npm install

# Check TypeScript errors
npx tsc --noEmit
```

#### WebSocket connection issues
```bash
# Check WebSocket endpoint
curl -i -N -H "Connection: Upgrade" -H "Upgrade: websocket" \
  -H "Sec-WebSocket-Version: 13" -H "Sec-WebSocket-Key: test" \
  http://localhost:8080/ws
```

#### Search not working
```bash
# Check Meilisearch health
curl http://localhost:7700/health

# Check search indexes
curl http://localhost:7700/indexes
```

## 📚 Useful Documentation

### Core Documentation
- [Project Overview](../README.md)
- [Implementation Status](./IMPLEMENTATION_STATUS.md)
- [API Documentation](./api/docs.go)

### Technical Guides
- [RBAC Implementation](./api/CASBIN_IMPLEMENTATION.md)
- [Search Implementation](./api/README.md)
- [WebSocket Guide](./api/README.md)

### Development
- [Implementation Examples](./project/IMPLEMENTATION_EXAMPLES.md)
- [Project Roadmap](./project/ROADMAP.md)
- [Environment Setup](./admin-frontend/ENVIRONMENT_SETUP.md)

## 🔗 Quick Links

### Development Tools
- [Swagger UI](http://localhost:8080/swagger/index.html)
- [RabbitMQ Management](http://localhost:15672)
- [Meilisearch Dashboard](http://localhost:7700)
- [Email Testing](http://localhost:8025)

### Documentation
- [API Reference](./api/docs.go)
- [Implementation Status](./IMPLEMENTATION_STATUS.md)
- [Project Roadmap](./project/ROADMAP.md)

---

*Last updated: December 2024* 