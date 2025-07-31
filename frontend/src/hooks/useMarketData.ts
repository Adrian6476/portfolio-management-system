import { useQuery, useMutation, useQueryClient } from 'react-query';
import apiClient from '@/lib/api';

// Query keys for market data
export const MARKET_DATA_QUERY_KEYS = {
  ASSET_SEARCH: 'assetSearch',
  ASSET_DETAILS: 'assetDetails',
  TRENDING_ASSETS: 'trendingAssets',
  MARKET_MOVERS: 'marketMovers',
} as const;

// Types for market data
export interface AssetSearchResult {
  symbol: string;
  name: string;
  type: 'stock' | 'etf' | 'mutual_fund' | 'crypto' | 'bond';
  exchange: string;
  currency: string;
  country: string;
  sector?: string;
  industry?: string;
  market_cap?: number;
  description?: string;
}

export interface AssetDetails {
  symbol: string;
  name: string;
  type: 'stock' | 'etf' | 'mutual_fund' | 'crypto' | 'bond';
  exchange: string;
  currency: string;
  country: string;
  sector?: string;
  industry?: string;
  market_cap?: number;
  description?: string;
  current_price: number;
  previous_close: number;
  change: number;
  change_percent: number;
  volume: number;
  avg_volume: number;
  day_high: number;
  day_low: number;
  week_52_high: number;
  week_52_low: number;
  pe_ratio?: number;
  dividend_yield?: number;
  beta?: number;
  eps?: number;
  market_cap_formatted?: string;
  last_updated: string;
}

export interface MarketMover {
  symbol: string;
  name: string;
  type: 'stock' | 'etf' | 'mutual_fund' | 'crypto' | 'bond';
  exchange: string;
  current_price: number;
  change: number;
  change_percent: number;
  volume: number;
}

export interface TrendingAsset {
  symbol: string;
  name: string;
  type: 'stock' | 'etf' | 'mutual_fund' | 'crypto' | 'bond';
  exchange: string;
  sector?: string;
  current_price: number;
  change_percent: number;
  volume: number;
  mentions: number;
  sentiment_score: number;
}

// Hook to search for assets
export const useAssetSearch = (query: string, enabled = true) => {
  return useQuery<AssetSearchResult[], Error>(
    [MARKET_DATA_QUERY_KEYS.ASSET_SEARCH, query],
    async () => {
      // Use the actual backend endpoint
      const { data } = await apiClient.get(`/market/assets?search=${encodeURIComponent(query)}&limit=20`);
      
      // Transform backend response to match our interface and filter out currencies
      return (data.assets || [])
        .filter((asset: any) => {
          const symbol = asset.symbol?.toUpperCase();
          const assetType = asset.asset_type?.toLowerCase();
          
          // For specific searches, allow all results
          // For general browsing, exclude currency symbols
          if (query.length >= 3) return true;
          
          // Exclude currency symbols and other non-tradable assets for general browsing
          const excludedSymbols = ['USD', 'EUR', 'GBP', 'JPY', 'CAD', 'AUD', 'CHF', 'CNY', 'INR'];
          const excludedTypes = ['currency', 'forex'];
          
          return !excludedSymbols.includes(symbol) && 
                 !excludedTypes.includes(assetType);
        })
        .map((asset: any) => ({
          symbol: asset.symbol,
          name: asset.name,
          type: asset.asset_type?.toLowerCase() || 'stock',
          exchange: asset.exchange,
          currency: asset.currency || 'USD',
          country: asset.country || 'US',
          sector: asset.sector,
          industry: asset.industry,
          market_cap: asset.market_cap,
          description: asset.description
        }));
    },
    {
      enabled: enabled && query.length >= 2, // Only search if query is at least 2 characters
      staleTime: 5 * 60 * 1000, // 5 minutes cache
      retry: 2,
      refetchOnWindowFocus: false,
    }
  );
};

