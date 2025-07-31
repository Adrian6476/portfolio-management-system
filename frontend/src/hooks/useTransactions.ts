import { useQuery, useMutation, useQueryClient } from 'react-query';
import apiClient from '@/lib/api';
import { QUERY_KEYS } from '@/types/portfolio';
import { ANALYTICS_QUERY_KEYS } from '@/hooks/useAnalytics';
import { NOTIFICATIONS_QUERY_KEYS } from '@/hooks/useNotifications';

// Query keys for transactions
export const TRANSACTIONS_QUERY_KEYS = {
  TRANSACTIONS: 'transactions',
  TRANSACTION: 'transaction',
} as const;

// Types based on backend API
export interface Transaction {
  id: number;
  user_id: number;
  symbol: string;
  transaction_type: 'BUY' | 'SELL';
  quantity: number;
  price: number;
  fees?: number;
  transaction_date: string;
  notes?: string;
  created_at: string;
  updated_at: string;
}

export interface CreateTransactionRequest {
  symbol: string;
  transaction_type: 'BUY' | 'SELL';
  quantity: number;
  price: number;
  fees?: number;
  notes?: string;
}

export interface UpdateTransactionRequest extends CreateTransactionRequest {
  id: number;
}

// Helper function to create notifications
const createNotification = async (notification: {
  title: string;
  message: string;
  type: 'success' | 'info' | 'warning' | 'error' | 'transaction';
  priority?: 'low' | 'medium' | 'high';
  metadata?: Record<string, any>;
}) => {
  try {
    await apiClient.post('/notifications/', notification);
  } catch (error) {
    console.warn('Failed to create notification:', error);
    // Don't fail the main operation if notification creation fails
  }
};

// Hook to fetch all transactions
export const useTransactions = () => {
  return useQuery<Transaction[], Error>(
    TRANSACTIONS_QUERY_KEYS.TRANSACTIONS,
    async () => {
      const { data } = await apiClient.get('/transactions/');
      return data;
    },
    {
      staleTime: 2 * 60 * 1000, // 2 minutes cache
      retry: 2,
      refetchOnWindowFocus: false,
    }
  );
};

// Hook to fetch a specific transaction
export const useTransaction = (id: number) => {
  return useQuery<Transaction, Error>(
    [TRANSACTIONS_QUERY_KEYS.TRANSACTION, id],
    async () => {
      const { data } = await apiClient.get(`/transactions/${id}`);
      return data;
    },
    {
      enabled: !!id,
      staleTime: 5 * 60 * 1000, // 5 minutes cache
      retry: 2,
    }
  );
};

// Hook to create a new transaction
export const useCreateTransaction = () => {
  const queryClient = useQueryClient();
  
  return useMutation<Transaction, Error, CreateTransactionRequest>(
    async (transactionData: CreateTransactionRequest) => {
      const { data } = await apiClient.post('/transactions/', transactionData);
      return data;
    },
    {
      onSuccess: async (data, variables) => {
        // Create success notification
        await createNotification({
          title: 'Transaction Added',
          message: `Successfully ${variables.transaction_type.toLowerCase()}ed ${variables.quantity} shares of ${variables.symbol}`,
          type: 'transaction',
          priority: 'medium',
          metadata: {
            symbol: variables.symbol,
            transaction_id: data.id,
            transaction_type: variables.transaction_type,
            quantity: variables.quantity,
            price: variables.price
          }
        });

        // Invalidate and refetch transactions list
        queryClient.invalidateQueries(TRANSACTIONS_QUERY_KEYS.TRANSACTIONS);
        // Also invalidate portfolio data since transactions affect holdings
        queryClient.invalidateQueries(QUERY_KEYS.PORTFOLIO_HOLDINGS);
        queryClient.invalidateQueries(QUERY_KEYS.PORTFOLIO_SUMMARY);
        // Invalidate analytics data
        queryClient.invalidateQueries(ANALYTICS_QUERY_KEYS.RISK_METRICS);
        queryClient.invalidateQueries(ANALYTICS_QUERY_KEYS.PERFORMANCE_ANALYTICS);
        queryClient.invalidateQueries(ANALYTICS_QUERY_KEYS.ASSET_ALLOCATION);
        // Invalidate notifications to show the new one
        queryClient.invalidateQueries(NOTIFICATIONS_QUERY_KEYS.NOTIFICATIONS);
        queryClient.invalidateQueries(NOTIFICATIONS_QUERY_KEYS.UNREAD_COUNT);
      },
    }
  );
};

