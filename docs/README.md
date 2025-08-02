# 📚 Documentation

Welcome to the comprehensive documentation for the Go-Next project. This documentation is organized by component and functionality to help you find the information you need quickly.

## 📁 Documentation Structure

### 🏗️ Project Documentation (`/project/`)
High-level project documentation and planning:

- **[MIGRATION_SCRIPT.md](./project/MIGRATION_SCRIPT.md)** - Migration script for project structure improvement
- **[PROJECT_STRUCTURE_IMPROVEMENT.md](./project/PROJECT_STRUCTURE_IMPROVEMENT.md)** - Project structure improvement guidelines
- **[IMPLEMENTATION_EXAMPLES.md](./project/IMPLEMENTATION_EXAMPLES.md)** - Implementation examples and patterns
- **[ROADMAP.md](./project/ROADMAP.md)** - Project roadmap and future plans

### 🎨 Admin Frontend (`/admin-frontend/`)
Documentation for the React-based admin dashboard:

- **[README.md](./admin-frontend/README.md)** - Admin frontend overview and setup
- **[ENVIRONMENT_SETUP.md](./admin-frontend/ENVIRONMENT_SETUP.md)** - Environment configuration guide
- **env.example** - Complete environment variables template (located in `admin-frontend/` directory)
- **env.minimal** - Minimal environment variables template (located in `admin-frontend/` directory)

### ⚙️ Backend (`/backend/`)
Documentation for the Go backend server:

- **[README.md](./backend/README.md)** - Backend overview and setup instructions

### 🔌 API Documentation (`/api/`)
API documentation and technical guides:

- **[docs.go](./api/docs.go)** - Auto-generated API documentation
- **[swagger.json](./api/swagger.json)** - Swagger API specification (JSON)
- **[swagger.yaml](./api/swagger.yaml)** - Swagger API specification (YAML)
- **[PAGINATION_GUIDE.md](./api/PAGINATION_GUIDE.md)** - Pagination implementation guide
- **[PAGINATION_IMPLEMENTATION_SUMMARY.md](./api/PAGINATION_IMPLEMENTATION_SUMMARY.md)** - Pagination implementation summary
- **[VALIDATION_GUIDE.md](./api/VALIDATION_GUIDE.md)** - Input validation guide
- **[REDIS_TOKEN_CACHING.md](./api/REDIS_TOKEN_CACHING.md)** - Redis token caching implementation
- **[REDIS_TOKEN_CACHING_SUMMARY.md](./api/REDIS_TOKEN_CACHING_SUMMARY.md)** - Redis token caching summary
- **[MODELS_OPTIMIZATION_SUMMARY.md](./api/MODELS_OPTIMIZATION_SUMMARY.md)** - Database models optimization summary

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

## 📋 Documentation Standards

- All documentation should be in Markdown format
- Use clear, descriptive file names
- Include code examples where appropriate
- Keep documentation up-to-date with code changes
- Use consistent formatting and structure

## 🤝 Contributing to Documentation

When adding new documentation:

1. Place files in the appropriate subdirectory
2. Update this README.md to include new files
3. Follow the existing naming conventions
4. Include clear descriptions and examples
5. Link related documentation where appropriate

## 📞 Support

If you need help with the documentation or have questions:

1. Check the relevant documentation section first
2. Review the implementation examples
3. Check the project roadmap for planned features
4. Create an issue for missing or unclear documentation

---

*Last updated: $(date)* 