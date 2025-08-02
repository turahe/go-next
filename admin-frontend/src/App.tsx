
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { AuthProvider } from './context/AuthContext';
import { NotificationProvider } from './context/NotificationContext';
import { ThemeProvider } from './context/ThemeContext';
import { SidebarProvider } from './context/SidebarContext';
import { NotificationToastContainer } from './components/Notifications/NotificationToast';

// Auth Pages
import { LoginPage } from './pages/Auth/LoginPage';
import { RegisterPage } from './pages/Auth/RegisterPage';

// Layout
import AppLayout from './layout/AppLayout';

// Pages

import { UsersPage } from './pages/Users/UsersPage';
import { PostsPage } from './pages/Posts/PostsPage';
import { NotificationsPage } from './pages/Notifications/NotificationsPage';

// Protected Route Component
import { ProtectedRoute } from './components/Auth/ProtectedRoute';

// TailAdmin Pages
import Home from './pages/Dashboard/Home';
import Calendar from './pages/Calendar';
import UserProfiles from './pages/UserProfiles';
import FormElements from './pages/Forms/FormElements';
import BasicTables from './pages/Tables/BasicTables';
import Blank from './pages/Blank';
import NotFound from './pages/OtherPage/NotFound';
import LineChart from './pages/Charts/LineChart';
import BarChart from './pages/Charts/BarChart';
import Alerts from './pages/UiElements/Alerts';
import Avatars from './pages/UiElements/Avatars';
import Badges from './pages/UiElements/Badges';
import Buttons from './pages/UiElements/Buttons';
import Images from './pages/UiElements/Images';
import Videos from './pages/UiElements/Videos';
import SignIn from './pages/AuthPages/SignIn';
import SignUp from './pages/AuthPages/SignUp';

function App() {
  return (
    <ThemeProvider>
      <AuthProvider>
        <NotificationProvider>
          <SidebarProvider>
            <Router>
              <div className="App">
                <NotificationToastContainer />
                <Routes>
                  {/* Public Routes */}
                  <Route path="/login" element={<LoginPage />} />
                  <Route path="/register" element={<RegisterPage />} />
                  <Route path="/signin" element={<SignIn />} />
                  <Route path="/signup" element={<SignUp />} />

                  {/* Protected Routes with TailAdmin Layout */}
                  <Route path="/" element={<AppLayout />}>
                    <Route index element={<Navigate to="/dashboard" replace />} />
                    <Route path="dashboard" element={<ProtectedRoute><Home /></ProtectedRoute>} />
                    <Route path="users" element={<ProtectedRoute><UsersPage /></ProtectedRoute>} />
                    <Route path="posts" element={<ProtectedRoute><PostsPage /></ProtectedRoute>} />
                    <Route path="notifications" element={<ProtectedRoute><NotificationsPage /></ProtectedRoute>} />
                    
                    {/* TailAdmin Demo Pages */}
                    <Route path="calendar" element={<ProtectedRoute><Calendar /></ProtectedRoute>} />
                    <Route path="profile" element={<ProtectedRoute><UserProfiles /></ProtectedRoute>} />
                    <Route path="form-elements" element={<ProtectedRoute><FormElements /></ProtectedRoute>} />
                    <Route path="basic-tables" element={<ProtectedRoute><BasicTables /></ProtectedRoute>} />
                    <Route path="blank" element={<ProtectedRoute><Blank /></ProtectedRoute>} />
                    <Route path="line-chart" element={<ProtectedRoute><LineChart /></ProtectedRoute>} />
                    <Route path="bar-chart" element={<ProtectedRoute><BarChart /></ProtectedRoute>} />
                    <Route path="alerts" element={<ProtectedRoute><Alerts /></ProtectedRoute>} />
                    <Route path="avatars" element={<ProtectedRoute><Avatars /></ProtectedRoute>} />
                    <Route path="badge" element={<ProtectedRoute><Badges /></ProtectedRoute>} />
                    <Route path="buttons" element={<ProtectedRoute><Buttons /></ProtectedRoute>} />
                    <Route path="images" element={<ProtectedRoute><Images /></ProtectedRoute>} />
                    <Route path="videos" element={<ProtectedRoute><Videos /></ProtectedRoute>} />
                  </Route>

                  {/* Error Pages */}
                  <Route path="/error-404" element={<NotFound />} />
                  <Route path="*" element={<Navigate to="/error-404" replace />} />
                </Routes>
              </div>
            </Router>
          </SidebarProvider>
        </NotificationProvider>
      </AuthProvider>
    </ThemeProvider>
  );
}

export default App; 