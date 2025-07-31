import { Holding, PortfolioSummaryResponse, AssetAllocation } from '../types/portfolio';
import { PerformanceAnalytics, RiskMetrics, AssetAllocation as AnalyticsAssetAllocation } from '../hooks/useAnalytics';

// Export formats
export type ExportFormat = 'csv' | 'json';

// Export data types
export interface ExportData {
  holdings: Holding[];
  summary?: PortfolioSummaryResponse;
  analytics?: {
    performance?: PerformanceAnalytics;
    risk?: RiskMetrics;
    allocation?: AnalyticsAssetAllocation;
  };
  exportDate: string;
}

// Convert data to CSV format
export const convertToCSV = (data: any[], headers: string[]): string => {
  const csvHeaders = headers.join(',');
  const csvRows = data.map(row => 
    headers.map(header => {
      const value = row[header];
      // Handle values that might contain commas or quotes
      if (typeof value === 'string' && (value.includes(',') || value.includes('"'))) {
        return `"${value.replace(/"/g, '""')}"`;
      }
      return value ?? '';
    }).join(',')
  );
  
  return [csvHeaders, ...csvRows].join('\n');
};

// Export holdings as CSV
export const exportHoldingsCSV = (holdings: Holding[]): string => {
  const headers = [
    'symbol',
    'name', 
    'asset_type',
    'quantity',
    'average_cost',
    'total_cost',
    'purchase_date'
  ];
  
  const data = holdings.map(holding => ({
    symbol: holding.symbol,
    name: holding.name,
    asset_type: holding.asset_type,
    quantity: holding.quantity,
    average_cost: holding.average_cost.toFixed(2),
    total_cost: (holding.quantity * holding.average_cost).toFixed(2),
    purchase_date: holding.purchase_date
  }));
  
  return convertToCSV(data, headers);
};

// Export portfolio summary as CSV
export const exportSummaryCSV = (summary: PortfolioSummaryResponse): string => {
  const summaryData = [{
    total_holdings: summary.summary.total_holdings,
    total_cost: summary.summary.total_cost?.toFixed(2) || '0.00',
    total_market_value: summary.summary.total_market_value?.toFixed(2) || '0.00',
    daily_change: summary.summary.daily_change?.toFixed(2) || '0.00',
    daily_change_percent: summary.summary.daily_change_percent?.toFixed(2) || '0.00',
    unrealized_gain_loss: summary.summary.unrealized_gain_loss?.toFixed(2) || '0.00',
    unrealized_gain_loss_percent: summary.summary.unrealized_gain_loss_percent?.toFixed(2) || '0.00'
  }];
  
  const headers = Object.keys(summaryData[0]);
  return convertToCSV(summaryData, headers);
};

// Export asset allocation as CSV
export const exportAllocationCSV = (allocation: AnalyticsAssetAllocation): string => {
  const headers = ['type', 'category', 'count', 'value', 'percentage'];
  
  const assetTypeData = allocation.by_asset_type.map((item: any) => ({
    type: 'Asset Type',
    category: item.asset_type,
    count: item.count,
    value: item.value.toFixed(2),
    percentage: item.percentage.toFixed(2)
  }));
  
  const sectorData = allocation.by_sector.map((item: any) => ({
    type: 'Sector',
    category: item.sector,
    count: item.count,
    value: item.value.toFixed(2),
    percentage: item.percentage.toFixed(2)
  }));
  
  const allData = [...assetTypeData, ...sectorData];
  return convertToCSV(allData, headers);
};

// Export performance analytics as CSV
export const exportPerformanceCSV = (performance: PerformanceAnalytics): string => {
  const headers = [
    'symbol',
    'name',
    'quantity',
    'average_cost',
    'current_price',
    'total_cost',
    'current_value',
    'gain_loss',
    'gain_loss_percent'
  ];
  
  const data = performance.top_performers.map((performer: any) => ({
    symbol: performer.symbol,
    name: performer.name,
    quantity: performer.quantity,
    average_cost: performer.average_cost.toFixed(2),
    current_price: performer.current_price.toFixed(2),
    total_cost: performer.total_cost.toFixed(2),
    current_value: performer.current_value.toFixed(2),
    gain_loss: performer.gain_loss.toFixed(2),
    gain_loss_percent: performer.gain_loss_percent.toFixed(2)
  }));
  
  return convertToCSV(data, headers);
};

