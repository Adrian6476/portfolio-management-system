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
