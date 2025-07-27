'use client'

import { useEffect, useState } from 'react'
import DashboardLayout from '../components/DashboardLayout'

// Placeholder components for developers
function PortfolioSummaryPlaceholder() {
  return (
    <div className="bank-card min-h-[200px] flex items-center justify-center">
      <div className="text-center">
        <h3 className="text-lg font-semibold text-gray-900 mb-2">Portfolio Summary</h3>
        <p className="text-gray-500">Developer A will implement this component</p>
        <div className="text-xs text-gray-400 mt-2">API: GET /api/v1/portfolio/summary</div>
      </div>
    </div>
  )
}

function HoldingsTablePlaceholder() {
  return (
    <div className="bank-card min-h-[400px] flex items-center justify-center">
      <div className="text-center">
        <h3 className="text-lg font-semibold text-gray-900 mb-2">Holdings Table</h3>
        <p className="text-gray-500">Developer B will implement this component</p>
        <div className="text-xs text-gray-400 mt-2">
          API: GET /api/v1/portfolio, PUT/DELETE /api/v1/portfolio/holdings/:id
        </div>
      </div>
    </div>
  )
}

function AddHoldingFormPlaceholder() {
  return (
    <div className="bank-card min-h-[300px] flex items-center justify-center">
      <div className="text-center">
        <h3 className="text-lg font-semibold text-gray-900 mb-2">Add Holding Form</h3>
        <p className="text-gray-500">Developer C will implement this component</p>
        <div className="text-xs text-gray-400 mt-2">API: POST /api/v1/portfolio/holdings</div>
      </div>
    </div>
  )
}

function PortfolioChartPlaceholder() {
  return (
    <div className="bank-card min-h-[300px] flex items-center justify-center">
      <div className="text-center">
        <h3 className="text-lg font-semibold text-gray-900 mb-2">Portfolio Chart</h3>
        <p className="text-gray-500">Developer C will implement this component</p>
        <div className="text-xs text-gray-400 mt-2">Uses Recharts for asset allocation visualization</div>
      </div>
    </div>
  )
}

export default function HomePage() {
  const [mounted, setMounted] = useState(false)
  const [activeView, setActiveView] = useState('overview')

  useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary"></div>
      </div>
    )
  }

  const renderContent = () => {
    switch (activeView) {
      case 'overview':
        return (
          <div className="space-y-6">
            {/* Portfolio Summary - Full Width */}
            <PortfolioSummaryPlaceholder />
            
            {/* Holdings Table - Full Width */}
            <HoldingsTablePlaceholder />
          </div>
        )
      
      case 'manage':
        return (
          <div className="space-y-6">
            {/* Holdings Management Focus */}
            <HoldingsTablePlaceholder />
          </div>
        )
      
      case 'add':
        return (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Add Form */}
            <AddHoldingFormPlaceholder />
            {/* Portfolio Summary for context */}
            <PortfolioSummaryPlaceholder />
          </div>
        )
      
      case 'analytics':
        return (
          <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
            {/* Chart gets more space */}
            <PortfolioChartPlaceholder />
            {/* Summary for context */}
            <PortfolioSummaryPlaceholder />
          </div>
        )
      
      default:
        return renderContent()
    }
  }

  return (
    <DashboardLayout activeView={activeView} setActiveView={setActiveView}>
      <div className="space-y-6">
        {renderContent()}
        
        {/* Development Status - Only show in overview */}
        {activeView === 'overview' && (
          <div className="bank-card">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">Development Status</h3>
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="p-4 bg-blue-50 rounded-lg border border-blue-200">
                <h4 className="font-medium text-blue-900 mb-2">Developer A Tasks</h4>
                <ul className="text-sm text-blue-700 space-y-1">
                  <li>✅ TypeScript interfaces defined</li>
                  <li>✅ API client configured</li>
                  <li>🔄 Portfolio Summary component</li>
                  <li>🔄 Dashboard layout integration</li>
                </ul>
              </div>
              <div className="p-4 bg-green-50 rounded-lg border border-green-200">
                <h4 className="font-medium text-green-900 mb-2">Developer B Tasks</h4>
                <ul className="text-sm text-green-700 space-y-1">
                  <li>⏳ Holdings data fetching hooks</li>
                  <li>⏳ Interactive holdings table</li>
                  <li>⏳ CRUD operations</li>
                  <li>⏳ Real-time price integration</li>
                </ul>
              </div>
              <div className="p-4 bg-purple-50 rounded-lg border border-purple-200">
                <h4 className="font-medium text-purple-900 mb-2">Developer C Tasks</h4>
                <ul className="text-sm text-purple-700 space-y-1">
                  <li>⏳ Add holding form with validation</li>
                  <li>⏳ Portfolio allocation chart</li>
                  <li>⏳ Form integration with API</li>
                  <li>⏳ Chart responsiveness</li>
                </ul>
              </div>
            </div>
          </div>
        )}
      </div>
    </DashboardLayout>
  )
}