'use client'

import React, { useState } from 'react';
import { usePortfolioExport } from '../hooks/usePortfolioExport';
import { ExportFormat } from '../lib/exportUtils';

interface ExportModalProps {
  isOpen: boolean;
  onClose: () => void;
}

const ExportModal: React.FC<ExportModalProps> = ({ isOpen, onClose }) => {
  const [selectedFormat, setSelectedFormat] = useState<ExportFormat>('csv');
  const [selectedType, setSelectedType] = useState<'complete' | 'holdings' | 'summary' | 'analytics'>('complete');
  const [includeAnalytics, setIncludeAnalytics] = useState(true);
  const [includeSummary, setIncludeSummary] = useState(true);

  const {
    exportData,
    isExporting,
    exportError,
    dataStatus
  } = usePortfolioExport();

  if (!isOpen) return null;

  const handleExport = async () => {
    try {
      await exportData(selectedFormat, selectedType, {
        includeAnalytics: selectedType === 'complete' ? includeAnalytics : undefined,
        includeSummary: selectedType === 'complete' ? includeSummary : undefined
      });
      
      // Close modal after successful export
      setTimeout(() => {
        onClose();
      }, 1000);
    } catch (error) {
      console.error('Export failed:', error);
    }
  };

  const getExportDescription = () => {
    switch (selectedType) {
      case 'complete':
        return 'Export all portfolio data including holdings, summary, and analytics';
      case 'holdings':
        return 'Export only your portfolio holdings data';
      case 'summary':
        return 'Export portfolio summary and performance metrics';
      case 'analytics':
        return 'Export detailed analytics including risk metrics and performance data';
      default:
        return '';
    }
  };

  const isDataAvailable = () => {
    switch (selectedType) {
      case 'holdings':
        return dataStatus.holdings.available;
      case 'summary':
        return dataStatus.summary.available;
      case 'analytics':
        return dataStatus.analytics.risk.available || 
               dataStatus.analytics.performance.available || 
               dataStatus.analytics.allocation.available;
      case 'complete':
        return dataStatus.holdings.available;
      default:
        return false;
    }
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-xl max-w-md w-full mx-4">
        <div className="p-6">
          {/* Header */}
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-xl font-semibold text-gray-900">Export Portfolio Data</h2>
            <button
              onClick={onClose}
              className="text-gray-400 hover:text-gray-600 transition-colors"
              disabled={isExporting}
            >
              <svg className="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </button>
          </div>

          {/* Export Type Selection */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-3">
              What would you like to export?
            </label>
            <div className="space-y-2">
              {[
                { value: 'complete', label: 'Complete Portfolio', available: dataStatus.holdings.available },
                { value: 'holdings', label: 'Holdings Only', available: dataStatus.holdings.available },
                { value: 'summary', label: 'Summary Only', available: dataStatus.summary.available },
                { value: 'analytics', label: 'Analytics Only', available: dataStatus.analytics.risk.available || dataStatus.analytics.performance.available || dataStatus.analytics.allocation.available }
              ].map((option) => (
                <label key={option.value} className="flex items-center">
                  <input
                    type="radio"
                    name="exportType"
                    value={option.value}
                    checked={selectedType === option.value}
                    onChange={(e) => setSelectedType(e.target.value as any)}
                    disabled={!option.available}
                    className="mr-3 text-blue-600 focus:ring-blue-500"
                  />
                  <span className={`${option.available ? 'text-gray-900' : 'text-gray-400'}`}>
                    {option.label}
                    {!option.available && ' (No data available)'}
                  </span>
                </label>
              ))}
            </div>
            <p className="text-sm text-gray-600 mt-2">
              {getExportDescription()}
            </p>
          </div>

          {/* Additional Options for Complete Export */}
          {selectedType === 'complete' && (
            <div className="mb-6">
              <label className="block text-sm font-medium text-gray-700 mb-3">
                Additional Options
              </label>
              <div className="space-y-2">
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={includeSummary}
                    onChange={(e) => setIncludeSummary(e.target.checked)}
                    disabled={!dataStatus.summary.available}
                    className="mr-3 text-blue-600 focus:ring-blue-500"
                  />
                  <span className={dataStatus.summary.available ? 'text-gray-900' : 'text-gray-400'}>
                    Include Portfolio Summary
                    {!dataStatus.summary.available && ' (No data available)'}
                  </span>
                </label>
                <label className="flex items-center">
                  <input
                    type="checkbox"
                    checked={includeAnalytics}
                    onChange={(e) => setIncludeAnalytics(e.target.checked)}
                    disabled={!dataStatus.analytics.risk.available && !dataStatus.analytics.performance.available && !dataStatus.analytics.allocation.available}
                    className="mr-3 text-blue-600 focus:ring-blue-500"
                  />
                  <span className={dataStatus.analytics.risk.available || dataStatus.analytics.performance.available || dataStatus.analytics.allocation.available ? 'text-gray-900' : 'text-gray-400'}>
                    Include Analytics Data
                    {!dataStatus.analytics.risk.available && !dataStatus.analytics.performance.available && !dataStatus.analytics.allocation.available && ' (No data available)'}
                  </span>
                </label>
              </div>
            </div>
          )}

          {/* Format Selection */}
          <div className="mb-6">
            <label className="block text-sm font-medium text-gray-700 mb-3">
              Export Format
            </label>
            <div className="flex space-x-4">
              <label className="flex items-center">
                <input
                  type="radio"
                  name="format"
                  value="csv"
                  checked={selectedFormat === 'csv'}
                  onChange={(e) => setSelectedFormat(e.target.value as ExportFormat)}
                  className="mr-2 text-blue-600 focus:ring-blue-500"
                />
                <span className="text-gray-900">CSV</span>
              </label>
              <label className="flex items-center">
                <input
                  type="radio"
                  name="format"
                  value="json"
                  checked={selectedFormat === 'json'}
                  onChange={(e) => setSelectedFormat(e.target.value as ExportFormat)}
                  className="mr-2 text-blue-600 focus:ring-blue-500"
                />
                <span className="text-gray-900">JSON</span>
              </label>
            </div>
            <p className="text-sm text-gray-600 mt-1">
              {selectedFormat === 'csv' 
                ? 'Comma-separated values format, suitable for Excel and other spreadsheet applications'
                : 'JavaScript Object Notation format, suitable for data analysis and backup purposes'
              }
            </p>
          </div>

          {/* Data Status */}
          <div className="mb-6 p-3 bg-gray-50 rounded-lg">
            <h4 className="text-sm font-medium text-gray-700 mb-2">Data Availability</h4>
            <div className="text-xs space-y-1">
              <div className="flex justify-between">
                <span>Holdings:</span>
                <span className={dataStatus.holdings.available ? 'text-green-600' : 'text-red-600'}>
                  {dataStatus.holdings.available ? `${dataStatus.holdings.count} items` : 'No data'}
                </span>
              </div>
              <div className="flex justify-between">
                <span>Summary:</span>
                <span className={dataStatus.summary.available ? 'text-green-600' : 'text-red-600'}>
                  {dataStatus.summary.available ? 'Available' : 'No data'}
                </span>
              </div>
              <div className="flex justify-between">
                <span>Analytics:</span>
                <span className={dataStatus.analytics.risk.available || dataStatus.analytics.performance.available || dataStatus.analytics.allocation.available ? 'text-green-600' : 'text-red-600'}>
                  {dataStatus.analytics.risk.available || dataStatus.analytics.performance.available || dataStatus.analytics.allocation.available ? 'Available' : 'No data'}
                </span>
              </div>
            </div>
          </div>

          {/* Error Display */}
          {exportError && (
            <div className="mb-4 p-3 bg-red-50 border border-red-200 rounded-lg">
              <p className="text-sm text-red-700">{exportError}</p>
            </div>
          )}

          {/* Action Buttons */}
          <div className="flex justify-end space-x-3">
            <button
              onClick={onClose}
              disabled={isExporting}
              className="px-4 py-2 text-gray-700 bg-gray-200 hover:bg-gray-300 rounded-md font-medium transition-colors disabled:opacity-50"
            >
              Cancel
            </button>
            <button
              onClick={handleExport}
              disabled={isExporting || !isDataAvailable()}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-md font-medium transition-colors disabled:opacity-50 flex items-center"
            >
              {isExporting ? (
                <>
                  <svg className="animate-spin -ml-1 mr-2 h-4 w-4 text-white" fill="none" viewBox="0 0 24 24">
                    <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
                    <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
                  </svg>
                  Exporting...
                </>
              ) : (
                <>
                  <svg className="w-4 h-4 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                    <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
                  </svg>
                  Export
                </>
              )}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ExportModal;