import { useQuery } from 'react-query';

import apiClient from '@/lib/api';

import { PortfolioSummaryResponse, QUERY_KEYS } from '@/types/portfolio';

  

// Hook specifically for header summary display

export const usePortfolioSummary = () => {

return useQuery<PortfolioSummaryResponse, Error>(

QUERY_KEYS.PORTFOLIO_SUMMARY,

async () => {

const response = await apiClient.get('/portfolio/summary');

return response.data;

},

{

retry: 2,

refetchOnWindowFocus: true,

staleTime: 30 * 1000, // 30 seconds cache for daily changes

refetchInterval: 60 * 1000, // Auto-refresh every minute

}

);

};

  

// Utility function to format currency

export const formatCurrency = (amount: number): string => {

return new Intl.NumberFormat('en-US', {

style: 'currency',

currency: 'USD',

minimumFractionDigits: 0,

maximumFractionDigits: 0,

}).format(amount);

};

  

// Utility function to format percentage

export const formatPercentage = (percentage: number): string => {

const sign = percentage >= 0 ? '+' : '';

return `${sign}${percentage.toFixed(2)}%`;

};