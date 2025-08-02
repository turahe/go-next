# Environment Setup for Admin Frontend

This document explains how to configure environment variables for the admin-frontend application.

## Quick Start

1. Copy the environment template:
   ```bash
   cp env.example .env
   ```

2. Or use the minimal configuration:
   ```bash
   cp env.minimal .env
   ```

3. Update the values in `.env` according to your environment.

**Note:** Both `env.example` and `env.minimal` files are located in the `admin-frontend/` directory for easy access.

## Environment Files

### `env.example` - Complete Configuration
Contains all possible environment variables with detailed comments and default values. Use this for:
- Production deployments
- Complete feature set
- Advanced configuration

### `env.minimal` - Essential Variables Only
Contains only the essential variables that are actually used in the codebase. Use this for:
- Quick development setup
- Minimal configuration
- Basic functionality

## Essential Variables

### API Configuration
```bash
# Backend API URL (required)
VITE_API_URL=http://localhost:8080

# WebSocket URL (optional - auto-derived from API URL)
VITE_WS_URL=ws://localhost:8080
```

### Application Settings
```bash
# Application name
VITE_APP_NAME=Admin Dashboard

# Environment
VITE_NODE_ENV=development

# Debug mode
VITE_DEBUG_MODE=false
```

### Feature Flags
```bash
# Enable WebSocket connections
VITE_ENABLE_WEBSOCKET=true

# Enable real-time notifications
VITE_ENABLE_REALTIME_NOTIFICATIONS=true
```

## Development vs Production

### Development
```bash
VITE_NODE_ENV=development
VITE_DEBUG_MODE=true
VITE_ENABLE_CONSOLE_LOGS=true
VITE_API_URL=http://localhost:8080
```

### Production
```bash
VITE_NODE_ENV=production
VITE_DEBUG_MODE=false
VITE_ENABLE_CONSOLE_LOGS=false
VITE_API_URL=https://your-api-domain.com
```

## Backend Integration

The admin-frontend is designed to work with the Go backend. Ensure your backend is running and accessible at the URL specified in `VITE_API_URL`.

### Backend Requirements
- API endpoints at `/api/v1/*`
- WebSocket endpoint at `/api/v1/ws/connect`
- Authentication endpoints at `/login` and `/register`
- Health check endpoint at `/health`

## WebSocket Configuration

The WebSocket service automatically:
- Converts HTTP URLs to WebSocket URLs (http → ws, https → wss)
- Handles reconnection with exponential backoff
- Manages connection state

## Authentication

The application uses JWT tokens stored in localStorage:
- `authToken` - Main authentication token
- `refreshToken` - Token for refreshing authentication

## File Upload

Configure file upload limits and allowed types:
```bash
VITE_MAX_FILE_SIZE=10485760  # 10MB
VITE_ALLOWED_FILE_TYPES=image/*,application/pdf
```

## Notifications

Configure notification behavior:
```bash
VITE_NOTIFICATION_DURATION=5000  # 5 seconds
VITE_MAX_NOTIFICATIONS=5
```

## Troubleshooting

### Common Issues

1. **API Connection Failed**
   - Check `VITE_API_URL` is correct
   - Ensure backend is running
   - Verify CORS is configured on backend

2. **WebSocket Connection Failed**
   - Check `VITE_WS_URL` is correct
   - Ensure WebSocket endpoint is available
   - Verify WebSocket path is correct

3. **Authentication Issues**
   - Check token storage keys
   - Verify JWT token format
   - Ensure refresh token is working

### Debug Mode

Enable debug mode to see detailed logs:
```bash
VITE_DEBUG_MODE=true
VITE_ENABLE_CONSOLE_LOGS=true
```

## Security Notes

- Never commit `.env` files to version control
- Use different API URLs for development and production
- Consider using environment-specific files (`.env.development`, `.env.production`)
- Validate environment variables at startup

## Vite Environment Variables

All environment variables must be prefixed with `VITE_` to be accessible in the client-side code. Vite automatically exposes these variables to your application.

## Example Usage in Code

```typescript
// Access environment variables
const apiUrl = import.meta.env.VITE_API_URL;
const debugMode = import.meta.env.VITE_DEBUG_MODE;
const appName = import.meta.env.VITE_APP_NAME;

// Use in components
if (import.meta.env.VITE_DEBUG_MODE) {
  console.log('Debug mode enabled');
}
``` 