// Hook to update an existing transaction
export const useUpdateTransaction = () => {
  const queryClient = useQueryClient();
  
  return useMutation<Transaction, Error, UpdateTransactionRequest>(
    async (transactionData: UpdateTransactionRequest) => {
      const { id, ...updateData } = transactionData;
      const { data } = await apiClient.put(`/transactions/${id}`, updateData);
      return data;
    },
    {
      onSuccess: async (data, variables) => {
        // Create success notification
        await createNotification({
          title: 'Transaction Updated',
          message: `Updated ${variables.transaction_type.toLowerCase()} transaction for ${variables.symbol}`,
          type: 'transaction',
          priority: 'medium',
          metadata: {
            symbol: variables.symbol,
            transaction_id: data.id,
            transaction_type: variables.transaction_type
          }
        });

        // Update the specific transaction in cache
        queryClient.setQueryData([TRANSACTIONS_QUERY_KEYS.TRANSACTION, data.id], data);
        // Invalidate transactions list
        queryClient.invalidateQueries(TRANSACTIONS_QUERY_KEYS.TRANSACTIONS);
        // Also invalidate portfolio data
        queryClient.invalidateQueries(QUERY_KEYS.PORTFOLIO_HOLDINGS);
        queryClient.invalidateQueries(QUERY_KEYS.PORTFOLIO_SUMMARY);
        // Invalidate analytics data
        queryClient.invalidateQueries(ANALYTICS_QUERY_KEYS.RISK_METRICS);
        queryClient.invalidateQueries(ANALYTICS_QUERY_KEYS.PERFORMANCE_ANALYTICS);
        queryClient.invalidateQueries(ANALYTICS_QUERY_KEYS.ASSET_ALLOCATION);
        // Invalidate notifications to show the new one
        queryClient.invalidateQueries(NOTIFICATIONS_QUERY_KEYS.NOTIFICATIONS);
        queryClient.invalidateQueries(NOTIFICATIONS_QUERY_KEYS.UNREAD_COUNT);
      },
    }
  );
};

// Hook to delete a transaction
export const useDeleteTransaction = () => {
  const queryClient = useQueryClient();
  
  return useMutation<void, Error, number>(
    async (id: number) => {
      await apiClient.delete(`/transactions/${id}`);
    },
    {
      onSuccess: async (_, transactionId) => {
        // Create success notification
        await createNotification({
          title: 'Transaction Deleted',
          message: `Transaction has been successfully deleted`,
          type: 'transaction',
          priority: 'medium',
          metadata: {
            transaction_id: transactionId
          }
        });

        // Invalidate and refetch transactions list
        queryClient.invalidateQueries(TRANSACTIONS_QUERY_KEYS.TRANSACTIONS);
        // Also invalidate portfolio data
        queryClient.invalidateQueries(QUERY_KEYS.PORTFOLIO_HOLDINGS);
        queryClient.invalidateQueries(QUERY_KEYS.PORTFOLIO_SUMMARY);
        // Invalidate analytics data
        queryClient.invalidateQueries(ANALYTICS_QUERY_KEYS.RISK_METRICS);
        queryClient.invalidateQueries(ANALYTICS_QUERY_KEYS.PERFORMANCE_ANALYTICS);
        queryClient.invalidateQueries(ANALYTICS_QUERY_KEYS.ASSET_ALLOCATION);
        // Invalidate notifications to show the new one
        queryClient.invalidateQueries(NOTIFICATIONS_QUERY_KEYS.NOTIFICATIONS);
        queryClient.invalidateQueries(NOTIFICATIONS_QUERY_KEYS.UNREAD_COUNT);
      },
    }
  );
};

// Combined hook for transaction management
export const useTransactionManagement = () => {
  const createTransaction = useCreateTransaction();
  const updateTransaction = useUpdateTransaction();
  const deleteTransaction = useDeleteTransaction();
  
  return {
    createTransaction,
    updateTransaction,
    deleteTransaction,
    isLoading: createTransaction.isLoading || updateTransaction.isLoading || deleteTransaction.isLoading,
    error: createTransaction.error || updateTransaction.error || deleteTransaction.error,
  };
};
