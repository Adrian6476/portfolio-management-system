import { useEffect, useRef, useState, useCallback } from 'react';
import { useQueryClient } from 'react-query';
import { QUERY_KEYS, PortfolioSummaryResponse, CurrentPriceResponse } from '../types/portfolio';

export interface WebSocketMessage {
  type: 'connected' | 'portfolio_update' | 'price_update' | 'market_update' | 'notification';
  symbol?: string;
  data: any;
  timestamp: number;
}

export interface PortfolioUpdateData {
  total_value: number;
  daily_change: number;
  daily_change_percent: number;
  unrealized_gain_loss: number;
  unrealized_gain_loss_percent: number;
}

export interface PriceUpdateData {
  symbol: string;
  current_price: number;
  change: number;
  change_percent: number;
  high: number;
  low: number;
  volume?: number;
}

export interface UseWebSocketOptions {
  enabled?: boolean;
  reconnectAttempts?: number;
  reconnectInterval?: number;
}

export const useWebSocket = (options: UseWebSocketOptions = {}) => {
  const {
    enabled = true,
    reconnectAttempts = 5,
    reconnectInterval = 3000,
  } = options;

  const queryClient = useQueryClient();
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectAttemptsRef = useRef(0);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);

  const [isConnected, setIsConnected] = useState(false);
  const [connectionStatus, setConnectionStatus] = useState<'connecting' | 'connected' | 'disconnected' | 'error'>('disconnected');

  const getWebSocketUrl = useCallback(() => {
    const baseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080/api/v1';
    const wsUrl = baseUrl
      .replace('http://', 'ws://')
      .replace('https://', 'wss://')
      + '/ws'; // Append /ws to the full API path
    return wsUrl;
  }, []);

  const handleMessage = useCallback((event: MessageEvent) => {
    try {
      const message: WebSocketMessage = JSON.parse(event.data);
      console.log('WebSocket message received:', message);
      
      switch (message.type) {
        case 'connected':
          console.log('WebSocket connected, client ID:', message.data?.client_id);
          break;
          
        case 'portfolio_update':
          // Update portfolio summary data with the new structure
          const portfolioData = message.data as PortfolioUpdateData;
          const portfolioSummary: PortfolioSummaryResponse = {
            summary: {
              total_holdings: 0, // Will be updated by other queries
              total_cost: portfolioData.total_value - portfolioData.unrealized_gain_loss,
              total_shares: 0, // Will be updated by other queries
              total_market_value: portfolioData.total_value,
              daily_change: portfolioData.daily_change,
              daily_change_percent: portfolioData.daily_change_percent,
              unrealized_gain_loss: portfolioData.unrealized_gain_loss,
              unrealized_gain_loss_percent: portfolioData.unrealized_gain_loss_percent,
            },
            asset_allocation: [], // Will be updated by other queries
            top_holdings: [] // Will be updated by other queries
          };
          queryClient.setQueryData(QUERY_KEYS.PORTFOLIO_SUMMARY, portfolioSummary);
          
          // Invalidate portfolio holdings to refresh the data
          queryClient.invalidateQueries(QUERY_KEYS.PORTFOLIO_HOLDINGS);
          break;
          
        case 'price_update':
          // Update individual stock price
          const priceData = message.data as PriceUpdateData;
          if (priceData.symbol) {
            const currentPriceData: CurrentPriceResponse = {
              symbol: priceData.symbol,
              current_price: priceData.current_price,
              change: priceData.change,
              change_percent: priceData.change_percent,
              high: priceData.high,
              low: priceData.low,
              open: priceData.current_price, // Approximation
              previous_close: priceData.current_price - priceData.change,
              timestamp: message.timestamp
            };
            queryClient.setQueryData(QUERY_KEYS.CURRENT_PRICE(priceData.symbol), currentPriceData);
            
            // Also invalidate portfolio holdings to update calculated values
            queryClient.invalidateQueries(QUERY_KEYS.PORTFOLIO_HOLDINGS);
          }
          break;
          
        case 'market_update':
          // Handle general market updates
          console.log('Market update received:', message.data);
          queryClient.invalidateQueries(QUERY_KEYS.PORTFOLIO_HOLDINGS);
          break;
          
        case 'notification':
          // Handle notifications (could be used for alerts, etc.)
          console.log('Notification received:', message.data);
          break;
          
        default:
          console.log('Unknown WebSocket message type:', message.type, message);
      }
    } catch (error) {
      console.error('Error parsing WebSocket message:', error, event.data);
    }
  }, [queryClient]);

  const connect = useCallback(() => {
    if (!enabled || wsRef.current?.readyState === WebSocket.CONNECTING) {
      return;
    }

    try {
      setConnectionStatus('connecting');
      const wsUrl = getWebSocketUrl();
      console.log('Connecting to WebSocket:', wsUrl); // Debug log
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        console.log('WebSocket connected successfully'); // Debug log
        setIsConnected(true);
        setConnectionStatus('connected');
        reconnectAttemptsRef.current = 0;

        // Send subscription message for portfolio updates
        ws.send(JSON.stringify({
          type: 'subscribe',
          data: 'portfolio'
        }));
        
        // Also subscribe to general price updates
        setTimeout(() => {
          ws.send(JSON.stringify({
            type: 'subscribe',
            data: 'prices'
          }));
        }, 100);
      };

      ws.onmessage = handleMessage;

      ws.onclose = (event) => {
        console.log('WebSocket disconnected:', event.code, event.reason); // Debug log
        setIsConnected(false);
        setConnectionStatus('disconnected');
        wsRef.current = null;

        // Attempt to reconnect if not manually closed
        if (event.code !== 1000 && enabled && reconnectAttemptsRef.current < reconnectAttempts) {
          reconnectAttemptsRef.current++;
          console.log(`Attempting to reconnect (${reconnectAttemptsRef.current}/${reconnectAttempts})...`);
          
          reconnectTimeoutRef.current = setTimeout(() => {
            connect();
          }, reconnectInterval);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        setConnectionStatus('error');
      };
    } catch (error) {
      console.error('Failed to create WebSocket connection:', error);
      setConnectionStatus('error');
    }
  }, [enabled, getWebSocketUrl, handleMessage, reconnectAttempts, reconnectInterval]);

  const disconnect = useCallback(() => {
    if (reconnectTimeoutRef.current) {
      clearTimeout(reconnectTimeoutRef.current);
      reconnectTimeoutRef.current = null;
    }

    if (wsRef.current) {
      wsRef.current.close(1000, 'Manual disconnect');
      wsRef.current = null;
    }
    
    setIsConnected(false);
    setConnectionStatus('disconnected');
  }, []);

  const sendMessage = useCallback((message: any) => {
    if (wsRef.current?.readyState === WebSocket.OPEN) {
      wsRef.current.send(JSON.stringify(message));
      return true;
    }
    return false;
  }, []);

  useEffect(() => {
    if (enabled) {
      connect();
    } else {
      disconnect();
    }

    return () => {
      disconnect();
    };
  }, [enabled, connect, disconnect]);

  return {
    isConnected,
    connectionStatus,
    sendMessage,
    connect,
    disconnect,
  };
};

// Hook for portfolio-specific WebSocket updates
export const usePortfolioWebSocket = (enabled = true) => {
  const queryClient = useQueryClient();
  
  const webSocket = useWebSocket({ enabled });

  // Subscribe to specific portfolio symbols when holdings change
  useEffect(() => {
    const portfolioData = queryClient.getQueryData(QUERY_KEYS.PORTFOLIO_HOLDINGS);
    
    if (webSocket.isConnected && portfolioData && (portfolioData as any).holdings) {
      const symbols = (portfolioData as any).holdings.map((holding: any) => holding.symbol);
      
      webSocket.sendMessage({
        type: 'subscribe_symbols',
        symbols: symbols
      });
    }
  }, [webSocket.isConnected, webSocket.sendMessage, queryClient]);

  return webSocket;
};
