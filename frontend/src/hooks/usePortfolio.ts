import React from 'react';

import { useQuery, useMutation, useQueryClient } from 'react-query';

import apiClient from '../lib/api';

import {

Holding,

PortfolioResponse,

CurrentPriceResponse,

QUERY_KEYS

} from '../types/portfolio';

import { NOTIFICATIONS_QUERY_KEYS } from './useNotifications';

  

// Helper function to create notifications

const createNotification = async (notification: {

title: string;

message: string;

type: 'success' | 'info' | 'warning' | 'error' | 'portfolio';

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

  

// Fetch all portfolio holdings

export const usePortfolioHoldings = () => {

return useQuery<PortfolioResponse, Error>(

QUERY_KEYS.PORTFOLIO_HOLDINGS,

async () => {

const { data } = await apiClient.get('/portfolio/');

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

onSuccess: async (data, variables) => {

// Create success notification

await createNotification({

title: 'Holding Updated',

message: `Updated holding for ${variables.symbol}`,

type: 'portfolio',

priority: 'medium',

metadata: {

symbol: variables.symbol,

holding_id: variables.id

}

});

  

// Invalidate both holdings and summary queries

queryClient.invalidateQueries(QUERY_KEYS.PORTFOLIO_HOLDINGS);

queryClient.invalidateQueries(QUERY_KEYS.PORTFOLIO_SUMMARY);

// Invalidate notifications to show the new one

queryClient.invalidateQueries(NOTIFICATIONS_QUERY_KEYS.NOTIFICATIONS);

queryClient.invalidateQueries(NOTIFICATIONS_QUERY_KEYS.UNREAD_COUNT);

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

onSuccess: async (id) => {

// Create success notification

await createNotification({

title: 'Holding Deleted',

message: `Holding has been successfully removed from portfolio`,

type: 'portfolio',

priority: 'medium',

metadata: {

holding_id: id

}

});

  

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

// Invalidate notifications to show the new one

queryClient.invalidateQueries(NOTIFICATIONS_QUERY_KEYS.NOTIFICATIONS);

queryClient.invalidateQueries(NOTIFICATIONS_QUERY_KEYS.UNREAD_COUNT);

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

  

const baseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';

const wsUrl = baseUrl

.replace('http://', 'ws://')

.replace('https://', 'wss://')

+ '/ws'; // Append /ws to the full API path

  

const ws = new WebSocket(wsUrl);

  

ws.onopen = () => {

// Subscribe to price updates for this symbol

ws.send(JSON.stringify({

type: 'subscribe_symbol',

symbol: symbol

}));

};

  

ws.onmessage = (event) => {

try {

const message = JSON.parse(event.data);

if (message.type === 'price_update' && message.data.symbol === symbol) {

queryClient.setQueryData(QUERY_KEYS.CURRENT_PRICE(symbol), message.data);

}

} catch (error) {

console.error('Error parsing WebSocket message:', error);

}

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
