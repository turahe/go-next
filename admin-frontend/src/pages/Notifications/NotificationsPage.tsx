import React, { useState } from 'react';
import { useNotifications } from '@/context/NotificationContext';
import { notificationApi } from '@/services/notificationApi';
import { useAuth } from '@/context/AuthContext';

export const NotificationsPage: React.FC = () => {
  const { notifications, unreadCount, isConnected, fetchNotifications, fetchUnreadCount } = useNotifications();
  const { state: authState } = useAuth();
  const [isLoading, setIsLoading] = useState(false);

  const handleTestNotification = async (type: 'success' | 'error' | 'warning' | 'info') => {
    if (!authState.user?.id) return;

    setIsLoading(true);
    try {
      await notificationApi.createNotification({
        user_id: authState.user.id.toString(),
        type,
        title: `${type.charAt(0).toUpperCase() + type.slice(1)} Notification`,
        message: `This is a test ${type} notification sent at ${new Date().toLocaleTimeString()}`,
        data: JSON.stringify({ 
          test: true, 
          timestamp: new Date().toISOString(),
          type 
        })
      });
    } catch (error) {
      console.error('Failed to create test notification:', error);
    } finally {
      setIsLoading(false);
    }
  };

  const handleRefresh = async () => {
    setIsLoading(true);
    try {
      await Promise.all([fetchNotifications(), fetchUnreadCount()]);
    } catch (error) {
      console.error('Failed to refresh notifications:', error);
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="p-6">
      <div className="mb-6">
        <h1 className="text-2xl font-bold text-gray-900 dark:text-white mb-2">
          Real-time Notifications
        </h1>
        <p className="text-gray-600 dark:text-gray-400">
          Test the real-time notification system. Notifications will appear as toasts and in the dropdown.
        </p>
      </div>

      {/* Connection Status */}
      <div className="mb-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-2">
          Connection Status
        </h2>
        <div className="flex items-center space-x-4">
          <div className="flex items-center space-x-2">
            <span className={`w-3 h-3 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`}></span>
            <span className="text-sm text-gray-600 dark:text-gray-400">
              WebSocket: {isConnected ? 'Connected' : 'Disconnected'}
            </span>
          </div>
          <div className="flex items-center space-x-2">
            <span className="text-sm text-gray-600 dark:text-gray-400">
              Unread: {unreadCount}
            </span>
          </div>
          <div className="flex items-center space-x-2">
            <span className="text-sm text-gray-600 dark:text-gray-400">
              Total: {notifications.length}
            </span>
          </div>
        </div>
      </div>

      {/* Test Controls */}
      <div className="mb-6 p-4 bg-gray-50 dark:bg-gray-800 rounded-lg">
        <h2 className="text-lg font-semibold text-gray-900 dark:text-white mb-4">
          Test Notifications
        </h2>
        <div className="flex flex-wrap gap-3">
          <button
            onClick={() => handleTestNotification('success')}
            disabled={isLoading}
            className="px-4 py-2 bg-green-600 text-white rounded-lg hover:bg-green-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Send Success
          </button>
          <button
            onClick={() => handleTestNotification('error')}
            disabled={isLoading}
            className="px-4 py-2 bg-red-600 text-white rounded-lg hover:bg-red-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Send Error
          </button>
          <button
            onClick={() => handleTestNotification('warning')}
            disabled={isLoading}
            className="px-4 py-2 bg-yellow-600 text-white rounded-lg hover:bg-yellow-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Send Warning
          </button>
          <button
            onClick={() => handleTestNotification('info')}
            disabled={isLoading}
            className="px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Send Info
          </button>
          <button
            onClick={handleRefresh}
            disabled={isLoading}
            className="px-4 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 disabled:opacity-50 disabled:cursor-not-allowed"
          >
            Refresh
          </button>
        </div>
      </div>

      {/* Notifications List */}
      <div className="bg-white dark:bg-gray-800 rounded-lg shadow">
        <div className="p-4 border-b border-gray-200 dark:border-gray-700">
          <h2 className="text-lg font-semibold text-gray-900 dark:text-white">
            Recent Notifications
          </h2>
        </div>
        <div className="divide-y divide-gray-200 dark:divide-gray-700">
          {notifications.length === 0 ? (
            <div className="p-4 text-center text-gray-500 dark:text-gray-400">
              No notifications yet
            </div>
          ) : (
            notifications.slice(0, 10).map((notification) => (
              <div
                key={notification.id}
                className={`p-4 ${!notification.read ? 'bg-blue-50 dark:bg-blue-900/20' : ''}`}
              >
                <div className="flex items-start space-x-3">
                  <span className="text-lg">
                    {notification.type === 'success' && '✅'}
                    {notification.type === 'error' && '❌'}
                    {notification.type === 'warning' && '⚠️'}
                    {notification.type === 'info' && 'ℹ️'}
                  </span>
                  <div className="flex-1">
                    <h3 className="text-sm font-medium text-gray-900 dark:text-white">
                      {notification.title}
                    </h3>
                    <p className="text-sm text-gray-600 dark:text-gray-300 mt-1">
                      {notification.message}
                    </p>
                    <p className="text-xs text-gray-500 dark:text-gray-400 mt-2">
                      {new Date(notification.timestamp).toLocaleString()}
                      {!notification.read && (
                        <span className="ml-2 inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200">
                          New
                        </span>
                      )}
                    </p>
                  </div>
                </div>
              </div>
            ))
          )}
        </div>
      </div>
    </div>
  );
}; 