# Go-Next Admin Panel

A modern, full-stack admin panel built with React (Frontend) and Go (Backend), featuring role-based access control, user management, content management, and real-time capabilities.

## 🚀 Features

### Frontend (React + TypeScript)
- **Modern UI/UX** - Clean, responsive design with Tailwind CSS
- **Dark Mode Support** - Toggle between light and dark themes
- **Role-Based Access Control** - Different permissions for different user roles
- **Authentication System** - Secure login/register with JWT tokens
- **User Management** - Full CRUD operations for user accounts
- **Posts Management** - Content management with status tracking
- **Real-time Notifications** - WebSocket-powered live notifications
- **Data Tables** - Sortable, searchable, and paginated data tables
- **Responsive Design** - Works seamlessly on desktop and mobile
- **Advanced Search** - Meilisearch-powered search with typo tolerance
- **Export Functionality** - CSV/Excel export for data tables

### Backend (Go + Gin)
- **RESTful API** - Clean, well-documented API endpoints
- **JWT Authentication** - Secure token-based authentication
- **Role-Based Authorization** - Casbin-based permission system
- **Database Integration** - GORM with PostgreSQL/MySQL support
- **User Management** - Complete user CRUD with role assignment
- **Content Management** - Posts, categories, and comments
- **Media Management** - File upload and association system
- **Dashboard Statistics** - Real-time analytics and metrics
- **WebSocket Integration** - Real-time notifications and live updates
- **Message Queue System** - RabbitMQ for asynchronous processing
- **Search Engine** - Meilisearch integration for fast search
- **Email Service** - SMTP integration with queue-based delivery
- **WhatsApp Integration** - OTP delivery via WhatsApp Business API

## 🏗️ Architecture

```
go-next/
├── admin-frontend/          # React TypeScript frontend
│   ├── src/
│   │   ├── components/      # Reusable UI components
│   │   ├── pages/          # Page components
│   │   ├── context/        # React context providers
│   │   ├── services/       # API service layer
│   │   └── utils/          # Utility functions
│   └── package.json
├── backend/                 # Go backend API
│   ├── cmd/                # Application entry points
│   ├── internal/           # Private application code
│   │   ├── http/          # HTTP handlers and middleware
│   │   ├── models/        # Database models
│   │   ├── services/      # Business logic
│   │   └── routers/       # Route definitions
│   └── pkg/               # Public packages
├── data/                   # Persistent data storage
│   ├── postgres/          # Database files
│   ├── redis/             # Cache and session data
│   ├── rabbitmq/          # Message queue data
│   ├── meilisearch/       # Search engine data
│   └── waha/              # WhatsApp session data
└── docker-compose.yml      # Development environment
```

## 🛠️ Tech Stack

### Frontend
- **React 19** - Modern React with hooks
- **TypeScript 5.8** - Type-safe JavaScript
- **Tailwind CSS 4.1** - Utility-first CSS framework
- **React Router 7** - Client-side routing
- **Vite 7** - Fast build tool and dev server
- **Socket.io Client** - Real-time WebSocket communication
- **ApexCharts** - Interactive charts and analytics
- **React DnD** - Drag and drop functionality

### Backend
- **Go 1.23** - High-performance language
- **Gin 1.10** - HTTP web framework
- **GORM 1.30** - Database ORM
- **JWT 5.2** - JSON Web Tokens for authentication
- **Casbin 2.108** - Authorization library
- **Gorilla WebSocket** - Real-time communication
- **RabbitMQ** - Message queue system
- **Meilisearch** - Fast search engine
- **Redis** - Caching and session storage
- **PostgreSQL/MySQL** - Database

### Infrastructure
- **Docker & Docker Compose** - Containerized development
- **RabbitMQ** - Message queuing for async operations
- **Meilisearch** - Typo-tolerant search engine
- **Mailpit** - Email testing and development
- **WAHA** - WhatsApp HTTP API integration

## 📚 Documentation

For comprehensive documentation, guides, and setup instructions, see the [Documentation](./docs/) directory:

- **[📚 Documentation Overview](./docs/README.md)** - Complete documentation index
- **[🏗️ Project Documentation](./docs/project/)** - Project structure, roadmap, and implementation guides
- **[🎨 Frontend Documentation](./docs/admin-frontend/)** - React admin panel setup and configuration
- **[⚙️ Backend Documentation](./docs/backend/)** - Go backend setup and configuration
- **[🔌 API Documentation](./docs/api/)** - API guides, Swagger specs, and technical documentation

## 🚀 Quick Start

### Prerequisites
- Node.js 22+ and npm 10+
- Go 1.23+
- Docker and Docker Compose
- PostgreSQL or MySQL

### 1. Clone the Repository
```bash
git clone <repository-url>
cd go-next
```

### 2. Start the Development Environment
```bash
# Start all services (PostgreSQL, Redis, RabbitMQ, Meilisearch, etc.)
docker-compose up -d

# Wait for services to be ready (check health status)
docker-compose ps
```

### 3. Start the Backend
```bash
cd backend
go mod download
go run main.go
```

The backend will start on `http://localhost:8080`

### 4. Start the Frontend
```bash
cd admin-frontend
npm install
npm run dev
```

The frontend will start on `http://localhost:5173`

### 5. Access Services
- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080
- **API Documentation**: http://localhost:8080/swagger/index.html
- **RabbitMQ Management**: http://localhost:15672 (admin/admin123)
- **Meilisearch Dashboard**: http://localhost:7700
- **Email Testing (Mailpit)**: http://localhost:8025

## 📋 User Roles & Permissions

### Admin
- Full access to all features
- User management (create, edit, delete)
- Content management
- System settings
- Role management
- Queue monitoring
- Search analytics