// Hook to get detailed asset information
export const useAssetDetails = (symbol: string, enabled = true) => {
  return useQuery<AssetDetails, Error>(
    [MARKET_DATA_QUERY_KEYS.ASSET_DETAILS, symbol],
    async () => {
      try {
        // Get asset details and current price from backend
        const [assetResponse, priceResponse] = await Promise.all([
          apiClient.get(`/market/assets/${symbol}`),
          apiClient.get(`/market/prices/${symbol}`)
        ]);
        
        // Combine asset info with current price data
        const asset = assetResponse.data;
        const price = priceResponse.data;
        
        return {
          symbol: asset.symbol,
          name: asset.name,
          type: asset.asset_type?.toLowerCase() || 'stock',
          exchange: asset.exchange,
          currency: asset.currency || 'USD',
          country: asset.country || 'US',
          sector: asset.sector,
          industry: asset.industry,
          market_cap: asset.market_cap,
          description: asset.description || `${asset.name} (${asset.symbol}) is a ${asset.asset_type?.toLowerCase() || 'stock'} traded on ${asset.exchange}.`,
          current_price: price.current_price,
          change: price.change,
          change_percent: price.change_percent,
          previous_close: price.previous_close,
          day_high: price.high,
          day_low: price.low,
          week_52_high: asset.week_52_high || price.high,
          week_52_low: asset.week_52_low || price.low,
          volume: asset.volume || 0,
          avg_volume: asset.avg_volume || 0,
          pe_ratio: asset.pe_ratio,
          dividend_yield: asset.dividend_yield,
          beta: asset.beta,
          eps: asset.eps,
          last_updated: new Date().toISOString()
        };
      } catch (error) {
        console.error('Error fetching asset details:', error);
        throw error;
      }
    },
    {
      enabled: enabled && !!symbol,
      staleTime: 2 * 60 * 1000, // 2 minutes cache for real-time data
      retry: 2,
      refetchInterval: 30 * 1000, // Auto-refresh every 30 seconds for price updates
    }
  );
};

// Hook to get trending assets
export const useTrendingAssets = () => {
  return useQuery<TrendingAsset[], Error>(
    MARKET_DATA_QUERY_KEYS.TRENDING_ASSETS,
    async () => {
      // Get all assets from backend and filter out currencies
      const { data } = await apiClient.get('/market/assets?limit=20');
      
      // Filter out currency symbols and other non-tradable assets
      const tradableAssets = (data.assets || []).filter((asset: any) => {
        const symbol = asset.symbol?.toUpperCase();
        const assetType = asset.asset_type?.toLowerCase();
        
        // Exclude currency symbols and other non-tradable assets
        const excludedSymbols = ['USD', 'EUR', 'GBP', 'JPY', 'CAD', 'AUD', 'CHF', 'CNY', 'INR'];
        const excludedTypes = ['currency', 'forex'];
        
        return !excludedSymbols.includes(symbol) && 
               !excludedTypes.includes(assetType) &&
               symbol && 
               asset.name;
      });
      
      // Get the first 5 tradable assets to make them "trending"
      const trendingSymbols = tradableAssets.slice(0, 5);
      
      const trendingAssets = await Promise.all(
        trendingSymbols.map(async (asset: any) => {
          try {
            const priceResponse = await apiClient.get(`/market/prices/${asset.symbol}`);
            const price = priceResponse.data;
            
            return {
              symbol: asset.symbol,
              name: asset.name,
              type: asset.asset_type?.toLowerCase() || 'stock',
              exchange: asset.exchange,
              sector: asset.sector,
              current_price: price.current_price,
              change_percent: price.change_percent,
              volume: Math.floor(Math.random() * 100000000), // Simulated volume for trending
              mentions: Math.floor(Math.random() * 2000) + 500, // Simulated mentions
              sentiment_score: Math.random() * 0.4 + 0.6 // Simulated sentiment (0.6-1.0)
            };
          } catch (error) {
            console.warn(`Failed to get price for ${asset.symbol}:`, error);
            // Return asset with fallback data if price fetch fails
            return {
              symbol: asset.symbol,
              name: asset.name,
              type: asset.asset_type?.toLowerCase() || 'stock',
              exchange: asset.exchange,
              sector: asset.sector,
              current_price: 100 + Math.random() * 400, // Fallback price
              change_percent: (Math.random() - 0.5) * 10, // Random change
              volume: Math.floor(Math.random() * 100000000),
              mentions: Math.floor(Math.random() * 2000) + 500,
              sentiment_score: Math.random() * 0.4 + 0.6
            };
          }
        })
      );
      
      return trendingAssets;
    },
    {
      staleTime: 5 * 60 * 1000, // 5 minutes cache
      retry: 2,
      refetchInterval: 5 * 60 * 1000, // Auto-refresh every 5 minutes
    }
  );
};

