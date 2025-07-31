import { useQuery, useMutation, useQueryClient } from 'react-query';
import apiClient from '../lib/api';

// Query keys for analytics
export const ANALYTICS_QUERY_KEYS = {
  RISK_METRICS: 'analytics-risk',
  PERFORMANCE_ANALYTICS: 'analytics-performance',
  ASSET_ALLOCATION: 'analytics-allocation',
  WHAT_IF_ANALYSIS: 'analytics-whatif',
} as const;

// Types based on actual backend implementation
export interface RiskMetrics {
  risk_assessment: {
    overall_risk_level: string;
    concentration_risk: string;
    herfindahl_index: number;
    diversification_score: number;
  };
  sector_diversification: Array<{
    sector: string;
    holdings_count: number;
    sector_value: number;
    percentage: number;
  }>;
  volatility_metrics: {
    portfolio_beta: number;
    sharpe_ratio: number;
    max_drawdown: number;
    var_95: number;
    expected_volatility: number;
    portfolio_return: number;
  };
  risk_recommendations: string[];
}

export interface PerformanceAnalytics {
  portfolio_performance: {
    total_cost: number;
    current_value: number;
    total_gain_loss: number;
    total_return_percent: number;
    total_holdings: number;
    period: string;
  };
  historical_snapshots: Array<{
    date: string;
    total_value: number;
    total_cost: number;
    unrealized_pnl: number;
  }>;
  top_performers: Array<{
    symbol: string;
    name: string;
    quantity: number;
    average_cost: number;
    current_price: number;
    total_cost: number;
    current_value: number;
    gain_loss: number;
    gain_loss_percent: number;
  }>;
  last_updated: string;
  warnings?: string[];
}

export interface AssetAllocation {
  allocation_summary: {
    total_portfolio_value: number;
    total_holdings: number;
    allocation_date: string;
  };
  by_asset_type: Array<{
    asset_type: string;
    count: number;
    value: number; // API uses 'value' not 'total_value'
    percentage: number;
  }>;
  by_sector: Array<{
    sector: string;
    count: number;
    value: number; // API uses 'value' not 'total_value'
    percentage: number;
  }>;
  top_holdings: Array<{
    symbol: string;
    name: string;
    quantity: number;
    average_cost: number;
    total_value: number; // Holdings use 'total_value'
    percentage: number;
  }>;
}

export interface WhatIfAnalysisRequest {
  action: 'buy' | 'sell';
  symbol: string;
  quantity: number;
  price: number;
}

export interface WhatIfAnalysisResponse {
  trade_details: {
    action: string;
    symbol: string;
    quantity: number;
    price: number;
    trade_value: number;
    position_change: string;
  };
  portfolio_impact: {
    current_total_value: number;
    new_total_value: number;
    value_change: number;
    current_holdings: number;
    new_holdings: number;
  };
  position_impact: {
    has_current_holding: boolean;
    current_quantity: number;
    current_avg_cost: number;
    new_quantity: number;
    new_avg_cost: number;
  };
  allocation_impact: {
    [key: string]: {
      current_percent: number;
      new_percent: number;
      current_value: number;
      change: number;
    };
  };
  risk_impact: {
    concentration_change: number;
    diversification_impact: string;
  };
  expected_returns: {
    symbol_volatility: number;
    annual_return_estimate: number;
    risk_premium: number;
    risk_adjusted_return: number;
  };
  recommendations: string[];
}

// Hook to fetch risk metrics
export const useRiskMetrics = () => {
  return useQuery<RiskMetrics, Error>(
    ANALYTICS_QUERY_KEYS.RISK_METRICS,
    async () => {
      const { data } = await apiClient.get('/analytics/risk');
      return data;
    },
    {
      staleTime: 5 * 60 * 1000, // 5 minutes cache
      retry: 2,
      refetchOnWindowFocus: false,
    }
  );
};

// Hook to fetch performance analytics
export const usePerformanceAnalytics = () => {
  return useQuery<PerformanceAnalytics, Error>(
    ANALYTICS_QUERY_KEYS.PERFORMANCE_ANALYTICS,
    async () => {
      const { data } = await apiClient.get('/analytics/performance');
      return data;
    },
    {
      staleTime: 5 * 60 * 1000, // 5 minutes cache
      retry: 2,
      refetchOnWindowFocus: false,
    }
  );
};

// Hook to fetch asset allocation
export const useAssetAllocation = () => {
  return useQuery<AssetAllocation, Error>(
    ANALYTICS_QUERY_KEYS.ASSET_ALLOCATION,
    async () => {
      const { data } = await apiClient.get('/analytics/allocation');
      return data;
    },
    {
      staleTime: 5 * 60 * 1000, // 5 minutes cache
      retry: 2,
      refetchOnWindowFocus: false,
    }
  );
};

// Hook for what-if analysis
export const useWhatIfAnalysis = () => {
  const queryClient = useQueryClient();
  
  return useMutation<WhatIfAnalysisResponse, Error, WhatIfAnalysisRequest>(
    async (analysisRequest: WhatIfAnalysisRequest) => {
      const { data } = await apiClient.post('/analytics/whatif', analysisRequest);
      return data;
    },
    {
      onSuccess: () => {
        // Optionally invalidate related queries after what-if analysis
        // This ensures fresh data if user decides to make the actual trade
        queryClient.invalidateQueries(ANALYTICS_QUERY_KEYS.RISK_METRICS);
        queryClient.invalidateQueries(ANALYTICS_QUERY_KEYS.PERFORMANCE_ANALYTICS);
        queryClient.invalidateQueries(ANALYTICS_QUERY_KEYS.ASSET_ALLOCATION);
      },
    }
  );
};

// Utility hook to refresh all analytics data
export const useRefreshAnalytics = () => {
  const queryClient = useQueryClient();
  
  return () => {
    queryClient.invalidateQueries(ANALYTICS_QUERY_KEYS.RISK_METRICS);
    queryClient.invalidateQueries(ANALYTICS_QUERY_KEYS.PERFORMANCE_ANALYTICS);
    queryClient.invalidateQueries(ANALYTICS_QUERY_KEYS.ASSET_ALLOCATION);
  };
};