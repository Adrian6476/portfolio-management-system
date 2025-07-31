import React, { useState } from 'react';
import { useWhatIfAnalysis } from '../hooks/useAnalytics';
import { Button, LoadingSpinner, ErrorMessage } from './ui';
import { UI_CONSTANTS } from './ui';
import AssetSelector from './AssetSelector';
import { AssetSearchResult, TrendingAsset } from '@/hooks/useMarketData';

interface WhatIfAnalysisModalProps {
  isOpen: boolean;
  onClose: () => void;
}

const WhatIfAnalysisModal: React.FC<WhatIfAnalysisModalProps> = ({ isOpen, onClose }) => {
  const [formData, setFormData] = useState({
    action: 'buy' as 'buy' | 'sell',
    symbol: '',
    quantity: '',
    price: '',
  });

  const [analysisResult, setAnalysisResult] = useState<any>(null);
  const [selectedAssetInfo, setSelectedAssetInfo] = useState<AssetSearchResult | TrendingAsset | null>(null);
  const { mutate: performAnalysis, isLoading, error } = useWhatIfAnalysis();

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleAssetSelect = (asset: AssetSearchResult | TrendingAsset) => {
    setSelectedAssetInfo(asset);
    setFormData(prev => ({
      ...prev,
      symbol: asset.symbol,
      // Auto-populate price if available
      price: 'current_price' in asset ? asset.current_price.toString() : prev.price
    }));
  };

  const handleSymbolChange = (symbol: string) => {
    setFormData(prev => ({
      ...prev,
      symbol: symbol
    }));
    // Clear selected asset info if user types manually
    if (selectedAssetInfo && selectedAssetInfo.symbol !== symbol) {
      setSelectedAssetInfo(null);
    }
  };

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.symbol || !formData.quantity || !formData.price) {
      return;
    }

    const analysisRequest = {
      action: formData.action,
      symbol: formData.symbol.toUpperCase(),
      quantity: parseFloat(formData.quantity),
      price: parseFloat(formData.price),
    };

    performAnalysis(analysisRequest, {
      onSuccess: (data) => {
        setAnalysisResult(data);
      },
    });
  };

  const resetForm = () => {
    setFormData({
      action: 'buy',
      symbol: '',
      quantity: '',
      price: '',
    });
    setAnalysisResult(null);
    setSelectedAssetInfo(null);
  };

  const handleClose = () => {
    resetForm();
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-4xl w-full max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex justify-between items-center p-6 border-b border-gray-200">
          <h2 className="text-2xl font-semibold text-gray-800">What-If Analysis</h2>
          <button
            onClick={handleClose}
            className="text-gray-500 hover:text-gray-700 text-2xl font-bold"
          >
            ×
          </button>
        </div>

        {/* Content */}
        <div className="p-6">
          {/* Form Section */}
          <div className="mb-6">
            <h3 className="text-lg font-medium text-gray-900 mb-4">Simulate Trade</h3>
            <form onSubmit={handleSubmit} className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                {/* Action Selection */}
                <div>
                  <label htmlFor="action" className="block text-sm font-medium text-gray-700 mb-1">
                    Action
                  </label>
                  <select
                    id="action"
                    name="action"
                    value={formData.action}
                    onChange={handleInputChange}
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  >
                    <option value="buy">Buy</option>
                    <option value="sell">Sell</option>
                  </select>
                </div>

                {/* Symbol Input */}
                <div>
                  <label htmlFor="symbol" className="block text-sm font-medium text-gray-700 mb-1">
                    Symbol
                  </label>
                  <AssetSelector
                    value={formData.symbol}
                    onChange={handleSymbolChange}
                    onAssetSelect={handleAssetSelect}
                    placeholder="Search for stocks, ETFs, crypto..."
                    showTrending={true}
                    enableDetailsModal={true}
                  />
                  {selectedAssetInfo && (
                    <div className="mt-2 p-2 bg-blue-50 rounded-md">
                      <p className="text-sm text-blue-800">
                        <strong>{selectedAssetInfo.symbol}</strong> - {selectedAssetInfo.name}
                        {selectedAssetInfo.exchange && (
                          <span className="text-blue-600 ml-2">({selectedAssetInfo.exchange})</span>
                        )}
                      </p>
                      {'current_price' in selectedAssetInfo && (
                        <p className="text-xs text-blue-600 mt-1">
                          Current Price: ${selectedAssetInfo.current_price?.toFixed(2)}
                        </p>
                      )}
                    </div>
                  )}
                </div>

                {/* Quantity Input */}
                <div>
                  <label htmlFor="quantity" className="block text-sm font-medium text-gray-700 mb-1">
                    Quantity
                  </label>
                  <input
                    type="number"
                    id="quantity"
                    name="quantity"
                    value={formData.quantity}
                    onChange={handleInputChange}
                    placeholder="Number of shares"
                    min="0.01"
                    step="0.01"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    required
                  />
                </div>

                {/* Price Input */}
                <div>
                  <label htmlFor="price" className="block text-sm font-medium text-gray-700 mb-1">
                    Price per Share ($)
                  </label>
                  <input
                    type="number"
                    id="price"
                    name="price"
                    value={formData.price}
                    onChange={handleInputChange}
                    placeholder="Price per share"
                    min="0.01"
                    step="0.01"
                    className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                    required
                  />
                </div>
              </div>

              {/* Submit Button */}
              <div className="flex justify-between items-center">
                <Button
                  type="submit"
                  disabled={isLoading || !formData.symbol || !formData.quantity || !formData.price}
                  className="flex items-center"
                >
                  {isLoading && <LoadingSpinner size="sm" className="mr-2" />}
                  {isLoading ? 'Analyzing...' : 'Run Analysis'}
                </Button>
                
                {analysisResult && (
                  <Button
                    type="button"
                    variant="secondary"
                    onClick={resetForm}
                  >
                    Reset
                  </Button>
                )}
              </div>
            </form>
          </div>

          {/* Error Display */}
          {error && (
            <div className="mb-6">
              <ErrorMessage message={error.message || 'Failed to perform analysis'} />
            </div>
          )}

          {/* Analysis Results */}
          {analysisResult && (
            <div className="space-y-6">
              <div className="border-t border-gray-200 pt-6">
                <h3 className="text-lg font-medium text-gray-900 mb-4">Analysis Results</h3>
                
                {/* Portfolio Impact Summary */}
                <div className="grid grid-cols-1 md:grid-cols-3 gap-4 mb-6">
                  <div className="bg-blue-50 p-4 rounded-lg border border-blue-200">
                    <div className="text-sm text-blue-600 font-medium">Value Change</div>
                    <div className={`text-xl font-bold ${
                      analysisResult.portfolio_impact.value_change >= 0 ? 'text-green-700' : 'text-red-700'
                    }`}>
                      {analysisResult.portfolio_impact.value_change >= 0 ? '+' : ''}
                      ${analysisResult.portfolio_impact.value_change.toLocaleString()}
                    </div>
                    <div className="text-xs text-blue-600">
                      ({((analysisResult.portfolio_impact.value_change / analysisResult.portfolio_impact.current_total_value) * 100).toFixed(2)}%)
                    </div>
                  </div>

                  <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
                    <div className="text-sm text-gray-600 font-medium">Current Value</div>
                    <div className="text-xl font-bold text-gray-900">
                      ${analysisResult.portfolio_impact.current_total_value.toLocaleString()}
                    </div>
                    <div className="text-xs text-gray-600">
                      {analysisResult.portfolio_impact.current_holdings} holdings
                    </div>
                  </div>

                  <div className="bg-green-50 p-4 rounded-lg border border-green-200">
                    <div className="text-sm text-green-600 font-medium">Projected Value</div>
                    <div className="text-xl font-bold text-green-700">
                      ${analysisResult.portfolio_impact.new_total_value.toLocaleString()}
                    </div>
                    <div className="text-xs text-green-600">
                      {analysisResult.portfolio_impact.new_holdings} holdings
                    </div>
                  </div>
                </div>

                {/* Trade Details & Position Impact */}
                <div className="mb-6">
                  <h4 className="text-md font-medium text-gray-900 mb-3">Trade Details</h4>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
                      <div className="text-sm text-gray-600 font-medium">Trade Action</div>
                      <div className="text-lg font-semibold text-gray-900 capitalize">
                        {analysisResult.trade_details.action} {analysisResult.trade_details.symbol}
                      </div>
                      <div className="text-sm text-gray-600 mt-1">
                        {analysisResult.trade_details.quantity} shares @ ${analysisResult.trade_details.price}
                      </div>
                      <div className="text-xs text-gray-500">
                        Total: ${analysisResult.trade_details.trade_value.toLocaleString()}
                      </div>
                    </div>

                    <div className="bg-purple-50 p-4 rounded-lg border border-purple-200">
                      <div className="text-sm text-purple-600 font-medium">Position Change</div>
                      <div className="text-lg font-semibold text-purple-700 capitalize">
                        {analysisResult.trade_details.position_change}
                      </div>
                      {analysisResult.position_impact.has_current_holding ? (
                        <div className="text-sm text-purple-600 mt-1">
                          {analysisResult.position_impact.current_quantity} → {analysisResult.position_impact.new_quantity} shares
                        </div>
                      ) : (
                        <div className="text-sm text-purple-600 mt-1">
                          New position: {analysisResult.position_impact.new_quantity} shares
                        </div>
                      )}
                      <div className="text-xs text-purple-500">
                        Avg Cost: ${analysisResult.position_impact.new_avg_cost}
                      </div>
                    </div>
                  </div>
                </div>

                {/* Expected Returns & Risk */}
                <div className="mb-6">
                  <h4 className="text-md font-medium text-gray-900 mb-3">Expected Returns & Risk</h4>
                  <div className="grid grid-cols-1 md:grid-cols-4 gap-4">
                    <div className="bg-green-50 p-3 rounded-lg border border-green-200">
                      <div className="text-sm text-green-600 font-medium">Annual Return</div>
                      <div className="text-lg font-semibold text-green-700">
                        {analysisResult.expected_returns.annual_return_estimate}%
                      </div>
                    </div>

                    <div className="bg-blue-50 p-3 rounded-lg border border-blue-200">
                      <div className="text-sm text-blue-600 font-medium">Risk Adjusted</div>
                      <div className="text-lg font-semibold text-blue-700">
                        {analysisResult.expected_returns.risk_adjusted_return}%
                      </div>
                    </div>

                    <div className="bg-yellow-50 p-3 rounded-lg border border-yellow-200">
                      <div className="text-sm text-yellow-600 font-medium">Volatility</div>
                      <div className="text-lg font-semibold text-yellow-700">
                        {analysisResult.expected_returns.symbol_volatility}%
                      </div>
                    </div>

                    <div className="bg-orange-50 p-3 rounded-lg border border-orange-200">
                      <div className="text-sm text-orange-600 font-medium">Risk Premium</div>
                      <div className="text-lg font-semibold text-orange-700">
                        {analysisResult.expected_returns.risk_premium}%
                      </div>
                    </div>
                  </div>
                </div>

                {/* Risk Impact */}
                <div className="mb-6">
                  <h4 className="text-md font-medium text-gray-900 mb-3">Portfolio Risk Impact</h4>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="bg-red-50 p-4 rounded-lg border border-red-200">
                      <div className="text-sm text-red-600 font-medium">Concentration Change</div>
                      <div className={`text-lg font-semibold ${
                        analysisResult.risk_impact.concentration_change >= 0 ? 'text-red-700' : 'text-green-700'
                      }`}>
                        {analysisResult.risk_impact.concentration_change >= 0 ? '+' : ''}
                        {analysisResult.risk_impact.concentration_change.toFixed(2)}%
                      </div>
                      <div className="text-xs text-red-600 capitalize">
                        {analysisResult.risk_impact.diversification_impact}
                      </div>
                    </div>

                    <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
                      <div className="text-sm text-gray-600 font-medium">Diversification</div>
                      <div className="text-lg font-semibold text-gray-700 capitalize">
                        {analysisResult.risk_impact.diversification_impact}
                      </div>
                    </div>
                  </div>
                </div>

                {/* Allocation Changes */}
                <div className="mb-6">
                  <h4 className="text-md font-medium text-gray-900 mb-3">Allocation Changes</h4>
                  <div className="space-y-2">
                    {Object.entries(analysisResult.allocation_impact).map(([assetType, impact]: [string, any]) => (
                      <div key={assetType} className="flex justify-between items-center p-3 bg-gray-50 rounded-lg">
                        <span className="font-medium text-gray-900">{assetType}</span>
                        <div className="text-right">
                          <div className="text-sm text-gray-600">
                            {impact.current_percent.toFixed(1)}% → {impact.new_percent.toFixed(1)}%
                          </div>
                          <div className={`text-sm font-medium ${
                            impact.change >= 0 ? 'text-green-600' : 'text-red-600'
                          }`}>
                            {impact.change >= 0 ? '+' : ''}{impact.change.toFixed(1)}%
                          </div>
                        </div>
                      </div>
                    ))}
                  </div>
                </div>

                {/* Recommendations */}
                {analysisResult.recommendations && analysisResult.recommendations.length > 0 && (
                  <div>
                    <h4 className="text-md font-medium text-gray-900 mb-3">Recommendations</h4>
                    <div className="bg-blue-50 border border-blue-200 rounded-lg p-4">
                      <ul className="space-y-2">
                        {analysisResult.recommendations.map((recommendation: string, index: number) => (
                          <li key={index} className="flex items-start">
                            <span className="text-blue-500 mr-2">•</span>
                            <span className="text-blue-800 text-sm">{recommendation}</span>
                          </li>
                        ))}
                      </ul>
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex justify-end p-6 border-t border-gray-200">
          <Button variant="secondary" onClick={handleClose}>
            Close
          </Button>
        </div>
      </div>
    </div>
  );
};

export default WhatIfAnalysisModal;