// Hook to get market movers (gainers/losers)
export const useMarketMovers = (type: 'gainers' | 'losers' = 'gainers') => {
  return useQuery<MarketMover[], Error>(
    [MARKET_DATA_QUERY_KEYS.MARKET_MOVERS, type],
    async () => {
      // Get all assets from backend
      const { data } = await apiClient.get('/market/assets?limit=30');
      
      // Filter out currency symbols and other non-tradable assets
      const tradableAssets = (data.assets || []).filter((asset: any) => {
        const symbol = asset.symbol?.toUpperCase();
        const assetType = asset.asset_type?.toLowerCase();
        
        // Exclude currency symbols and other non-tradable assets
        const excludedSymbols = ['USD', 'EUR', 'GBP', 'JPY', 'CAD', 'AUD', 'CHF', 'CNY', 'INR'];
        const excludedTypes = ['currency', 'forex'];
        
        return !excludedSymbols.includes(symbol) && 
               !excludedTypes.includes(assetType) &&
               symbol && 
               asset.name;
      });
      
      // Get price data for all tradable assets
      const assetsWithPrices = await Promise.all(
        tradableAssets.map(async (asset: any) => {
          try {
            const priceResponse = await apiClient.get(`/market/prices/${asset.symbol}`);
            const price = priceResponse.data;
            
            return {
              symbol: asset.symbol,
              name: asset.name,
              type: asset.asset_type?.toLowerCase() || 'stock',
              exchange: asset.exchange,
              current_price: price.current_price,
              change: price.change,
              change_percent: price.change_percent,
              volume: Math.floor(Math.random() * 100000000) // Simulated volume
            };
          } catch (error) {
            console.warn(`Failed to get price for ${asset.symbol}:`, error);
            // Skip assets that don't have price data
            return null;
          }
        })
      );
      
      // Filter out failed requests and sort by change_percent
      const validAssets = assetsWithPrices.filter(asset => asset !== null) as MarketMover[];
      
      if (type === 'gainers') {
        // Sort by highest positive change percentage
        return validAssets
          .filter(asset => asset.change_percent > 0)
          .sort((a, b) => b.change_percent - a.change_percent)
          .slice(0, 10);
      } else {
        // Sort by lowest negative change percentage (biggest losers)
        return validAssets
          .filter(asset => asset.change_percent < 0)
          .sort((a, b) => a.change_percent - b.change_percent)
          .slice(0, 10);
      }
    },
    {
      staleTime: 2 * 60 * 1000, // 2 minutes cache
      retry: 2,
      refetchInterval: 2 * 60 * 1000, // Auto-refresh every 2 minutes
    }
  );
};

// Hook to add asset to watchlist (if watchlist functionality exists)
export const useAddToWatchlist = () => {
  const queryClient = useQueryClient();
  
  return useMutation<void, Error, string>(
    async (symbol: string) => {
      await apiClient.post('/watchlist/', { symbol });
    },
    {
      onSuccess: () => {
        // Invalidate watchlist queries if they exist
        queryClient.invalidateQueries('watchlist');
      },
    }
  );
};
