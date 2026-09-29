import { apiClient } from './apiClient';

export interface NotificationItem {
  id: string;
  message: string;
  timestamp: string;
  read: boolean;
  type: 'info' | 'success' | 'warning';
  title: string;
}

export const notificationsService = {
  list: () => apiClient.get<NotificationItem[]>('/notifications'),
  markAllRead: () => apiClient.post('/notifications/read'),
};
