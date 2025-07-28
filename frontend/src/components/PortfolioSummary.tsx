'use client'

import { useQuery } from 'react-query'
import apiClient from '@/lib/api'
import { PortfolioSummaryResponse, QUERY_KEYS } from '@/types/portfolio'

// Loading spinner component
function LoadingSpinner() {
  return (
    <div className="flex items-center justify-center p-8">
      <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>
      <span className="ml-2 text-gray-600">Loading portfolio summary...</span>
    </div>
  )
}

// Error alert component
function ErrorAlert({ message }: { message: string }) {
  return (
    <div className="bg-red-50 border border-red-200 rounded-lg p-4">
      <div className="flex">
        <div className="flex-shrink-0">
          <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
            <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
          </svg>
        </div>
        <div className="ml-3">
          <h3 className="text-sm font-medium text-red-800">Error loading portfolio summary</h3>
          <p className="mt-1 text-sm text-red-700">{message}</p>
        </div>
      </div>
    </div>
  )
}

// Metric card component
function MetricCard({ title, value, subtitle }: { title: string; value: number; subtitle?: string }) {
  const formatValue = (val: number) => {
    if (title.toLowerCase().includes('cost') || title.toLowerCase().includes('value')) {
      return new Intl.NumberFormat('en-US', {
        style: 'currency',
        currency: 'USD',
      }).format(val)
    }
    return val.toLocaleString()
  }

  return (
    <div className="bg-white rounded-lg shadow p-6">
      <div className="flex items-center">
        <div className="flex-1">
          <p className="text-sm font-medium text-gray-600">{title}</p>
          <p className="text-2xl font-bold text-gray-900">{formatValue(value)}</p>
          {subtitle && <p className="text-sm text-gray-500">{subtitle}</p>}
        </div>
      </div>
    </div>
  )
}

// Asset allocation card component
function AssetAllocationCard({ allocations }: { allocations: any[] }) {
  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h3 className="text-lg font-semibold text-gray-900 mb-4">Asset Allocation</h3>
      <div className="space-y-3">
        {allocations.map((allocation, index) => (
          <div key={index} className="flex items-center justify-between">
            <div className="flex items-center">
              <div className="w-3 h-3 rounded-full bg-blue-500 mr-3"></div>
              <span className="text-sm font-medium text-gray-700">{allocation.asset_type}</span>
            </div>
            <div className="text-right">
              <div className="text-sm font-semibold text-gray-900">
                {new Intl.NumberFormat('en-US', {
                  style: 'currency',
                  currency: 'USD',
                }).format(allocation.total_value)}
              </div>
              <div className="text-xs text-gray-500">{allocation.percentage.toFixed(1)}%</div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

// Top holdings card component
function TopHoldingsCard({ holdings }: { holdings: any[] }) {
  return (
    <div className="bg-white rounded-lg shadow p-6">
      <h3 className="text-lg font-semibold text-gray-900 mb-4">Top Holdings</h3>
      <div className="space-y-3">
        {holdings.slice(0, 5).map((holding, index) => (
          <div key={index} className="flex items-center justify-between">
            <div>
              <div className="text-sm font-medium text-gray-900">{holding.symbol}</div>
              <div className="text-xs text-gray-500">{holding.name}</div>
            </div>
            <div className="text-right">
              <div className="text-sm font-semibold text-gray-900">
                {new Intl.NumberFormat('en-US', {
                  style: 'currency',
                  currency: 'USD',
                }).format(holding.total_value || (holding.quantity * holding.average_cost))}
              </div>
              <div className="text-xs text-gray-500">{holding.quantity} shares</div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}

export default function PortfolioSummary() {
  const { data, isLoading, isError, error } = useQuery<PortfolioSummaryResponse>(
    QUERY_KEYS.PORTFOLIO_SUMMARY,
    async () => {
      const response = await apiClient.get('/portfolio/summary')
      return response.data
    },
    {
      retry: 2,
      refetchOnWindowFocus: false,
    }
  )

  if (isLoading) {
    return (
      <div className="bg-white rounded-lg shadow">
        <LoadingSpinner />
      </div>
    )
  }

  if (isError) {
    const errorMessage = error instanceof Error ? error.message : 'Failed to fetch portfolio summary'
    return <ErrorAlert message={errorMessage} />
  }

  if (!data) {
    return <ErrorAlert message="No portfolio data available" />
  }

  return (
    <div className="space-y-6">
      {/* Summary metrics */}
      <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
        <MetricCard 
          title="Total Holdings" 
          value={data.summary.total_holdings}
          subtitle="Number of positions"
        />
        <MetricCard 
          title="Total Cost" 
          value={data.summary.total_cost}
          subtitle="Total invested amount"
        />
        <MetricCard 
          title="Total Shares" 
          value={data.summary.total_shares}
          subtitle="Total number of shares"
        />
      </div>

      {/* Asset allocation and top holdings */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {data.asset_allocation && data.asset_allocation.length > 0 && (
          <AssetAllocationCard allocations={data.asset_allocation} />
        )}
        {data.top_holdings && data.top_holdings.length > 0 && (
          <TopHoldingsCard holdings={data.top_holdings} />
        )}
      </div>
    </div>
  )
}