### Editor
- Content creation and editing
- Media management
- Limited user management
- Search functionality

### Moderator
- Content moderation
- Comment management
- User monitoring
- Basic search

### User
- Basic access
- Profile management
- Content viewing
- Search functionality

## 🔧 Configuration

### Environment Variables

#### Backend (.env)
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
RABBITMQ_VHOST=/

# Meilisearch
MEILI_MASTER_KEY=your-super-secret-master-key

# Email
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USER=
SMTP_PASSWORD=
```

#### Frontend (.env)
```env
VITE_API_URL=http://localhost:8080/api
VITE_WS_URL=ws://localhost:8080/ws
```

**Quick Setup:**
```bash
# Copy the minimal environment template
cp admin-frontend/env.minimal admin-frontend/.env

# Or copy the complete template
cp admin-frontend/env.example admin-frontend/.env
```

## 📚 API Documentation

### Authentication Endpoints
- `POST /api/login` - User login
- `POST /api/register` - User registration
- `POST /api/v1/auth/refresh` - Refresh JWT token

### User Management
- `GET /api/v1/users` - List users (paginated)
- `GET /api/v1/users/:id` - Get user profile
- `POST /api/v1/users` - Create user
- `PUT /api/v1/users/:id` - Update user
- `DELETE /api/v1/users/:id` - Delete user

### Dashboard
- `GET /api/v1/dashboard/stats` - Get dashboard statistics

### Content Management
- `GET /api/v1/posts` - List posts
- `POST /api/v1/posts` - Create post
- `PUT /api/v1/posts/:id` - Update post
- `DELETE /api/v1/posts/:id` - Delete post

### Search
- `GET /api/v1/search` - Search across content
- `GET /api/v1/search/suggestions` - Get search suggestions

### WebSocket
- `WS /ws` - Real-time notifications and updates

## 🎨 UI Components

### Reusable Components
- **DataTable** - Sortable, searchable data tables
- **UserModal** - User creation/editing modal
- **Layout** - Main application layout
- **Sidebar** - Navigation sidebar
- **Header** - Top navigation bar
- **SearchComponent** - Advanced search with filters
- **NotificationDropdown** - Real-time notifications
- **ChartComponents** - Interactive charts and analytics

### Styling
- **Tailwind CSS 4.1** - Utility-first styling
- **Dark Mode** - Automatic theme switching
- **Responsive** - Mobile-first design
- **Accessibility** - WCAG compliant

## 🔒 Security Features

- **JWT Authentication** - Secure token-based auth
- **Role-Based Access Control** - Granular permissions with Casbin
- **CORS Protection** - Cross-origin request handling
- **Input Validation** - Request validation and sanitization
- **Rate Limiting** - API rate limiting
- **Password Hashing** - Secure password storage
- **WebSocket Security** - Authenticated WebSocket connections

## 📊 Dashboard Features

- **Real-time Statistics** - Live data from database
- **User Analytics** - User growth and activity
- **Content Metrics** - Posts and comments tracking
- **Revenue Tracking** - Financial metrics (mock data)
- **Quick Actions** - Common admin tasks
- **Search Analytics** - Search performance metrics
- **Queue Monitoring** - RabbitMQ queue status

## 🚀 Deployment

### Frontend Deployment
```bash
cd admin-frontend
npm run build
# Deploy dist/ folder to your hosting service
```

### Backend Deployment
```bash
cd backend
go build -o main cmd/main.go
# Deploy the binary to your server
```

### Docker Deployment
```bash
docker-compose -f docker-compose.prod.yml up -d
```

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support and questions:
- Create an issue in the repository
- Check the documentation
- Review the API endpoints

## 🔮 Roadmap

### ✅ Completed Features
- [x] Real-time notifications (WebSocket implementation)
- [x] File upload system (Backend media service implemented)
- [x] Email notifications (SMTP service with queue)
- [x] Basic audit logging (Logging infrastructure with Sentry)
- [x] API rate limiting (Backend middleware implemented)
- [x] WebSocket integration (Real-time communication)
- [x] RabbitMQ queue system (Message queuing)
- [x] Meilisearch search engine (Fast, typo-tolerant search)
- [x] WhatsApp integration (OTP delivery)
- [x] Advanced search functionality (Frontend and backend)
- [x] Export functionality (CSV/Excel export)

### 🚧 In Progress
- [ ] Advanced analytics dashboard (Enhanced charts and metrics)
- [ ] Multi-language support (i18n infrastructure)
- [ ] Micro-frontend architecture (Container application)

### 🔄 Next Phase (Q1 2025)
- [ ] **AI-Powered Content Management** - AI-assisted content creation and optimization
- [ ] **Forgot Password Feature** - Complete email-based password reset flow
- [ ] **Internationalization** - Multi-language support with i18n
- [ ] **Audit Trail UI** - User activity tracking interface
- [ ] **API Documentation UI** - Interactive API explorer
- [ ] **Bulk Operations** - Mass actions for users and content
- [ ] **Advanced Media Management** - Image editing and optimization
- [ ] **Notification Preferences** - User-configurable notification settings

### 🔮 Future Features (Q2 2025)
- [ ] **Mobile App** - React Native mobile application
- [ ] **Advanced RBAC** - Dynamic role creation and permission management
- [ ] **Workflow Engine** - Content approval workflows
- [ ] **API Versioning** - Backward-compatible API evolution
- [ ] **Performance Monitoring** - Real-time performance metrics
- [ ] **Backup & Recovery** - Automated backup system
- [ ] **Multi-tenancy** - Support for multiple organizations
- [ ] **Plugin System** - Extensible architecture for custom modules
- [ ] **Advanced Reporting** - Custom report builder
- [ ] **Integration Hub** - Third-party service integrations

---

Built with ❤️ using React and Go 