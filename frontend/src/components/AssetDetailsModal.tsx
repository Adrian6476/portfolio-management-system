import React, { useState } from 'react';
import { useAssetDetails } from '@/hooks/useMarketData';
import { LoadingSpinner, ErrorMessage, Button } from './ui';

interface AssetDetailsModalProps {
  symbol: string;
  isOpen: boolean;
  onClose: () => void;
  onAddToPortfolio?: (symbol: string, name: string) => void;
}

const AssetDetailsModal: React.FC<AssetDetailsModalProps> = ({
  symbol,
  isOpen,
  onClose,
  onAddToPortfolio
}) => {
  const [activeTab, setActiveTab] = useState<'overview' | 'financials' | 'performance'>('overview');
  const { data: assetDetails, isLoading, error } = useAssetDetails(symbol, isOpen);

  if (!isOpen) return null;

  const formatCurrency = (value: number | undefined, currency = 'USD') => {
    if (value === undefined || value === null) return 'N/A';
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: currency,
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(value);
  };

  const formatNumber = (value: number | undefined) => {
    if (value === undefined || value === null) return 'N/A';
    return new Intl.NumberFormat('en-US').format(value);
  };

  const formatPercentage = (value: number | undefined) => {
    if (value === undefined || value === null) return 'N/A';
    return `${value > 0 ? '+' : ''}${value.toFixed(2)}%`;
  };

  const formatMarketCap = (value: number | undefined) => {
    if (value === undefined || value === null) return 'N/A';
    
    if (value >= 1e12) return `$${(value / 1e12).toFixed(2)}T`;
    if (value >= 1e9) return `$${(value / 1e9).toFixed(2)}B`;
    if (value >= 1e6) return `$${(value / 1e6).toFixed(2)}M`;
    return formatCurrency(value);
  };

  const getChangeColor = (change: number | undefined) => {
    if (change === undefined || change === null) return 'text-gray-600';
    return change >= 0 ? 'text-green-600' : 'text-red-600';
  };

  const handleAddToPortfolio = () => {
    if (onAddToPortfolio && assetDetails) {
      onAddToPortfolio(assetDetails.symbol, assetDetails.name);
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-hidden">
        {/* Header */}
        <div className="p-6 border-b border-gray-200">
          <div className="flex items-center justify-between">
            <div className="flex items-center">
              <div>
                <h2 className="text-2xl font-bold text-gray-900">
                  {symbol}
                </h2>
                {assetDetails && (
                  <p className="text-gray-600">{assetDetails.name}</p>
                )}
              </div>
              {assetDetails && (
                <div className="ml-6">
                  <div className="text-right">
                    <div className="text-2xl font-bold text-gray-900">
                      {formatCurrency(assetDetails.current_price, assetDetails.currency)}
                    </div>
                    <div className={`text-sm font-medium ${getChangeColor(assetDetails.change)}`}>
                      {formatCurrency(assetDetails.change, assetDetails.currency)} ({formatPercentage(assetDetails.change_percent)})
                    </div>
                  </div>
                </div>
              )}
            </div>
            
            <div className="flex items-center space-x-3">
              {onAddToPortfolio && (
                <Button
                  onClick={handleAddToPortfolio}
                  disabled={!assetDetails}
                  className="flex items-center"
                >
                  <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                  </svg>
                  Add to Portfolio
                </Button>
              )}
              
              <button
                onClick={onClose}
                className="text-gray-400 hover:text-gray-600 p-2"
              >
                <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
                </svg>
              </button>
            </div>
          </div>
        </div>

        {/* Content */}
        <div className="flex-1 overflow-y-auto">
          {isLoading ? (
            <div className="flex items-center justify-center py-12">
              <LoadingSpinner size="lg" />
              <span className="ml-3 text-gray-600">Loading asset details...</span>
            </div>
          ) : error ? (
            <div className="p-6">
              <ErrorMessage message={error.message || 'Failed to load asset details'} />
            </div>
          ) : assetDetails ? (
            <>
              {/* Tabs */}
              <div className="border-b border-gray-200">
                <nav className="flex">
                  <button
                    onClick={() => setActiveTab('overview')}
                    className={`px-6 py-3 text-sm font-medium border-b-2 ${
                      activeTab === 'overview'
                        ? 'border-blue-500 text-blue-600'
                        : 'border-transparent text-gray-500 hover:text-gray-700'
                    }`}
                  >
                    Overview
                  </button>
                  <button
                    onClick={() => setActiveTab('financials')}
                    className={`px-6 py-3 text-sm font-medium border-b-2 ${
                      activeTab === 'financials'
                        ? 'border-blue-500 text-blue-600'
                        : 'border-transparent text-gray-500 hover:text-gray-700'
                    }`}
                  >
                    Financials
                  </button>
                  <button
                    onClick={() => setActiveTab('performance')}
                    className={`px-6 py-3 text-sm font-medium border-b-2 ${
                      activeTab === 'performance'
                        ? 'border-blue-500 text-blue-600'
                        : 'border-transparent text-gray-500 hover:text-gray-700'
                    }`}
                  >
                    Performance
                  </button>
                </nav>
              </div>

              {/* Tab Content */}
              <div className="p-6">
                {activeTab === 'overview' && (
                  <div className="space-y-6">
                    {/* Basic Info */}
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      <div className="space-y-4">
                        <h3 className="text-lg font-semibold text-gray-900">Basic Information</h3>
                        <div className="space-y-2">
                          <div className="flex justify-between">
                            <span className="text-gray-600">Symbol:</span>
                            <span className="font-medium">{assetDetails.symbol}</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-gray-600">Exchange:</span>
                            <span className="font-medium">{assetDetails.exchange}</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-gray-600">Type:</span>
                            <span className="font-medium capitalize">{assetDetails.type}</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-gray-600">Currency:</span>
                            <span className="font-medium">{assetDetails.currency}</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-gray-600">Country:</span>
                            <span className="font-medium">{assetDetails.country}</span>
                          </div>
                          {assetDetails.sector && (
                            <div className="flex justify-between">
                              <span className="text-gray-600">Sector:</span>
                              <span className="font-medium">{assetDetails.sector}</span>
                            </div>
                          )}
                          {assetDetails.industry && (
                            <div className="flex justify-between">
                              <span className="text-gray-600">Industry:</span>
                              <span className="font-medium">{assetDetails.industry}</span>
                            </div>
                          )}
                        </div>
                      </div>

                      <div className="space-y-4">
                        <h3 className="text-lg font-semibold text-gray-900">Price Information</h3>
                        <div className="space-y-2">
                          <div className="flex justify-between">
                            <span className="text-gray-600">Previous Close:</span>
                            <span className="font-medium">{formatCurrency(assetDetails.previous_close, assetDetails.currency)}</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-gray-600">Day High:</span>
                            <span className="font-medium">{formatCurrency(assetDetails.day_high, assetDetails.currency)}</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-gray-600">Day Low:</span>
                            <span className="font-medium">{formatCurrency(assetDetails.day_low, assetDetails.currency)}</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-gray-600">52W High:</span>
                            <span className="font-medium">{formatCurrency(assetDetails.week_52_high, assetDetails.currency)}</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-gray-600">52W Low:</span>
                            <span className="font-medium">{formatCurrency(assetDetails.week_52_low, assetDetails.currency)}</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-gray-600">Volume:</span>
                            <span className="font-medium">{formatNumber(assetDetails.volume)}</span>
                          </div>
                          <div className="flex justify-between">
                            <span className="text-gray-600">Avg Volume:</span>
                            <span className="font-medium">{formatNumber(assetDetails.avg_volume)}</span>
                          </div>
                        </div>
                      </div>
                    </div>

                    {/* Description */}
                    {assetDetails.description && (
                      <div>
                        <h3 className="text-lg font-semibold text-gray-900 mb-3">Description</h3>
                        <p className="text-gray-600 leading-relaxed">{assetDetails.description}</p>
                      </div>
                    )}
                  </div>
                )}

                {activeTab === 'financials' && (
                  <div className="space-y-6">
                    <h3 className="text-lg font-semibold text-gray-900">Financial Metrics</h3>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      <div className="space-y-2">
                        <div className="flex justify-between">
                          <span className="text-gray-600">Market Cap:</span>
                          <span className="font-medium">{formatMarketCap(assetDetails.market_cap)}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-gray-600">P/E Ratio:</span>
                          <span className="font-medium">{assetDetails.pe_ratio?.toFixed(2) || 'N/A'}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-gray-600">EPS:</span>
                          <span className="font-medium">{formatCurrency(assetDetails.eps, assetDetails.currency)}</span>
                        </div>
                      </div>
                      <div className="space-y-2">
                        <div className="flex justify-between">
                          <span className="text-gray-600">Dividend Yield:</span>
                          <span className="font-medium">{formatPercentage(assetDetails.dividend_yield)}</span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-gray-600">Beta:</span>
                          <span className="font-medium">{assetDetails.beta?.toFixed(2) || 'N/A'}</span>
                        </div>
                      </div>
                    </div>
                  </div>
                )}

                {activeTab === 'performance' && (
                  <div className="space-y-6">
                    <h3 className="text-lg font-semibold text-gray-900">Performance Metrics</h3>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                      <div className="space-y-2">
                        <div className="flex justify-between">
                          <span className="text-gray-600">Today's Change:</span>
                          <span className={`font-medium ${getChangeColor(assetDetails.change)}`}>
                            {formatCurrency(assetDetails.change, assetDetails.currency)} ({formatPercentage(assetDetails.change_percent)})
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-gray-600">52W Range:</span>
                          <span className="font-medium">
                            {formatCurrency(assetDetails.week_52_low, assetDetails.currency)} - {formatCurrency(assetDetails.week_52_high, assetDetails.currency)}
                          </span>
                        </div>
                      </div>
                      <div className="space-y-2">
                        <div className="flex justify-between">
                          <span className="text-gray-600">Volume Ratio:</span>
                          <span className="font-medium">
                            {assetDetails.avg_volume ? ((assetDetails.volume / assetDetails.avg_volume) * 100).toFixed(0) + '%' : 'N/A'}
                          </span>
                        </div>
                        <div className="flex justify-between">
                          <span className="text-gray-600">Last Updated:</span>
                          <span className="font-medium text-sm">
                            {new Date(assetDetails.last_updated).toLocaleString()}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>
                )}
              </div>
            </>
          ) : (
            <div className="p-6 text-center text-gray-500">
              <p>No asset details available</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default AssetDetailsModal;
