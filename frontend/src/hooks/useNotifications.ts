import { useQuery, useMutation, useQueryClient } from 'react-query';
import apiClient from '@/lib/api';

// Query keys for notifications
export const NOTIFICATIONS_QUERY_KEYS = {
  NOTIFICATIONS: 'notifications',
  NOTIFICATION: 'notification',
  UNREAD_COUNT: 'unreadNotificationCount',
} as const;

// Types based on backend API
export interface Notification {
  id: number;
  user_id: number;
  title: string;
  message: string;
  type: 'info' | 'success' | 'warning' | 'error' | 'market' | 'portfolio' | 'transaction';
  priority: 'low' | 'medium' | 'high' | 'urgent';
  is_read: boolean;
  action_url?: string;
  metadata?: {
    symbol?: string;
    transaction_id?: number;
    portfolio_change?: number;
    price_change?: number;
  };
  created_at: string;
  read_at?: string;
}

export interface CreateNotificationRequest {
  title: string;
  message: string;
  type: 'info' | 'success' | 'warning' | 'error' | 'market' | 'portfolio' | 'transaction';
  priority?: 'low' | 'medium' | 'high' | 'urgent';
  action_url?: string;
  metadata?: Record<string, any>;
}

export interface NotificationsResponse {
  notifications: Notification[];
  total: number;
  unread_count: number;
}

// Hook to fetch all notifications
export const useNotifications = (limit = 50, offset = 0) => {
  return useQuery<NotificationsResponse, Error>(
    [NOTIFICATIONS_QUERY_KEYS.NOTIFICATIONS, limit, offset],
    async () => {
      const { data } = await apiClient.get(`/notifications/?limit=${limit}&offset=${offset}`);
      return data;
    },
    {
      staleTime: 1 * 60 * 1000, // 1 minute cache
      retry: 2,
      refetchOnWindowFocus: true,
      refetchInterval: 2 * 60 * 1000, // Auto-refresh every 2 minutes
    }
  );
};

// Hook to fetch unread notification count
export const useUnreadNotificationCount = () => {
  return useQuery<{ count: number }, Error>(
    NOTIFICATIONS_QUERY_KEYS.UNREAD_COUNT,
    async () => {
      const { data } = await apiClient.get('/notifications/unread-count');
      return data;
    },
    {
      staleTime: 30 * 1000, // 30 seconds cache
      retry: 2,
      refetchOnWindowFocus: true,
      refetchInterval: 1 * 60 * 1000, // Auto-refresh every minute
    }
  );
};

// Hook to mark a notification as read
export const useMarkNotificationAsRead = () => {
  const queryClient = useQueryClient();
  
  return useMutation<Notification, Error, number>(
    async (notificationId: number) => {
      const { data } = await apiClient.put(`/notifications/${notificationId}/read`);
      return data;
    },
    {
      onSuccess: (data) => {
        // Update the specific notification in cache
        queryClient.setQueryData([NOTIFICATIONS_QUERY_KEYS.NOTIFICATION, data.id], data);
        
        // Invalidate notifications list to update read status
        queryClient.invalidateQueries(NOTIFICATIONS_QUERY_KEYS.NOTIFICATIONS);
        
        // Invalidate unread count
        queryClient.invalidateQueries(NOTIFICATIONS_QUERY_KEYS.UNREAD_COUNT);
      },
    }
  );
};

// Hook to mark all notifications as read
export const useMarkAllNotificationsAsRead = () => {
  const queryClient = useQueryClient();
  
  return useMutation<void, Error>(
    async () => {
      await apiClient.put('/notifications/mark-all-read');
    },
    {
      onSuccess: () => {
        // Invalidate all notification queries
        queryClient.invalidateQueries(NOTIFICATIONS_QUERY_KEYS.NOTIFICATIONS);
        queryClient.invalidateQueries(NOTIFICATIONS_QUERY_KEYS.UNREAD_COUNT);
      },
    }
  );
};

// Hook to delete a notification
export const useDeleteNotification = () => {
  const queryClient = useQueryClient();
  
  return useMutation<void, Error, number>(
    async (notificationId: number) => {
      await apiClient.delete(`/notifications/${notificationId}`);
    },
    {
      onSuccess: () => {
        // Invalidate notifications list
        queryClient.invalidateQueries(NOTIFICATIONS_QUERY_KEYS.NOTIFICATIONS);
        queryClient.invalidateQueries(NOTIFICATIONS_QUERY_KEYS.UNREAD_COUNT);
      },
    }
  );
};

// Combined hook for notification management
export const useNotificationManagement = () => {
  const markAsRead = useMarkNotificationAsRead();
  const markAllAsRead = useMarkAllNotificationsAsRead();
  const deleteNotification = useDeleteNotification();
  
  return {
    markAsRead,
    markAllAsRead,
    deleteNotification,
    isLoading: markAsRead.isLoading || markAllAsRead.isLoading || deleteNotification.isLoading,
    error: markAsRead.error || markAllAsRead.error || deleteNotification.error,
  };
};
