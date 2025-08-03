# 📚 Documentation

Welcome to the comprehensive documentation for the Go-Next project. This documentation is organized by component and functionality to help you find the information you need quickly.

## 📁 Documentation Structure

### 🏗️ Project Documentation (`/project/`)
High-level project documentation and planning:

- **[MIGRATION_SCRIPT.md](./project/MIGRATION_SCRIPT.md)** - Migration script for project structure improvement
- **[PROJECT_STRUCTURE_IMPROVEMENT.md](./project/PROJECT_STRUCTURE_IMPROVEMENT.md)** - Project structure improvement guidelines
- **[IMPLEMENTATION_EXAMPLES.md](./project/IMPLEMENTATION_EXAMPLES.md)** - Implementation examples and patterns
- **[ROADMAP.md](./project/ROADMAP.md)** - Project roadmap and future plans

### 📊 Implementation Status
- **[IMPLEMENTATION_STATUS.md](./IMPLEMENTATION_STATUS.md)** - Complete feature implementation status and metrics
- **[QUICK_REFERENCE.md](./QUICK_REFERENCE.md)** - Quick reference guide for common tasks and configurations

### 🎨 Admin Frontend (`/admin-frontend/`)
Documentation for the React-based admin dashboard:

- **[README.md](./admin-frontend/README.md)** - Admin frontend overview and setup
- **[ENVIRONMENT_SETUP.md](./admin-frontend/ENVIRONMENT_SETUP.md)** - Environment configuration guide
- **env.example** - Complete environment variables template (located in `admin-frontend/` directory)
- **env.minimal** - Minimal environment variables template (located in `admin-frontend/` directory)

### ⚙️ Backend (`/backend/`)
Documentation for the Go backend server:

- **[README.md](./backend/README.md)** - Backend overview and setup instructions

### 🚀 API Documentation (`/api/`)
Comprehensive API documentation and implementation guides:

- **[SEARCH_IMPLEMENTATION.md](./api/SEARCH_IMPLEMENTATION.md)** - Complete Meilisearch search implementation guide
- **[SEARCH_INTEGRATION.md](./api/SEARCH_INTEGRATION.md)** - Search indexing integration with CRUD operations
- **[EMAIL_VERIFICATION.md](./api/EMAIL_VERIFICATION.md)** - Email verification implementation guide
- **[LOGIN_SUCCESS_EMAIL.md](./api/LOGIN_SUCCESS_EMAIL.md)** - Login success email notification implementation
- **[LOGOUT_IMPLEMENTATION.md](./api/LOGOUT_IMPLEMENTATION.md)** - Logout functionality with Redis token blacklisting
- **[API_DOCUMENTATION.md](./api/API_DOCUMENTATION.md)** - Complete API reference and examples
- **[AUTHENTICATION.md](./api/AUTHENTICATION.md)** - JWT authentication implementation
- **[AUTHORIZATION.md](./api/AUTHORIZATION.md)** - Casbin RBAC authorization system
- **[WEBSOCKET_API.md](./api/WEBSOCKET_API.md)** - Real-time WebSocket API documentation
- **[RABBITMQ_INTEGRATION.md](./api/RABBITMQ_INTEGRATION.md)** - Message queue integration guide

## 🚀 Quick Start

### For Developers
1. Start with the [Project README](../README.md) for an overview
2. Check the [Roadmap](./project/ROADMAP.md) for current development status
3. Review [Implementation Examples](./project/IMPLEMENTATION_EXAMPLES.md) for coding patterns

### For Frontend Development
1. Read the [Admin Frontend README](./admin-frontend/README.md)
2. Configure environment variables using [Environment Setup](./admin-frontend/ENVIRONMENT_SETUP.md)
3. Use the environment templates: `env.example` or `env.minimal` (both in admin-frontend directory)

### For Backend Development
1. Read the [Backend README](./backend/README.md)
2. Review the [API Documentation](./api/docs.go)
3. Check the [Swagger Specification](./api/swagger.json) for API endpoints

### For API Integration
1. Review the [Swagger Documentation](./api/swagger.yaml)
2. Check [Validation Guide](./api/VALIDATION_GUIDE.md) for input validation
3. Review [Pagination Guide](./api/PAGINATION_GUIDE.md) for list endpoints
4. Study [Casbin Implementation](./api/CASBIN_IMPLEMENTATION.md) for authorization

## 🔧 Environment Setup

### Frontend Environment
```bash
# Copy environment template (both files are in admin-frontend directory)
cp admin-frontend/env.example admin-frontend/.env

# Or use minimal configuration
cp admin-frontend/env.minimal admin-frontend/.env
```

### Backend Environment
```bash
# Follow the backend README for setup
# See: docs/backend/README.md
```

### Infrastructure Services
```bash
# Start all services with Docker Compose
docker-compose up -d

# Access service dashboards:
# - RabbitMQ Management: http://localhost:15672 (admin/admin123)
# - Meilisearch Dashboard: http://localhost:7700
# - Email Testing (Mailpit): http://localhost:8025
```

## 🆕 Recently Implemented Features

### ✅ Completed (Q4 2024)
- **WebSocket Integration** - Real-time notifications and live updates
- **RabbitMQ Queue System** - Asynchronous message processing
- **Meilisearch Search Engine** - Fast, typo-tolerant search
- **WhatsApp Integration** - OTP delivery via WhatsApp Business API
- **Advanced Search** - Frontend and backend search functionality
- **Export Functionality** - CSV/Excel export for data tables
- **Enhanced RBAC** - Organization-based role management
- **Queue Monitoring** - RabbitMQ queue health monitoring

### 🚧 In Progress
- **Advanced Analytics Dashboard** - Enhanced charts and metrics
- **Multi-language Support** - i18n infrastructure
- **Micro-frontend Architecture** - Container application setup

## 📋 Documentation Standards

- All documentation should be in Markdown format
- Use clear, descriptive file names
- Include code examples where appropriate
- Keep documentation up-to-date with code changes
- Use consistent formatting and structure
- Include implementation status and completion dates

## 🤝 Contributing to Documentation

When adding new documentation:

1. Place files in the appropriate subdirectory
2. Update this README.md to include new files
3. Follow the existing naming conventions
4. Include clear descriptions and examples
5. Link related documentation where appropriate
6. Update the "Recently Implemented Features" section

## 📞 Support

If you need help with the documentation or have questions:

1. Check the relevant documentation section first
2. Review the implementation examples
3. Check the project roadmap for planned features
4. Create an issue for missing or unclear documentation

## 🔗 Quick Links

### Development
- [Project Overview](../README.md)
- [API Documentation](./api/docs.go)
- [Swagger UI](http://localhost:8080/swagger/index.html)

### Services
- [RabbitMQ Management](http://localhost:15672)
- [Meilisearch Dashboard](http://localhost:7700)
- [Email Testing](http://localhost:8025)

### Documentation
- [Implementation Examples](./project/IMPLEMENTATION_EXAMPLES.md)
- [RBAC Implementation](./api/CASBIN_IMPLEMENTATION.md)
- [Search Implementation](./api/README.md)

---

*Last updated: December 2024* 