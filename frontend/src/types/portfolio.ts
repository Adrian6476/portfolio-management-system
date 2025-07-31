// This file contains the shared TypeScript types for the portfolio.
// Developer A must define these interfaces FIRST for team coordination.

export interface Holding {
  id: string;
  symbol: string;
  name: string;
  asset_type: string;
  quantity: number;
  average_cost: number;
  purchase_date: string;
}

export interface PortfolioResponse {
  holdings: Holding[];
  total_holdings: number;
}

export interface Summary {
  total_holdings: number;
  total_cost: number;
  total_shares: number;
  total_market_value?: number;
  daily_change?: number;
  daily_change_percent?: number;
  unrealized_gain_loss?: number;
  unrealized_gain_loss_percent?: number;
}

export interface AssetAllocation {
  asset_type: string;
  count: number;
  total_value: number;
  percentage: number;
}

export interface PortfolioSummaryResponse {
  summary: Summary;
  asset_allocation: AssetAllocation[];
  top_holdings: Holding[];
}

export interface CurrentPriceResponse {
  symbol: string;
  current_price: number;
  change: number;
  change_percent: number;
  high: number;
  low: number;
  open: number;
  previous_close: number;
  timestamp: number;
}

// React Query Keys - MUST USE EXACTLY
export const QUERY_KEYS = {
  PORTFOLIO_SUMMARY: 'portfolioSummary',
  PORTFOLIO_HOLDINGS: 'portfolioHoldings',
  CURRENT_PRICE: (symbol: string) => ['currentPrice', symbol],
} as const;

// Market Data Types
export interface AssetSearchResult {
  symbol: string;
  name: string;
  type: string;
  exchange: string;
  sector?: string;
}

export interface TrendingAsset {
  symbol: string;
  name: string;
  type: string;
  exchange: string;
  sector?: string;
  price?: number;
  change?: number;
  change_percent?: number;
}

export interface MarketMover {
  symbol: string;
  name: string;
  type: string;
  exchange: string;
  price: number;
  change: number;
  change_percent: number;
  volume?: number;
}

export interface MarketMoversResponse {
  gainers: MarketMover[];
  losers: MarketMover[];
}

export interface AssetDetails {
  symbol: string;
  name: string;
  type: string;
  exchange: string;
  currency: string;
  country: string;
  sector?: string;
  industry?: string;
  description?: string;
  current_price: number;
  change: number;
  change_percent: number;
  previous_close: number;
  day_high: number;
  day_low: number;
  week_52_high: number;
  week_52_low: number;
  volume: number;
  avg_volume: number;
  market_cap?: number;
  pe_ratio?: number;
  eps?: number;
  dividend_yield?: number;
  beta?: number;
  last_updated: string;
}
