# Admin Dashboard Frontend

A modern, responsive admin dashboard built with React, TypeScript, and Tailwind CSS. This application provides a comprehensive admin interface with user authentication, role-based access control, and data management capabilities.

## Features

### 🔐 Authentication & Authorization
- **User Authentication**: Login and registration system with form validation
- **Role-Based Access Control**: Three user roles (Admin, Editor, Viewer) with different permissions
- **Protected Routes**: Automatic redirection for unauthenticated users
- **Session Management**: Persistent login state with localStorage

### 📊 Dashboard
- **Overview Statistics**: Key metrics displayed in responsive cards
- **Real-time Data**: Mock API integration for demonstration
- **Quick Actions**: Easy access to common admin tasks
- **Recent Activity**: Timeline of system events

### 👥 User Management
- **User CRUD Operations**: Create, read, update, and delete user accounts
- **Advanced Filtering**: Filter by role, status, and search queries
- **Data Table**: Sortable and paginated user list
- **Status Management**: Inline status updates for users
- **Permission Checks**: Role-based access to sensitive operations

### 🎨 User Interface
- **Responsive Design**: Fully responsive layout that works on all devices
- **Dark Mode**: Toggle between light and dark themes
- **Modern UI**: Clean, professional design using Tailwind CSS
- **Accessibility**: Keyboard navigation and screen reader support

### 🔔 Notifications System
- **Real-time Notifications**: Toast-style notifications for user actions
- **Notification Center**: Dropdown with notification history
- **Multiple Types**: Success, error, warning, and info notifications
- **Mark as Read**: Interactive notification management

### 📱 Responsive Layout
- **Sidebar Navigation**: Collapsible sidebar with role-based menu items
- **Header**: User profile, notifications, and quick actions
- **Mobile-Friendly**: Optimized for mobile and tablet devices
- **Flexible Grid**: Adaptive layout system

## Technology Stack

- **React 18**: Modern React with hooks and functional components
- **TypeScript**: Type-safe development with strict type checking
- **Tailwind CSS**: Utility-first CSS framework for styling
- **React Router**: Client-side routing and navigation
- **Vite**: Fast build tool and development server
- **Context API**: State management for authentication and notifications

## Project Structure

```
src/
├── components/          # Reusable UI components
│   ├── Auth/           # Authentication components
│   ├── DataTable/      # Data table component
│   ├── Layout/         # Layout components (Sidebar, Header)
│   └── Notifications/  # Notification components
├── context/            # React Context providers
│   ├── AuthContext.tsx     # Authentication state
│   ├── DarkModeContext.tsx # Theme management
│   └── NotificationContext.tsx # Notifications state
├── pages/              # Page components
│   ├── Auth/           # Authentication pages
│   ├── Dashboard/      # Dashboard page
│   └── Users/          # User management page
├── services/           # API services and utilities
├── types/              # TypeScript type definitions
├── utils/              # Utility functions
└── hooks/              # Custom React hooks
```

## Getting Started

### Prerequisites

- Node.js (version 16 or higher)
- npm or yarn package manager

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd admin-frontend
   ```

2. **Install dependencies**
   ```bash
   npm install
   ```

3. **Start the development server**
   ```bash
   npm run dev
   ```

4. **Open your browser**
   Navigate to `http://localhost:3000`

### Available Scripts

- `npm run dev` - Start development server
- `npm run build` - Build for production
- `npm run preview` - Preview production build
- `npm run lint` - Run ESLint

## Usage

### Authentication

1. **Login**: Use the demo credentials or register a new account
   - Demo: `admin@example.com` / `Admin123!`
   - Or register a new account

2. **Role-based Access**:
   - **Admin**: Full access to all features
   - **Editor**: Can manage users but not delete them
   - **Viewer**: Read-only access to dashboard

### User Management

1. **View Users**: Navigate to the Users page to see all users
2. **Filter Users**: Use the search bar and filters to find specific users
3. **Update Status**: Change user status directly from the table
4. **Delete Users**: Admin users can delete user accounts (with confirmation)

### Dark Mode

Toggle dark mode using the switch in the sidebar. The preference is saved in localStorage.

## API Integration

The application includes a mock API service that simulates backend interactions. To integrate with a real backend:

1. Update the `api.ts` file in `src/services/`
2. Replace mock functions with actual API calls
3. Update authentication logic in `AuthContext.tsx`
4. Modify data structures as needed

## Customization

### Styling

- **Colors**: Modify the primary color palette in `tailwind.config.js`
- **Components**: Update component styles in `src/index.css`
- **Dark Mode**: Customize dark mode colors in the CSS variables

### Adding New Features

1. **New Pages**: Create components in the `pages/` directory
2. **New Components**: Add reusable components in `components/`
3. **New Routes**: Update routing in `App.tsx`
4. **New Types**: Add TypeScript interfaces in `types/index.ts`

## Best Practices

### Code Organization
- Use TypeScript for type safety
- Follow React functional component patterns
- Implement proper error handling
- Use React Context for global state

### Performance
- Implement proper loading states
- Use React.memo for expensive components
- Optimize re-renders with useMemo and useCallback
- Lazy load components when appropriate

### Security
- Validate all user inputs
- Implement proper authentication checks
- Use role-based access control
- Sanitize data before rendering

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License.

## Support

For support and questions, please open an issue in the repository or contact the development team. 