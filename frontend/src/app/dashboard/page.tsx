'use client'

import { useState } from 'react'
import Link from 'next/link'
import PortfolioSummary from '@/components/PortfolioSummary'
import HoldingsTable from '@/components/HoldingsTable'
import AddHoldingForm from '@/components/AddHoldingForm'
import PortfolioChart from '@/components/PortfolioChart'
import AnalyticsDashboard from '@/components/AnalyticsDashboard'
import WhatIfAnalysisModal from '@/components/WhatIfAnalysisModal'
import TransactionForm from '@/components/TransactionForm'
import ExportModal from '@/components/ExportModal'
import NotificationCenter from '@/components/NotificationCenter'
import AssetSearch from '@/components/AssetSearch'
import { usePortfolioWebSocket } from '@/hooks/useWebSocket'

// Global WebSocket Status Component
function WebSocketStatus({ isConnected, connectionStatus }: { isConnected: boolean; connectionStatus: string }) {
  const getStatusInfo = () => {
    switch (connectionStatus) {
      case 'connected':
        return { color: 'bg-green-500', text: 'Connected', pulse: true }
      case 'connecting':
        return { color: 'bg-yellow-500', text: 'Connecting...', pulse: true }
      case 'error':
        return { color: 'bg-red-500', text: 'Connection Error', pulse: false }
      default:
        return { color: 'bg-gray-500', text: 'Disconnected', pulse: false }
    }
  }

  const status = getStatusInfo()

  return (
    <div className="flex items-center space-x-2 bg-white px-3 py-2 rounded-lg shadow-sm border">
      <div className={`w-2 h-2 rounded-full ${status.color} ${status.pulse ? 'animate-pulse' : ''}`}></div>
      <span className="text-sm font-medium text-gray-700">{status.text}</span>
      {isConnected && (
        <span className="text-xs text-green-600 bg-green-50 px-2 py-1 rounded-full">
          Real-time
        </span>
      )}
    </div>
  )
}

export default function DashboardPage() {
  const [isWhatIfModalOpen, setIsWhatIfModalOpen] = useState(false);
  const [isTransactionFormOpen, setIsTransactionFormOpen] = useState(false);
  const [isExportModalOpen, setIsExportModalOpen] = useState(false);
  
  // Initialize WebSocket connection for the entire dashboard
  const webSocket = usePortfolioWebSocket(true);

  const handleAssetSelect = (symbol: string, name: string) => {
    // For now, we'll just log the selection
    // This could open a form to add to portfolio or trigger another action
    console.log(`Selected asset: ${symbol} - ${name}`);
    // You could set state to open AddHoldingForm with pre-filled data
  };

  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        {/* Header with WebSocket Status */}
        <div className="flex justify-between items-center mb-6">
          <div className="flex items-center">
            <Link
              href="/"
              className="text-gray-600 hover:text-gray-900 mr-4 flex items-center transition-colors"
            >
              <svg className="w-5 h-5 mr-1" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M15 19l-7-7 7-7" />
              </svg>
              Back
            </Link>
            <h1 className="text-3xl font-bold text-gray-900">Portfolio Dashboard</h1>
          </div>
          
          {/* WebSocket Status */}
          <WebSocketStatus
            isConnected={webSocket.isConnected}
            connectionStatus={webSocket.connectionStatus}
          />
        </div>

        {/* Action Buttons */}
        <div className="flex justify-end space-x-3 mb-8">
            {/* Export Button */}
            <button
              onClick={() => setIsExportModalOpen(true)}
              className="bg-indigo-600 hover:bg-indigo-700 text-white px-4 py-2 rounded-md font-medium transition-colors flex items-center"
            >
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 10v6m0 0l-3-3m3 3l3-3m2 8H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z" />
              </svg>
              Export Data
            </button>
            
            {/* Notification Center */}
            <NotificationCenter />
            
            {/* Market Data Link */}
            <Link
              href="/market-data"
              className="bg-purple-600 hover:bg-purple-700 text-white px-4 py-2 rounded-md font-medium transition-colors flex items-center"
            >
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M16 8v8m-4-5v5m-4-2v2m-2 4h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
              </svg>
              Market Data
            </Link>
            
            {/* Add Transaction Button */}
            <button
              onClick={() => setIsTransactionFormOpen(true)}
              className="bg-green-600 hover:bg-green-700 text-white px-4 py-2 rounded-md font-medium transition-colors flex items-center"
            >
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
              </svg>
              Add Transaction
            </button>
            
            {/* What-If Analysis Button */}
            <button
              onClick={() => setIsWhatIfModalOpen(true)}
              className="bg-blue-600 hover:bg-blue-700 text-white px-4 py-2 rounded-md font-medium transition-colors flex items-center"
            >
              <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
              </svg>
              What-If Analysis
            </button>
        </div>
        
        {/* Responsive grid layout */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
          {/* Portfolio Summary - spans full width on mobile, 3 columns on large screens */}
          <div className="lg:col-span-3">
            <PortfolioSummary />
          </div>
          
          {/* Holdings Table */}
          <div className="lg:col-span-2">
            <HoldingsTable />
          </div>
          
          {/* Add Holding Form & Asset Search */}
          <div className="lg:col-span-1 space-y-6">
            <div className="bg-white p-6 rounded-lg shadow">
              <h3 className="text-lg font-semibold text-gray-900 mb-4">Market Search</h3>
              <AssetSearch 
                onAssetSelect={handleAssetSelect}
                showTrending={true}
                enableDetailsModal={true}
                placeholder="Search stocks, ETFs, crypto..."
              />
            </div>
            <AddHoldingForm />
          </div>
          
          {/* Portfolio Chart */}
          <div className="lg:col-span-3">
            <PortfolioChart />
          </div>
          
          {/* Analytics Dashboard */}
          <div className="lg:col-span-3">
            <AnalyticsDashboard />
          </div>
        </div>
        
        {/* What-If Analysis Modal */}
        <WhatIfAnalysisModal 
          isOpen={isWhatIfModalOpen}
          onClose={() => setIsWhatIfModalOpen(false)}
        />
        
        {/* Transaction Form Modal */}
        <TransactionForm
          isOpen={isTransactionFormOpen}
          onClose={() => setIsTransactionFormOpen(false)}
          transaction={null}
        />
        
        {/* Export Modal */}
        <ExportModal
          isOpen={isExportModalOpen}
          onClose={() => setIsExportModalOpen(false)}
        />
      </div>
    </div>
  )
}
