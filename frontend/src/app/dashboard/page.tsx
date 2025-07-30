'use client'

import PortfolioSummary from '@/components/PortfolioSummary'
import HoldingsTable from '@/components/HoldingsTable'
import AddHoldingForm from '@/components/AddHoldingForm'
import PortfolioChart from '@/components/PortfolioChart'

export default function DashboardPage() {
  return (
    <div className="min-h-screen bg-gray-50 p-6">
      <div className="max-w-7xl mx-auto">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">Portfolio Dashboard</h1>
        
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
          
          {/* Add Holding Form */}
          <div className="lg:col-span-1">
            <AddHoldingForm />
          </div>
          
          {/* Portfolio Chart */}
          <div className="lg:col-span-3">
            <PortfolioChart />
          </div>
        </div>
      </div>
    </div>
  )
}