// Export risk metrics as CSV
export const exportRiskCSV = (risk: RiskMetrics): string => {
  const riskData = [{
    overall_risk_level: risk.risk_assessment.overall_risk_level,
    concentration_risk: risk.risk_assessment.concentration_risk,
    herfindahl_index: risk.risk_assessment.herfindahl_index.toFixed(4),
    diversification_score: risk.risk_assessment.diversification_score.toFixed(2),
    portfolio_beta: risk.volatility_metrics.portfolio_beta.toFixed(2),
    sharpe_ratio: risk.volatility_metrics.sharpe_ratio.toFixed(2),
    max_drawdown: risk.volatility_metrics.max_drawdown.toFixed(2),
    var_95: risk.volatility_metrics.var_95.toFixed(2),
    expected_volatility: risk.volatility_metrics.expected_volatility.toFixed(2),
    portfolio_return: risk.volatility_metrics.portfolio_return.toFixed(2)
  }];
  
  const headers = Object.keys(riskData[0]);
  return convertToCSV(riskData, headers);
};

// Generate filename with timestamp
export const generateFilename = (prefix: string, format: ExportFormat): string => {
  const timestamp = new Date().toISOString().split('T')[0]; // YYYY-MM-DD
  return `${prefix}_${timestamp}.${format}`;
};

// Download file
export const downloadFile = (content: string, filename: string, mimeType: string): void => {
  const blob = new Blob([content], { type: mimeType });
  const url = URL.createObjectURL(blob);
  
  const link = document.createElement('a');
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  
  // Cleanup
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
};

// Main export function
export const exportPortfolioData = (
  data: ExportData,
  format: ExportFormat,
  exportType: 'complete' | 'holdings' | 'summary' | 'analytics'
): void => {
  let content: string;
  let filename: string;
  let mimeType: string;
  
  if (format === 'json') {
    mimeType = 'application/json';
    
    switch (exportType) {
      case 'complete':
        content = JSON.stringify(data, null, 2);
        filename = generateFilename('portfolio_complete', format);
        break;
      case 'holdings':
        content = JSON.stringify({ holdings: data.holdings, exportDate: data.exportDate }, null, 2);
        filename = generateFilename('portfolio_holdings', format);
        break;
      case 'summary':
        content = JSON.stringify({ summary: data.summary, exportDate: data.exportDate }, null, 2);
        filename = generateFilename('portfolio_summary', format);
        break;
      case 'analytics':
        content = JSON.stringify({ analytics: data.analytics, exportDate: data.exportDate }, null, 2);
        filename = generateFilename('portfolio_analytics', format);
        break;
    }
  } else {
    mimeType = 'text/csv';
    
    switch (exportType) {
      case 'complete':
        // For CSV, we'll create a zip-like structure with multiple sections
        const sections = [];
        sections.push('=== HOLDINGS ===');
        sections.push(exportHoldingsCSV(data.holdings));
        
        if (data.summary) {
          sections.push('\n=== SUMMARY ===');
          sections.push(exportSummaryCSV(data.summary));
        }
        
        if (data.analytics?.allocation) {
          sections.push('\n=== ALLOCATION ===');
          sections.push(exportAllocationCSV(data.analytics.allocation));
        }
        
        if (data.analytics?.performance) {
          sections.push('\n=== PERFORMANCE ===');
          sections.push(exportPerformanceCSV(data.analytics.performance));
        }
        
        if (data.analytics?.risk) {
          sections.push('\n=== RISK METRICS ===');
          sections.push(exportRiskCSV(data.analytics.risk));
        }
        
        content = sections.join('\n');
        filename = generateFilename('portfolio_complete', format);
        break;
        
      case 'holdings':
        content = exportHoldingsCSV(data.holdings);
        filename = generateFilename('portfolio_holdings', format);
        break;
        
      case 'summary':
        if (!data.summary) {
          throw new Error('Summary data not available');
        }
        content = exportSummaryCSV(data.summary);
        filename = generateFilename('portfolio_summary', format);
        break;
        
      case 'analytics':
        if (!data.analytics) {
          throw new Error('Analytics data not available');
        }
        
        const analyticsSections = [];
        if (data.analytics.allocation) {
          analyticsSections.push('=== ALLOCATION ===');
          analyticsSections.push(exportAllocationCSV(data.analytics.allocation));
        }
        
        if (data.analytics.performance) {
          analyticsSections.push('\n=== PERFORMANCE ===');
          analyticsSections.push(exportPerformanceCSV(data.analytics.performance));
        }
        
        if (data.analytics.risk) {
          analyticsSections.push('\n=== RISK METRICS ===');
          analyticsSections.push(exportRiskCSV(data.analytics.risk));
        }
        
        content = analyticsSections.join('\n');
        filename = generateFilename('portfolio_analytics', format);
        break;
    }
  }
  
  downloadFile(content, filename, mimeType);
};