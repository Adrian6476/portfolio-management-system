'use client'

  

import { useQuery } from 'react-query'

import apiClient from '@/lib/api'

import { PortfolioSummaryResponse, QUERY_KEYS } from '@/types/portfolio'

import { usePortfolioWebSocket } from '@/hooks/useWebSocket'

import { useState, useEffect } from 'react'

  

// Loading spinner component

function LoadingSpinner() {

return (

<div className="flex items-center justify-center p-8">

<div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-600"></div>

<span className="ml-2 text-gray-600">Loading portfolio summary...</span>

</div>

)

}

  

// Error alert component

function ErrorAlert({ message }: { message: string }) {

return (

<div className="bg-red-50 border border-red-200 rounded-lg p-4">

<div className="flex">

<div className="flex-shrink-0">

<svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">

<path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />

</svg>

</div>

<div className="ml-3">

<h3 className="text-sm font-medium text-red-800">Error loading portfolio summary</h3>

<p className="mt-1 text-sm text-red-700">{message}</p>

</div>

</div>

</div>

)

}

  

// WebSocket connection status indicator

function ConnectionStatus({ isConnected, connectionStatus }: { isConnected: boolean; connectionStatus: string }) {

const getStatusColor = () => {

switch (connectionStatus) {

case 'connected': return 'bg-green-500'

case 'connecting': return 'bg-yellow-500'

case 'error': return 'bg-red-500'

default: return 'bg-gray-500'

}

}

  

return (

<div className="flex items-center space-x-2 text-sm">

<div className={`w-2 h-2 rounded-full ${getStatusColor()}`}></div>

<span className="text-gray-600">

{connectionStatus === 'connected' ? 'Live Updates' :

connectionStatus === 'connecting' ? 'Connecting...' :

connectionStatus === 'error' ? 'Connection Error' : 'Disconnected'}

</span>

</div>

)

}

  

// Enhanced metric card component with real-time updates

function MetricCard({

title,

value,

subtitle,

change,

changePercent,

isRealTime = false

}: {

title: string;

value: number;

subtitle?: string;

change?: number;

changePercent?: number;

isRealTime?: boolean;

}) {

const [isUpdating, setIsUpdating] = useState(false)

  

useEffect(() => {

if (isRealTime) {

setIsUpdating(true)

const timer = setTimeout(() => setIsUpdating(false), 500)

return () => clearTimeout(timer)

}

}, [value, isRealTime])

  

const formatValue = (val: number) => {

if (title.toLowerCase().includes('cost') || title.toLowerCase().includes('value')) {

return new Intl.NumberFormat('en-US', {

style: 'currency',

currency: 'USD',

}).format(val)

}

return val.toLocaleString()

}

  

const formatChange = (val: number) => {

const formatted = new Intl.NumberFormat('en-US', {

style: 'currency',

currency: 'USD',

signDisplay: 'always'

}).format(val)

return formatted

}

  

const getChangeColor = (val: number) => {

if (val > 0) return 'text-green-600'

if (val < 0) return 'text-red-600'

return 'text-gray-600'

}

  

return (

<div className={`bg-white rounded-lg shadow p-6 transition-all duration-300 ${isUpdating ? 'ring-2 ring-blue-200 bg-blue-50' : ''}`}>

<div className="flex items-center">

<div className="flex-1">

<div className="flex items-center justify-between">

<p className="text-sm font-medium text-gray-600">{title}</p>

{isRealTime && (

<div className="flex items-center space-x-1">

<div className="w-1.5 h-1.5 bg-green-500 rounded-full animate-pulse"></div>

<span className="text-xs text-green-600">Live</span>

</div>

)}

</div>

<p className="text-2xl font-bold text-gray-900">{formatValue(value)}</p>

{change !== undefined && changePercent !== undefined && (

<div className={`text-sm font-medium ${getChangeColor(change)}`}>

{formatChange(change)} ({changePercent >= 0 ? '+' : ''}{changePercent.toFixed(2)}%)

</div>

)}

{subtitle && <p className="text-sm text-gray-500">{subtitle}</p>}

</div>

</div>

</div>

)

}

  

// Asset allocation card component

