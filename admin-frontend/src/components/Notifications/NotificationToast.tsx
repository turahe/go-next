import React, { useEffect, useState } from 'react';
import { useNotifications } from '@/context/NotificationContext';
import { Notification } from '@/types';

interface ToastProps {
  notification: Notification;
  onClose: () => void;
  duration?: number;
}

const NotificationToast: React.FC<ToastProps> = ({ notification, onClose, duration = 5000 }) => {
  const [isVisible, setIsVisible] = useState(true);
  const [progress, setProgress] = useState(100);

  useEffect(() => {
    const timer = setTimeout(() => {
      setIsVisible(false);
      setTimeout(onClose, 300); // Wait for fade out animation
    }, duration);

    const progressTimer = setInterval(() => {
      setProgress(prev => Math.max(0, prev - (100 / (duration / 100))));
    }, 100);

    return () => {
      clearTimeout(timer);
      clearInterval(progressTimer);
    };
  }, [duration, onClose]);

  const getNotificationStyles = (type: string) => {
    switch (type) {
      case 'success':
        return 'bg-green-50 border-green-200 text-green-800 dark:bg-green-900/20 dark:border-green-700 dark:text-green-300';
      case 'error':
        return 'bg-red-50 border-red-200 text-red-800 dark:bg-red-900/20 dark:border-red-700 dark:text-red-300';
      case 'warning':
        return 'bg-yellow-50 border-yellow-200 text-yellow-800 dark:bg-yellow-900/20 dark:border-yellow-700 dark:text-yellow-300';
      case 'info':
        return 'bg-blue-50 border-blue-200 text-blue-800 dark:bg-blue-900/20 dark:border-blue-700 dark:text-blue-300';
      default:
        return 'bg-gray-50 border-gray-200 text-gray-800 dark:bg-gray-900/20 dark:border-gray-700 dark:text-gray-300';
    }
  };

  const getNotificationIcon = (type: string) => {
    switch (type) {
      case 'success':
        return '✅';
      case 'error':
        return '❌';
      case 'warning':
        return '⚠️';
      case 'info':
        return 'ℹ️';
      default:
        return '📢';
    }
  };

  return (
    <div
      className={`fixed top-4 right-4 w-80 max-w-sm transform transition-all duration-300 ease-in-out z-50 ${
        isVisible ? 'translate-x-0 opacity-100' : 'translate-x-full opacity-0'
      }`}
    >
      <div className={`relative p-4 rounded-lg border shadow-lg ${getNotificationStyles(notification.type)}`}>
        {/* Progress bar */}
        <div className="absolute top-0 left-0 w-full h-1 bg-gray-200 dark:bg-gray-700 rounded-t-lg overflow-hidden">
          <div
            className="h-full bg-current opacity-30 transition-all duration-100 ease-linear"
            style={{ width: `${progress}%` }}
          />
        </div>

        {/* Close button */}
        <button
          onClick={() => {
            setIsVisible(false);
            setTimeout(onClose, 300);
          }}
          className="absolute top-2 right-2 text-gray-400 hover:text-gray-600 dark:text-gray-500 dark:hover:text-gray-300 transition-colors"
        >
          ✕
        </button>

        {/* Content */}
        <div className="flex items-start space-x-3 pr-6">
          <span className="text-lg flex-shrink-0">{getNotificationIcon(notification.type)}</span>
          <div className="flex-1 min-w-0">
            <h4 className="text-sm font-medium mb-1">{notification.title}</h4>
            <p className="text-sm opacity-90">{notification.message}</p>
            {notification.data && (
              <div className="mt-2 text-xs opacity-75">
                <details>
                  <summary className="cursor-pointer hover:opacity-100">Additional data</summary>
                  <pre className="mt-1 whitespace-pre-wrap">{notification.data}</pre>
                </details>
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export const NotificationToastContainer: React.FC = () => {
  const { notifications } = useNotifications();
  const [toasts, setToasts] = useState<Notification[]>([]);

  useEffect(() => {
    // Only show toasts for new, unread notifications
    const newNotifications = notifications.filter(n => !n.read);
    setToasts(prev => {
      const existingIds = new Set(prev.map(t => t.id));
      const newToasts = newNotifications.filter(n => !existingIds.has(n.id));
      return [...prev, ...newToasts];
    });
  }, [notifications]);

  const removeToast = (id: string) => {
    setToasts(prev => prev.filter(t => t.id !== id));
  };

  return (
    <div className="fixed top-4 right-4 z-50 space-y-2">
      {toasts.map((notification, index) => (
        <div
          key={notification.id}
          style={{ transform: `translateY(${index * 20}px)` }}
        >
          <NotificationToast
            notification={notification}
            onClose={() => removeToast(notification.id)}
            duration={6000}
          />
        </div>
      ))}
    </div>
  );
};

export default NotificationToast; 