import React from 'react';
import { useQuery, useMutation, useQueryClient } from 'react-query';
import apiClient from '../lib/api';
import { 
  Holding, 
  PortfolioResponse, 
  CurrentPriceResponse,
  QUERY_KEYS 
} from '../types/portfolio';

// Fetch all portfolio holdings
export const usePortfolioHoldings = () => {
  return useQuery<PortfolioResponse, Error>(
    QUERY_KEYS.PORTFOLIO_HOLDINGS,
    async () => {
      const { data } = await apiClient.get('/portfolio');
      return data;
    },
    {
      staleTime: 5 * 60 * 1000, // 5 minutes cache
    }
  );
};

// Update a holding
export const useUpdateHolding = () => {
  const queryClient = useQueryClient();
  
  return useMutation(
    async (holding: Holding) => {
      const { data } = await apiClient.put(
        `/portfolio/holdings/${holding.id}`, 
        holding
      );
      return data;
    },
    {
      onSuccess: () => {
        // Invalidate both holdings and summary queries
        queryClient.invalidateQueries(QUERY_KEYS.PORTFOLIO_HOLDINGS);
        queryClient.invalidateQueries(QUERY_KEYS.PORTFOLIO_SUMMARY);
      },
    }
  );
};

// Delete a holding
export const useDeleteHolding = () => {
  const queryClient = useQueryClient();
  
  return useMutation(
    async (id: string) => {
      await apiClient.delete(`/portfolio/holdings/${id}`);
      return id;
    },
    {
      onSuccess: (id) => {
        // Optimistically remove from cache
        queryClient.setQueryData<PortfolioResponse | undefined>(
          QUERY_KEYS.PORTFOLIO_HOLDINGS,
          (old) => {
            if (!old) return undefined;
            return {
              ...old,
              holdings: old.holdings.filter(h => h.id !== id),
              total_holdings: old.total_holdings - 1
            };
          }
        );
        // Invalidate summary
        queryClient.invalidateQueries(QUERY_KEYS.PORTFOLIO_SUMMARY);
      },
    }
  );
};

// Fetch current price for a symbol with WebSocket updates
export const useCurrentPrice = (symbol: string) => {
  const queryClient = useQueryClient();
  
  const query = useQuery<CurrentPriceResponse, Error>(
    QUERY_KEYS.CURRENT_PRICE(symbol),
    async () => {
      const { data } = await apiClient.get(`/market/prices/${symbol}`);
      return data;
    },
    {
      staleTime: 60 * 1000, // 1 minute cache
      enabled: !!symbol, // Only fetch if symbol exists
    }
  );

  // Setup WebSocket for real-time updates
  React.useEffect(() => {
    if (!symbol) return;

    const ws = new WebSocket(`wss://api.example.com/market/prices/${symbol}/ws`);

    ws.onmessage = (event) => {
      const data = JSON.parse(event.data) as CurrentPriceResponse;
      queryClient.setQueryData(QUERY_KEYS.CURRENT_PRICE(symbol), data);
    };

    ws.onerror = (error) => {
      console.error('WebSocket error:', error);
    };

    return () => {
      ws.close();
    };
  }, [symbol, queryClient]);

  return {
    ...query,
    isStreaming: !query.isStale && !query.isError,
  };
};
