import React from 'react';
import { useRiskMetrics, usePerformanceAnalytics } from '../hooks/useAnalytics';

const AnalyticsDashboard: React.FC = () => {
  const { data: riskData, isLoading: riskLoading, error: riskError } = useRiskMetrics();
  const { data: performanceData, isLoading: performanceLoading, error: performanceError } = usePerformanceAnalytics();

  const isLoading = riskLoading || performanceLoading;
  const hasError = riskError || performanceError;

  if (isLoading) {
    return (
      <div className="bg-white rounded-lg shadow-md p-6 mb-6">
        <div className="flex items-center justify-center h-64">
          <div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>
          <span className="ml-3 text-gray-600">Loading analytics...</span>
        </div>
      </div>
    );
  }

  if (hasError) {
    return (
      <div className="bg-white rounded-lg shadow-md p-6 mb-6">
        <div className="bg-red-50 border border-red-200 rounded-lg p-4">
          <h3 className="text-red-800 font-medium">Error Loading Analytics</h3>
          <p className="text-red-600 text-sm mt-1">
            {riskError?.message || performanceError?.message || 'Failed to load analytics data'}
          </p>
        </div>
      </div>
    );
  }

  // Calculate top 3 holdings percentage for concentration risk
  const top3Percentage = performanceData?.top_performers
    ?.slice(0, 3)
    .reduce((sum, performer) => {
      const percentage = (performer.current_value / performanceData.portfolio_performance.current_value) * 100;
      return sum + percentage;
    }, 0) || 0;

  return (
    <div className="bg-white rounded-lg shadow-md p-6 mb-6">
      <h2 className="text-2xl font-semibold text-gray-800 mb-6">Portfolio Analytics</h2>
      
      <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
        {/* Risk Metrics Cards */}
        <div className="space-y-4">
          <h3 className="text-lg font-semibold text-gray-900">Risk Metrics</h3>
          
          {/* Portfolio Beta Card */}
          <div className="bg-blue-50 p-4 rounded-lg border border-blue-200">
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium text-blue-900">Portfolio Beta</span>
              <span className="text-lg font-bold text-blue-700">
                {riskData?.volatility_metrics.portfolio_beta?.toFixed(2) || '--'}
              </span>
            </div>
            <p className="text-xs text-blue-600 mt-1">Market correlation measure</p>
          </div>

          {/* Sharpe Ratio Card */}
          <div className="bg-green-50 p-4 rounded-lg border border-green-200">
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium text-green-900">Sharpe Ratio</span>
              <span className="text-lg font-bold text-green-700">
                {riskData?.volatility_metrics.sharpe_ratio?.toFixed(2) || '--'}
              </span>
            </div>
            <p className="text-xs text-green-600 mt-1">Risk-adjusted returns</p>
          </div>

          {/* Value at Risk Card */}
          <div className="bg-red-50 p-4 rounded-lg border border-red-200">
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium text-red-900">Value at Risk</span>
              <span className="text-lg font-bold text-red-700">
                {riskData?.volatility_metrics.var_95 ? `${riskData.volatility_metrics.var_95.toFixed(1)}%` : '--'}
              </span>
            </div>
            <p className="text-xs text-red-600 mt-1">95% confidence level</p>
          </div>
        </div>

        {/* Performance Metrics */}
        <div className="space-y-4">
          <h3 className="text-lg font-semibold text-gray-900">Performance</h3>
          
          <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium text-gray-900">Total Return</span>
              <span className={`text-lg font-bold ${
                (performanceData?.portfolio_performance.total_return_percent || 0) >= 0
                  ? 'text-green-700'
                  : 'text-red-700'
              }`}>
                {performanceData?.portfolio_performance.total_return_percent ? `${performanceData.portfolio_performance.total_return_percent.toFixed(2)}%` : '--'}
              </span>
            </div>
            <p className="text-xs text-gray-600 mt-1">Overall portfolio performance</p>
          </div>

          <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium text-gray-900">Volatility</span>
              <span className="text-lg font-bold text-gray-700">
                {riskData?.volatility_metrics.expected_volatility ? `${riskData.volatility_metrics.expected_volatility.toFixed(1)}%` : '--'}
              </span>
            </div>
            <p className="text-xs text-gray-600 mt-1">Portfolio volatility estimate</p>
          </div>

          <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium text-gray-900">Max Drawdown</span>
              <span className="text-lg font-bold text-red-700">
                {riskData?.volatility_metrics.max_drawdown ? `${riskData.volatility_metrics.max_drawdown.toFixed(1)}%` : '--'}
              </span>
            </div>
            <p className="text-xs text-gray-600 mt-1">Estimated maximum loss</p>
          </div>
        </div>

        {/* Diversification Analysis */}
        <div className="space-y-4">
          <h3 className="text-lg font-semibold text-gray-900">Diversification</h3>
          
          <div className="bg-yellow-50 p-4 rounded-lg border border-yellow-200">
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium text-yellow-900">Sector Count</span>
              <span className="text-lg font-bold text-yellow-700">
                {riskData?.sector_diversification?.length || '--'}
              </span>
            </div>
            <p className="text-xs text-yellow-600 mt-1">Number of sectors</p>
          </div>

          <div className="bg-yellow-50 p-4 rounded-lg border border-yellow-200">
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium text-yellow-900">HHI Index</span>
              <span className="text-lg font-bold text-yellow-700">
                {riskData?.risk_assessment.herfindahl_index?.toFixed(3) || '--'}
              </span>
            </div>
            <p className="text-xs text-yellow-600 mt-1">Concentration measure</p>
          </div>

          <div className="bg-yellow-50 p-4 rounded-lg border border-yellow-200">
            <div className="flex justify-between items-center">
              <span className="text-sm font-medium text-yellow-900">Top 3 Weight</span>
              <span className="text-lg font-bold text-yellow-700">
                {top3Percentage.toFixed(1)}%
              </span>
            </div>
            <p className="text-xs text-yellow-600 mt-1">Concentration risk</p>
          </div>
        </div>
      </div>

      {/* Risk Assessment Summary */}
      {riskData && (
        <div className="mt-6 p-4 bg-gray-50 rounded-lg">
          <h4 className="font-medium text-gray-900 mb-2">Risk Assessment</h4>
          <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-sm">
            <div>
              <span className="font-medium">Risk Level: </span>
              <span className={`px-2 py-1 rounded text-xs font-medium ${
                riskData.risk_assessment.overall_risk_level === 'High' 
                  ? 'bg-red-100 text-red-800'
                  : riskData.risk_assessment.overall_risk_level === 'Medium'
                  ? 'bg-yellow-100 text-yellow-800'
                  : 'bg-green-100 text-green-800'
              }`}>
                {riskData.risk_assessment.overall_risk_level}
              </span>
            </div>
            <div>
              <span className="font-medium">Diversification Score: </span>
              <span>{riskData.risk_assessment.diversification_score?.toFixed(1)}%</span>
            </div>
          </div>
          <p className="text-sm text-gray-600 mt-2">
            {riskData.risk_assessment.concentration_risk}
          </p>
        </div>
      )}

      {/* Warnings */}
      {performanceData?.warnings && performanceData.warnings.length > 0 && (
        <div className="mt-4 p-3 bg-yellow-50 border border-yellow-200 rounded-lg">
          <h4 className="text-sm font-medium text-yellow-800 mb-1">Data Warnings</h4>
          <ul className="text-xs text-yellow-700 space-y-1">
            {performanceData.warnings.map((warning, index) => (
              <li key={index}>• {warning}</li>
            ))}
          </ul>
        </div>
      )}
    </div>
  );
};

export default AnalyticsDashboard;