function AssetAllocationCard({ allocations }: { allocations: any[] }) {

return (

<div className="bg-white rounded-lg shadow p-6">

<h3 className="text-lg font-semibold text-gray-900 mb-4">Asset Allocation</h3>

<div className="space-y-3">

{allocations.map((allocation, index) => (

<div key={index} className="flex items-center justify-between">

<div className="flex items-center">

<div className="w-3 h-3 rounded-full bg-blue-500 mr-3"></div>

<span className="text-sm font-medium text-gray-700">{allocation.asset_type}</span>

</div>

<div className="text-right">

<div className="text-sm font-semibold text-gray-900">

{new Intl.NumberFormat('en-US', {

style: 'currency',

currency: 'USD',

}).format(allocation.total_value)}

</div>

<div className="text-xs text-gray-500">{allocation.percentage.toFixed(1)}%</div>

</div>

</div>

))}

</div>

</div>

)

}

  

// Top holdings card component

function TopHoldingsCard({ holdings }: { holdings: any[] }) {

return (

<div className="bg-white rounded-lg shadow p-6">

<h3 className="text-lg font-semibold text-gray-900 mb-4">Top Holdings</h3>

<div className="space-y-3">

{holdings.slice(0, 5).map((holding, index) => (

<div key={index} className="flex items-center justify-between">

<div>

<div className="text-sm font-medium text-gray-900">{holding.symbol}</div>

<div className="text-xs text-gray-500">{holding.name}</div>

</div>

<div className="text-right">

<div className="text-sm font-semibold text-gray-900">

{new Intl.NumberFormat('en-US', {

style: 'currency',

currency: 'USD',

}).format(holding.total_value || (holding.quantity * holding.average_cost))}

</div>

<div className="text-xs text-gray-500">{holding.quantity} shares</div>

</div>

</div>

))}

</div>

</div>

)

}

  

export default function PortfolioSummary() {

// WebSocket connection for real-time updates

const webSocket = usePortfolioWebSocket(true)

const { data, isLoading, isError, error } = useQuery<PortfolioSummaryResponse>(

QUERY_KEYS.PORTFOLIO_SUMMARY,

async () => {

const response = await apiClient.get('/portfolio/summary')

return response.data

},

{

retry: 2,

refetchOnWindowFocus: false,

refetchInterval: webSocket.isConnected ? false : 30000, // Only poll if WebSocket is disconnected

}

)

  

if (isLoading) {

return (

<div className="bg-white rounded-lg shadow">

<LoadingSpinner />

</div>

)

}

  

if (isError) {

const errorMessage = error instanceof Error ? error.message : 'Failed to fetch portfolio summary'

return <ErrorAlert message={errorMessage} />

}

  

if (!data) {

return <ErrorAlert message="No portfolio data available" />

}

  

return (

<div className="space-y-6">

{/* Connection Status Header */}

<div className="flex items-center justify-between">

<h2 className="text-xl font-semibold text-gray-900">Portfolio Summary</h2>

<ConnectionStatus

isConnected={webSocket.isConnected}

connectionStatus={webSocket.connectionStatus}

/>

</div>

  

{/* Enhanced Summary metrics with real-time data */}

<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">

<MetricCard

title="Total Value"

value={data.summary.total_market_value || 0}

change={data.summary.daily_change}

changePercent={data.summary.daily_change_percent}

isRealTime={webSocket.isConnected}

subtitle="Current market value"

/>

<MetricCard

title="Total Cost"

value={data.summary.total_cost}

subtitle="Total invested amount"

/>

<MetricCard

title="Unrealized P&L"

value={data.summary.unrealized_gain_loss || 0}

changePercent={data.summary.unrealized_gain_loss_percent}

isRealTime={webSocket.isConnected}

subtitle="Unrealized gains/losses"

/>

<MetricCard

title="Total Holdings"

value={data.summary.total_holdings}

subtitle="Number of positions"

/>

</div>

  

{/* Asset allocation and top holdings */}

<div className="grid grid-cols-1 lg:grid-cols-2 gap-6">

{data.asset_allocation && data.asset_allocation.length > 0 && (

<AssetAllocationCard allocations={data.asset_allocation} />

)}

{data.top_holdings && data.top_holdings.length > 0 && (

<TopHoldingsCard holdings={data.top_holdings} />

)}

</div>

</div>

)

}