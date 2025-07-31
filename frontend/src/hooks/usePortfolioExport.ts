import { useState } from 'react';
import { usePortfolioHoldings } from './usePortfolio';
import { useQuery } from 'react-query';
import { useRiskMetrics, usePerformanceAnalytics, useAssetAllocation } from './useAnalytics';
import apiClient from '../lib/api';
import { PortfolioSummaryResponse, QUERY_KEYS } from '../types/portfolio';
import { ExportData, ExportFormat, exportPortfolioData } from '../lib/exportUtils';

// Hook to fetch portfolio summary
const usePortfolioSummary = () => {
  return useQuery<PortfolioSummaryResponse, Error>(
    QUERY_KEYS.PORTFOLIO_SUMMARY,
    async () => {
      const { data } = await apiClient.get('/portfolio/summary');
      return data;
    },
    {
      staleTime: 5 * 60 * 1000, // 5 minutes cache
    }
  );
};

export const usePortfolioExport = () => {
  const [isExporting, setIsExporting] = useState(false);
  const [exportError, setExportError] = useState<string | null>(null);

  // Fetch all required data
  const { data: holdings, isLoading: holdingsLoading, error: holdingsError } = usePortfolioHoldings();
  const { data: summary, isLoading: summaryLoading, error: summaryError } = usePortfolioSummary();
  const { data: riskMetrics, isLoading: riskLoading, error: riskError } = useRiskMetrics();
  const { data: performanceAnalytics, isLoading: performanceLoading, error: performanceError } = usePerformanceAnalytics();
  const { data: assetAllocation, isLoading: allocationLoading, error: allocationError } = useAssetAllocation();

  // Check if any data is loading
  const isLoading = holdingsLoading || summaryLoading || riskLoading || performanceLoading || allocationLoading;

  // Check for errors
  const hasError = holdingsError || summaryError || riskError || performanceError || allocationError;

  // Export function
  const exportData = async (
    format: ExportFormat,
    exportType: 'complete' | 'holdings' | 'summary' | 'analytics',
    options?: {
      includeAnalytics?: boolean;
      includeSummary?: boolean;
    }
  ) => {
    setIsExporting(true);
    setExportError(null);

    try {
      // Validate required data based on export type
      if (!holdings?.holdings) {
        throw new Error('Holdings data is not available');
      }

      if (exportType === 'summary' && !summary) {
        throw new Error('Summary data is not available');
      }

      if (exportType === 'analytics' && !riskMetrics && !performanceAnalytics && !assetAllocation) {
        throw new Error('Analytics data is not available');
      }

      // Prepare export data
      const exportData: ExportData = {
        holdings: holdings.holdings,
        exportDate: new Date().toISOString(),
      };

      // Add summary if available and requested
      if (summary && (exportType === 'complete' || exportType === 'summary' || options?.includeSummary)) {
        exportData.summary = summary;
      }

      // Add analytics if available and requested
      if ((exportType === 'complete' || exportType === 'analytics' || options?.includeAnalytics)) {
        exportData.analytics = {};
        
        if (riskMetrics) {
          exportData.analytics.risk = riskMetrics;
        }
        
        if (performanceAnalytics) {
          exportData.analytics.performance = performanceAnalytics;
        }
        
        if (assetAllocation) {
          exportData.analytics.allocation = assetAllocation;
        }
      }

      // Perform the export
      exportPortfolioData(exportData, format, exportType);

    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'An unknown error occurred during export';
      setExportError(errorMessage);
      console.error('Export error:', error);
    } finally {
      setIsExporting(false);
    }
  };

  // Quick export functions for convenience
  const exportHoldings = (format: ExportFormat) => exportData(format, 'holdings');
  const exportSummary = (format: ExportFormat) => exportData(format, 'summary');
  const exportAnalytics = (format: ExportFormat) => exportData(format, 'analytics');
  const exportComplete = (format: ExportFormat) => exportData(format, 'complete');

  // Get available data status
  const getDataStatus = () => {
    return {
      holdings: {
        available: !!holdings?.holdings?.length,
        count: holdings?.holdings?.length || 0,
        loading: holdingsLoading,
        error: holdingsError
      },
      summary: {
        available: !!summary,
        loading: summaryLoading,
        error: summaryError
      },
      analytics: {
        risk: {
          available: !!riskMetrics,
          loading: riskLoading,
          error: riskError
        },
        performance: {
          available: !!performanceAnalytics,
          loading: performanceLoading,
          error: performanceError
        },
        allocation: {
          available: !!assetAllocation,
          loading: allocationLoading,
          error: allocationError
        }
      }
    };
  };

  return {
    // Export functions
    exportData,
    exportHoldings,
    exportSummary,
    exportAnalytics,
    exportComplete,
    
    // Status
    isLoading,
    isExporting,
    hasError,
    exportError,
    
    // Data availability
    dataStatus: getDataStatus(),
    
    // Raw data (for preview purposes)
    data: {
      holdings: holdings?.holdings,
      summary,
      analytics: {
        risk: riskMetrics,
        performance: performanceAnalytics,
        allocation: assetAllocation
      }
    }
  };
};