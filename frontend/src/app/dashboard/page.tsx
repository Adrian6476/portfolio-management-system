'use client'

import PortfolioSummary from '@/components/PortfolioSummary'

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
          
          {/* Placeholders for other components that teammates will build */}
          <div className="lg:col-span-2">
            <div className="bg-white rounded-lg shadow p-6 h-96">
              <h2 className="text-xl font-semibold text-gray-800 mb-4">Holdings Table</h2>
              <p className="text-gray-500">Component will be implemented by Developer B</p>
            </div>
          </div>
          
          <div className="lg:col-span-1">
            <div className="bg-white rounded-lg shadow p-6 h-96">
              <h2 className="text-xl font-semibold text-gray-800 mb-4">Add Holding Form</h2>
              <p className="text-gray-500">Component will be implemented by Developer C</p>
            </div>
          </div>
          
          <div className="lg:col-span-3">
            <div className="bg-white rounded-lg shadow p-6 h-96">
              <h2 className="text-xl font-semibold text-gray-800 mb-4">Portfolio Chart</h2>
              <p className="text-gray-500">Component will be implemented by Developer C</p>
            </div>
          </div>
        </div>
      </div>
    </div>
  )
}
