# Go-Next Admin Panel

A modern, full-stack admin panel built with React (Frontend) and Go (Backend), featuring role-based access control, user management, and content management capabilities.

## 🚀 Features

### Frontend (React + TypeScript)
- **Modern UI/UX** - Clean, responsive design with Tailwind CSS
- **Dark Mode Support** - Toggle between light and dark themes
- **Role-Based Access Control** - Different permissions for different user roles
- **Authentication System** - Secure login/register with JWT tokens
- **User Management** - Full CRUD operations for user accounts
- **Posts Management** - Content management with status tracking
- **Real-time Notifications** - Toast notifications for user feedback
- **Data Tables** - Sortable, searchable, and paginated data tables
- **Responsive Design** - Works seamlessly on desktop and mobile

### Backend (Go + Gin)
- **RESTful API** - Clean, well-documented API endpoints
- **JWT Authentication** - Secure token-based authentication
- **Role-Based Authorization** - Casbin-based permission system
- **Database Integration** - GORM with PostgreSQL/MySQL support
- **User Management** - Complete user CRUD with role assignment
- **Content Management** - Posts, categories, and comments
- **Media Management** - File upload and association system
- **Dashboard Statistics** - Real-time analytics and metrics

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
└── docker-compose.yml      # Development environment
```

## 🛠️ Tech Stack

### Frontend
- **React 19** - Modern React with hooks
- **TypeScript** - Type-safe JavaScript
- **Tailwind CSS** - Utility-first CSS framework
- **React Router** - Client-side routing
- **Vite** - Fast build tool and dev server

### Backend
- **Go 1.21+** - High-performance language
- **Gin** - HTTP web framework
- **GORM** - Database ORM
- **JWT** - JSON Web Tokens for authentication
- **Casbin** - Authorization library
- **PostgreSQL/MySQL** - Database

## 📚 Documentation

For comprehensive documentation, guides, and setup instructions, see the [Documentation](./docs/) directory:

- **[📚 Documentation Overview](./docs/README.md)** - Complete documentation index
- **[🏗️ Project Documentation](./docs/project/)** - Project structure, roadmap, and implementation guides
- **[🎨 Frontend Documentation](./docs/admin-frontend/)** - React admin panel setup and configuration
- **[⚙️ Backend Documentation](./docs/backend/)** - Go backend setup and configuration
- **[🔌 API Documentation](./docs/api/)** - API guides, Swagger specs, and technical documentation

## 🚀 Quick Start

### Prerequisites
- Node.js 18+ and npm
- Go 1.21+
- Docker and Docker Compose
- PostgreSQL or MySQL

### 1. Clone the Repository
```bash
git clone <repository-url>
cd go-next
```

### 2. Start the Backend
```bash
cd backend
go mod download
go run main.go
```

The backend will start on `http://localhost:8080`

### 3. Start the Frontend
```bash
cd admin-frontend
npm install
npm run dev
```

The frontend will start on `http://localhost:5173`

### 4. Using Docker (Alternative)
```bash
docker-compose up -d
```

## 📋 User Roles & Permissions

### Admin
- Full access to all features
- User management (create, edit, delete)
- Content management
- System settings
- Role management

### Editor
- Content creation and editing
- Media management
- Limited user management

### Moderator
- Content moderation
- Comment management
- User monitoring

### User
- Basic access
- Profile management
- Content viewing

## 🔧 Configuration

### Environment Variables

#### Backend (.env)
```env
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=password
DB_NAME=admin_panel
JWT_SECRET=your-secret-key
REDIS_URL=redis://localhost:6379
```

#### Frontend (.env)
```env
VITE_API_URL=http://localhost:8080/api
```

**Quick Setup:**
```bash
# Copy the minimal environment template
cp admin-frontend/env.minimal admin-frontend/.env

# Or copy the complete template from docs
cp docs/admin-frontend/env.example admin-frontend/.env
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

## 🎨 UI Components

### Reusable Components
- **DataTable** - Sortable, searchable data tables
- **UserModal** - User creation/editing modal
- **Layout** - Main application layout
- **Sidebar** - Navigation sidebar
- **Header** - Top navigation bar

### Styling
- **Tailwind CSS** - Utility-first styling
- **Dark Mode** - Automatic theme switching
- **Responsive** - Mobile-first design
- **Accessibility** - WCAG compliant

## 🔒 Security Features

- **JWT Authentication** - Secure token-based auth
- **Role-Based Access Control** - Granular permissions
- **CORS Protection** - Cross-origin request handling
- **Input Validation** - Request validation and sanitization
- **Rate Limiting** - API rate limiting
- **Password Hashing** - Secure password storage

## 📊 Dashboard Features

- **Real-time Statistics** - Live data from database
- **User Analytics** - User growth and activity
- **Content Metrics** - Posts and comments tracking
- **Revenue Tracking** - Financial metrics (mock data)
- **Quick Actions** - Common admin tasks

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
- [x] Real-time notifications (Frontend context implemented)
- [x] File upload system (Backend media service implemented)
- [x] Email notifications (SMTP service implemented)
- [x] Basic audit logging (Logging infrastructure in place)
- [x] API rate limiting (Backend middleware implemented)

### 🚧 In Progress
- [ ] Advanced analytics dashboard (Basic stats implemented, needs enhancement)
- [ ] Multi-language support (i18n infrastructure needed)
- [ ] Advanced search and filtering (Basic search exists, needs improvement)

### 🔄 Next Phase (Q1 2024)
- [ ] **WebSocket Integration** - Real-time notifications with WebSocket
- [ ] **Advanced Analytics** - Enhanced dashboard with charts and metrics
- [ ] **AI-Powered Content Management** - AI-assisted content creation and optimization
- [ ] **Forgot Password Feature** - Complete email-based password reset flow
- [ ] **Internationalization** - Multi-language support with i18n
- [ ] **Advanced Search** - Full-text search with filters and sorting
- [ ] **Export Functionality** - CSV/Excel export for data tables
- [ ] **Audit Trail UI** - User activity tracking interface
- [ ] **API Documentation UI** - Interactive API explorer
- [ ] **Bulk Operations** - Mass actions for users and content
- [ ] **Advanced Media Management** - Image editing and optimization
- [ ] **Notification Preferences** - User-configurable notification settings

### 🔮 Future Features (Q2 2024)
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