'use client'

import { useEffect, useState, ReactNode } from 'react'
import { useSearchParams } from 'next/navigation'
import DashboardLayout from '../components/DashboardLayout'
import PortfolioSummary from '../components/PortfolioSummary'
import AddHoldingForm from '../components/AddHoldingForm'
import PortfolioChart from '../components/PortfolioChart'
import AnalyticsDashboard from '../components/AnalyticsDashboard'
import WhatIfAnalysisModal from '../components/WhatIfAnalysisModal'
import TransactionForm from '../components/TransactionForm'
import Link from 'next/link'

import dynamic from 'next/dynamic';

const HoldingsTable = dynamic(
  () => import('../components/HoldingsTable'),
  {
    loading: () => (
      <div className="bank-card min-h-[400px] flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-t-2 border-b-2 border-blue-500"></div>
      </div>
    ),
    ssr: false
  }
);

export default function HomePage() {
  const searchParams = useSearchParams()
  const [mounted, setMounted] = useState(false)
  const [activeView, setActiveView] = useState('overview')
  const [isWhatIfModalOpen, setIsWhatIfModalOpen] = useState(false)
  const [isTransactionFormOpen, setIsTransactionFormOpen] = useState(false)
  const [preSelectedAsset, setPreSelectedAsset] = useState<{symbol: string, name: string} | null>(null)

  useEffect(() => {
    setMounted(true)
    
    // Check URL parameters for pre-selected asset
    const view = searchParams.get('view')
    const symbol = searchParams.get('symbol')
    const name = searchParams.get('name')
    
    if (view === 'add' && symbol && name) {
      setActiveView('add')
      setPreSelectedAsset({ symbol, name: decodeURIComponent(name) })
    }
  }, [searchParams])

  if (!mounted) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gray-50">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary"></div>
      </div>
    )
  }

  const renderContent = (): ReactNode => {
    switch (activeView) {
      case 'overview':
        return (
          <div className="space-y-6">
            <PortfolioSummary />
            <HoldingsTable />
          </div>
        )
      
      case 'manage':
        return <HoldingsTable />
      
      case 'add':
        return (
          <div className="space-y-6">
            <AddHoldingForm preSelectedAsset={preSelectedAsset} />
            <PortfolioSummary />
          </div>
        )
      
      case 'analytics':
        return (
          <div className="space-y-6">
            <AnalyticsDashboard />
            <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
              <PortfolioChart />
              <PortfolioSummary />
            </div>
          </div>
        )
      
      default:
        return renderContent()
    }
  }

  return (
    <DashboardLayout activeView={activeView} setActiveView={setActiveView}>
      <div className="space-y-6">
        {/* Navigation and Actions Header */}
        <div className="bank-card">
          <div className="flex flex-col sm:flex-row justify-between items-start sm:items-center gap-4">
            <div>
              <h2 className="text-xl font-semibold text-gray-900">Portfolio Management</h2>
              <p className="text-sm text-gray-600 mt-1">Manage your portfolio and analyze potential trades</p>
            </div>
            
            <div className="flex flex-col sm:flex-row gap-3">
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
              
              {/* Market Data Link */}
              <Link 
                href="/market-data"
                className="bg-purple-600 hover:bg-purple-700 text-white px-4 py-2 rounded-md font-medium transition-colors flex items-center"
              >
                <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 7h8m0 0v8m0-8l-8 8-4-4-6 6" />
                </svg>
                Market Data
              </Link>
              
              {/* Link to Alternative Dashboard */}
              <Link 
                href="/dashboard"
                className="bg-gray-600 hover:bg-gray-700 text-white px-4 py-2 rounded-md font-medium transition-colors flex items-center"
              >
                <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2V6zM14 6a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2V6zM4 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2H6a2 2 0 01-2-2v-2zM14 16a2 2 0 012-2h2a2 2 0 012 2v2a2 2 0 01-2 2h-2a2 2 0 01-2-2v-2z" />
                </svg>
                Grid Dashboard
              </Link>
            </div>
          </div>
        </div>

        {renderContent()}
        
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
      
    </DashboardLayout>
  )
}
