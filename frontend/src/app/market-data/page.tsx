'use client'

import { useState } from 'react'
import { useRouter } from 'next/navigation'
import AssetSearch from '@/components/AssetSearch'
import { useTrendingAssets, useMarketMovers, TrendingAsset, MarketMover } from '@/hooks/useMarketData'
import { LoadingSpinner, ErrorMessage } from '@/components/ui'
import AssetDetailsModal from '@/components/AssetDetailsModal'

export default function MarketDataPage() {
  const router = useRouter()
  const [selectedAsset, setSelectedAsset] = useState<string | null>(null);
  const [showAssetModal, setShowAssetModal] = useState(false);
  const [activeSection, setActiveSection] = useState<'gainers' | 'losers'>('gainers');

  const { data: trendingAssets, isLoading: trendingLoading, error: trendingError } = useTrendingAssets();
  const { data: gainers, isLoading: gainersLoading, error: gainersError } = useMarketMovers('gainers');
  const { data: losers, isLoading: losersLoading, error: losersError } = useMarketMovers('losers');

  const handleAssetSelect = (symbol: string, name: string) => {
    // Navigate to root page with add view and pre-selected asset
    router.push(`/?view=add&symbol=${symbol}&name=${encodeURIComponent(name)}`);
  };

  const handleViewDetails = (symbol: string) => {
    setSelectedAsset(symbol);
    setShowAssetModal(true);
  };

  const handleAssetFromModal = (symbol: string, name: string) => {
    handleAssetSelect(symbol, name);
    setShowAssetModal(false);
    setSelectedAsset(null);
  };

  const formatCurrency = (value: number | undefined) => {
    if (value === undefined || value === null) return 'N/A';
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(value);
  };

  const formatPercentage = (value: number | undefined) => {
    if (value === undefined || value === null) return 'N/A';
    return `${value > 0 ? '+' : ''}${value.toFixed(2)}%`;
  };

  const getChangeColor = (change: number | undefined) => {
    if (change === undefined || change === null) return 'text-gray-600';
    return change >= 0 ? 'text-green-600' : 'text-red-600';
  };

  const getAssetTypeIcon = (type: string) => {
    switch (type) {
      case 'stock':
        return (
          <div className="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
            <svg className="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
          </div>
        );
      case 'etf':
        return (
          <div className="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center">
            <svg className="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
          </div>
        );
      case 'crypto':
        return (
          <div className="w-8 h-8 bg-orange-100 rounded-full flex items-center justify-center">
            <svg className="w-4 h-4 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
        );
      default:
        return (
          <div className="w-8 h-8 bg-gray-100 rounded-full flex items-center justify-center">
            <svg className="w-4 h-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12h6m-6 4h6m2 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
            </svg>
          </div>
        );
    }
  };

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="flex justify-between items-center mb-8">
          <div className="flex items-center">
            <button
              onClick={() => router.back()}
              className="text-gray-600 hover:text-gray-900 mr-4 flex items-center transition-colors"
            >
              <svg className="w-5 h-5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
              </svg>
              Back
            </button>
            <h1 className="text-3xl font-bold text-gray-900">Market Data</h1>
          </div>
        </div>

        {/* Search Section */}
        <div className="bg-white rounded-lg shadow p-6 mb-8">
          <h2 className="text-xl font-semibold text-gray-900 mb-4">Asset Search</h2>
          <div className="max-w-2xl">
            <AssetSearch 
              onAssetSelect={handleAssetSelect}
              showTrending={true}
              enableDetailsModal={true}
              placeholder="Search for any stock, ETF, cryptocurrency..."
              className="w-full"
            />
          </div>
        </div>

        {/* Market Overview */}
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-8">
          {/* Trending Assets */}
          <div className="bg-white rounded-lg shadow">
            <div className="p-6 border-b border-gray-200">
              <h3 className="text-lg font-semibold text-gray-900">Trending Assets</h3>
              <p className="text-sm text-gray-600 mt-1">Most searched assets today</p>
            </div>
            <div className="p-6">
              {trendingLoading ? (
                <div className="flex items-center justify-center py-8">
                  <LoadingSpinner size="lg" />
                  <span className="ml-3 text-gray-600">Loading trending assets...</span>
                </div>
              ) : trendingError ? (
                <ErrorMessage message={trendingError.message || 'Failed to load trending assets'} />
              ) : trendingAssets && trendingAssets.length > 0 ? (
                <div className="space-y-3">
                  {trendingAssets.slice(0, 10).map((asset, index) => (
                    <div key={asset.symbol} className="flex items-center justify-between p-3 hover:bg-gray-50 rounded-lg transition-colors cursor-pointer"
                      onClick={() => handleViewDetails(asset.symbol)}
                      title="Click to view details"
                    >
                      <div className="flex items-center">
                        <span className="text-sm text-gray-500 w-6">{index + 1}</span>
                        {getAssetTypeIcon(asset.type)}
                        <div className="ml-3">
                          <p className="font-medium text-gray-900">{asset.symbol}</p>
                          <p className="text-sm text-gray-600 truncate max-w-40">{asset.name}</p>
                        </div>
                      </div>
                      <div className="flex items-center space-x-2">
                        <div className="text-right">
                          <p className="font-medium text-gray-900">
                            {formatCurrency(asset.current_price)}
                          </p>
                          <p className={`text-sm ${getChangeColor(asset.change_percent)}`}>
                            {formatPercentage(asset.change_percent)}
                          </p>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-gray-500 text-center py-8">No trending assets available</p>
              )}
            </div>
          </div>

          {/* Market Movers */}
          <div className="bg-white rounded-lg shadow">
            <div className="p-6 border-b border-gray-200">
              <div className="flex items-center justify-between">
                <div>
                  <h3 className="text-lg font-semibold text-gray-900">Market Movers</h3>
                  <p className="text-sm text-gray-600 mt-1">Top performers and losers</p>
                </div>
                <div className="flex bg-gray-100 rounded-lg p-1">
                  <button
                    onClick={() => setActiveSection('gainers')}
                    className={`px-3 py-1 text-sm rounded-md transition-colors ${
                      activeSection === 'gainers' ? 'bg-white shadow text-gray-900' : 'text-gray-600'
                    }`}
                  >
                    Gainers
                  </button>
                  <button
                    onClick={() => setActiveSection('losers')}
                    className={`px-3 py-1 text-sm rounded-md transition-colors ${
                      activeSection === 'losers' ? 'bg-white shadow text-gray-900' : 'text-gray-600'
                    }`}
                  >
                    Losers
                  </button>
                </div>
              </div>
            </div>
            <div className="p-6">
              {(gainersLoading || losersLoading) ? (
                <div className="flex items-center justify-center py-8">
                  <LoadingSpinner size="lg" />
                  <span className="ml-3 text-gray-600">Loading market movers...</span>
                </div>
              ) : (gainersError || losersError) ? (
                <ErrorMessage message={(gainersError || losersError)?.message || 'Failed to load market movers'} />
              ) : (gainers || losers) ? (
                <div className="space-y-3">
                  {(activeSection === 'gainers' ? gainers || [] : losers || [])
                    .slice(0, 10)
                    .map((asset: MarketMover, index: number) => (
                    <div key={asset.symbol} className="flex items-center justify-between p-3 hover:bg-gray-50 rounded-lg transition-colors cursor-pointer"
                      onClick={() => handleViewDetails(asset.symbol)}
                      title="Click to view details"
                    >
                      <div className="flex items-center">
                        <span className="text-sm text-gray-500 w-6">{index + 1}</span>
                        {getAssetTypeIcon(asset.type)}
                        <div className="ml-3">
                          <p className="font-medium text-gray-900">{asset.symbol}</p>
                          <p className="text-sm text-gray-600 truncate max-w-40">{asset.name}</p>
                        </div>
                      </div>
                      <div className="flex items-center space-x-2">
                        <div className="text-right">
                          <p className="font-medium text-gray-900">{formatCurrency(asset.current_price)}</p>
                          <p className={`text-sm font-medium ${getChangeColor(asset.change_percent)}`}>
                            {formatPercentage(asset.change_percent)}
                          </p>
                        </div>
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <p className="text-gray-500 text-center py-8">No market movers available</p>
              )}
            </div>
          </div>
        </div>

        {/* Asset Details Modal */}
        {selectedAsset && (
          <AssetDetailsModal
            symbol={selectedAsset}
            isOpen={showAssetModal}
            onClose={() => {
              setShowAssetModal(false);
              setSelectedAsset(null);
            }}
            onAddToPortfolio={handleAssetFromModal}
          />
        )}
      </div>
    </div>
  )
}
