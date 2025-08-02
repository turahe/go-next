import React, { createContext, useContext, useReducer, ReactNode, useEffect, useCallback } from 'react';
import { Notification, WebSocketMessage } from '../types';
import { websocketService } from '../services/websocket';
import { notificationApi } from '../services/notificationApi';

interface NotificationState {
  notifications: Notification[];
  unreadCount: number;
}

type NotificationAction =
  | { type: 'ADD_NOTIFICATION'; payload: Omit<Notification, 'id' | 'timestamp' | 'read'> }
  | { type: 'ADD_NOTIFICATION_FROM_WS'; payload: Notification }
  | { type: 'REMOVE_NOTIFICATION'; payload: string }
  | { type: 'MARK_AS_READ'; payload: string }
  | { type: 'CLEAR_ALL' }
  | { type: 'SET_NOTIFICATIONS'; payload: Notification[] }
  | { type: 'SET_UNREAD_COUNT'; payload: number }
  | { type: 'UPDATE_UNREAD_COUNT'; payload: number };

const initialState: NotificationState = {
  notifications: [],
  unreadCount: 0,
};

const notificationReducer = (state: NotificationState, action: NotificationAction): NotificationState => {
  switch (action.type) {
    case 'ADD_NOTIFICATION':
      const newNotification: Notification = {
        ...action.payload,
        id: Date.now().toString(),
        timestamp: new Date().toISOString(),
        read: false,
      };
      return {
        ...state,
        notifications: [newNotification, ...state.notifications],
        unreadCount: state.unreadCount + 1,
      };
    case 'ADD_NOTIFICATION_FROM_WS':
      return {
        ...state,
        notifications: [action.payload, ...state.notifications],
        unreadCount: state.unreadCount + 1,
      };
    case 'REMOVE_NOTIFICATION':
      const removedNotification = state.notifications.find(n => n.id === action.payload);
      return {
        ...state,
        notifications: state.notifications.filter(n => n.id !== action.payload),
        unreadCount: removedNotification && !removedNotification.read 
          ? Math.max(0, state.unreadCount - 1) 
          : state.unreadCount,
      };
    case 'MARK_AS_READ':
      const wasUnread = state.notifications.find(n => n.id === action.payload)?.read === false;
      return {
        ...state,
        notifications: state.notifications.map(n =>
          n.id === action.payload ? { ...n, read: true } : n
        ),
        unreadCount: wasUnread ? Math.max(0, state.unreadCount - 1) : state.unreadCount,
      };
    case 'CLEAR_ALL':
      return {
        ...state,
        notifications: [],
        unreadCount: 0,
      };
    case 'SET_NOTIFICATIONS':
      return {
        ...state,
        notifications: action.payload,
        unreadCount: action.payload.filter(n => !n.read).length,
      };
    case 'SET_UNREAD_COUNT':
      return {
        ...state,
        unreadCount: action.payload,
      };
    case 'UPDATE_UNREAD_COUNT':
      return {
        ...state,
        unreadCount: Math.max(0, state.unreadCount + action.payload),
      };
    default:
      return state;
  }
};

interface NotificationContextType extends NotificationState {
  addNotification: (notification: Omit<Notification, 'id' | 'timestamp' | 'read'>) => void;
  removeNotification: (id: string) => void;
  markAsRead: (id: string) => void;
  markAllAsRead: () => void;
  clearAll: () => void;
  fetchNotifications: () => Promise<void>;
  fetchUnreadCount: () => Promise<void>;
  deleteNotification: (id: string) => Promise<void>;
  deleteAllNotifications: () => Promise<void>;
  isConnected: boolean;
}

const NotificationContext = createContext<NotificationContextType | undefined>(undefined);

export const useNotifications = () => {
  const context = useContext(NotificationContext);
  if (context === undefined) {
    throw new Error('useNotifications must be used within a NotificationProvider');
  }
  return context;
};

interface NotificationProviderProps {
  children: ReactNode;
}

export const NotificationProvider: React.FC<NotificationProviderProps> = ({ children }) => {
  const [state, dispatch] = useReducer(notificationReducer, initialState);

  const addNotification = (notification: Omit<Notification, 'id' | 'timestamp' | 'read'>) => {
    dispatch({ type: 'ADD_NOTIFICATION', payload: notification });
  };

  const removeNotification = (id: string) => {
    dispatch({ type: 'REMOVE_NOTIFICATION', payload: id });
  };

  const markAsRead = (id: string) => {
    dispatch({ type: 'MARK_AS_READ', payload: id });
  };

  const markAllAsRead = async () => {
    try {
      await notificationApi.markAllAsRead();
      dispatch({ type: 'SET_UNREAD_COUNT', payload: 0 });
      // Update all notifications to read
      dispatch({ 
        type: 'SET_NOTIFICATIONS', 
        payload: state.notifications.map(n => ({ ...n, read: true }))
      });
    } catch (error) {
      console.error('Failed to mark all as read:', error);
    }
  };

  const clearAll = () => {
    dispatch({ type: 'CLEAR_ALL' });
  };

  const fetchNotifications = async () => {
    try {
      const notifications = await notificationApi.getNotifications();
      dispatch({ type: 'SET_NOTIFICATIONS', payload: notifications });
    } catch (error) {
      console.error('Failed to fetch notifications:', error);
    }
  };

  const fetchUnreadCount = async () => {
    try {
      const count = await notificationApi.getUnreadCount();
      dispatch({ type: 'SET_UNREAD_COUNT', payload: count });
    } catch (error) {
      console.error('Failed to fetch unread count:', error);
    }
  };

  const deleteNotification = async (id: string) => {
    try {
      await notificationApi.deleteNotification(id);
      dispatch({ type: 'REMOVE_NOTIFICATION', payload: id });
    } catch (error) {
      console.error('Failed to delete notification:', error);
    }
  };

  const deleteAllNotifications = async () => {
    try {
      await notificationApi.deleteAllNotifications();
      dispatch({ type: 'CLEAR_ALL' });
    } catch (error) {
      console.error('Failed to delete all notifications:', error);
    }
  };

  // WebSocket message handler
  const handleWebSocketMessage = useCallback((message: WebSocketMessage) => {
    if (message.type === 'notification') {
      const notification: Notification = {
        id: message.data.id || Date.now().toString(),
        type: message.data.type,
        title: message.data.title,
        message: message.data.message,
        timestamp: message.data.timestamp || new Date().toISOString(),
        read: false,
        data: message.data.data,
      };
      dispatch({ type: 'ADD_NOTIFICATION_FROM_WS', payload: notification });
    }
  }, []);

  // Setup WebSocket connection and message handlers
  useEffect(() => {
    websocketService.onMessage('notification', handleWebSocketMessage);
    
    // Connect to WebSocket
    websocketService.connect();

    // Fetch initial data
    fetchNotifications();
    fetchUnreadCount();

    return () => {
      websocketService.offMessage('notification', handleWebSocketMessage);
    };
  }, [handleWebSocketMessage]);

  const value: NotificationContextType = {
    ...state,
    addNotification,
    removeNotification,
    markAsRead,
    markAllAsRead,
    clearAll,
    fetchNotifications,
    fetchUnreadCount,
    deleteNotification,
    deleteAllNotifications,
    isConnected: websocketService.isConnected(),
  };

  return (
    <NotificationContext.Provider value={value}>
      {children}
    </NotificationContext.Provider>
  );
}; 