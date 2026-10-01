import { apiClient } from './apiClient';

export interface NotificationItem {
  id: string;
  message: string;
  timestamp: string;
  read: boolean;
  type: 'info' | 'success' | 'warning';
  title: string;
}

export interface NotificationPage {
  items: NotificationItem[];
  next_cursor: string;
}

export const notificationsService = {
  list: () => apiClient.get<NotificationPage>('/notifications'),
  markAllRead: () => apiClient.post('/notifications/read'),
};
