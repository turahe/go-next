import { api } from './api';
import { Notification } from '../types';

export interface NotificationResponse {
  data: Notification[];
  message: string;
  success: boolean;
}

export interface NotificationCountResponse {
  data: number;
  message: string;
  success: boolean;
}

export class NotificationApiService {
  private baseUrl = '/api/v1/notifications';

  async getNotifications(limit: number = 50, offset: number = 0): Promise<Notification[]> {
    try {
      const response = await api.request<NotificationResponse>(`${this.baseUrl}?limit=${limit}&offset=${offset}`, {
        method: 'GET',
      });
      return response.data || [];
    } catch (error) {
      console.error('Failed to fetch notifications:', error);
      throw error;
    }
  }

  async getUnreadCount(): Promise<number> {
    try {
      const response = await api.request<NotificationCountResponse>(`${this.baseUrl}/unread-count`, {
        method: 'GET',
      });
      return response.data;
    } catch (error) {
      console.error('Failed to fetch unread count:', error);
      throw error;
    }
  }

  async markAsRead(notificationId: string): Promise<void> {
    try {
      await api.request(`${this.baseUrl}/${notificationId}/read`, {
        method: 'PUT',
      });
    } catch (error) {
      console.error('Failed to mark notification as read:', error);
      throw error;
    }
  }

  async markAllAsRead(): Promise<void> {
    try {
      await api.request(`${this.baseUrl}/mark-all-read`, {
        method: 'PUT',
      });
    } catch (error) {
      console.error('Failed to mark all notifications as read:', error);
      throw error;
    }
  }

  async deleteNotification(notificationId: string): Promise<void> {
    try {
      await api.request(`${this.baseUrl}/${notificationId}`, {
        method: 'DELETE',
      });
    } catch (error) {
      console.error('Failed to delete notification:', error);
      throw error;
    }
  }

  async deleteAllNotifications(): Promise<void> {
    try {
      await api.request(`${this.baseUrl}`, {
        method: 'DELETE',
      });
    } catch (error) {
      console.error('Failed to delete all notifications:', error);
      throw error;
    }
  }

  async createNotification(notification: {
    user_id: string;
    type: 'success' | 'error' | 'warning' | 'info';
    title: string;
    message: string;
    data?: string;
  }): Promise<Notification> {
    try {
      const response = await api.request<{ data: Notification; message: string; success: boolean }>('/api/v1/admin/notifications', {
        method: 'POST',
        body: JSON.stringify(notification),
      });
      return response.data;
    } catch (error) {
      console.error('Failed to create notification:', error);
      throw error;
    }
  }
}

export const notificationApi = new NotificationApiService();
export default notificationApi; 