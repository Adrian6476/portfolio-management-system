import React, { useState } from 'react';

import { useAssetAllocation } from '../hooks/useAnalytics';

import { Card, LoadingSpinner, ErrorMessage } from './ui';

import { UI_CONSTANTS } from './ui';

import { PieChart, Pie, Cell, ResponsiveContainer, Tooltip, Legend } from 'recharts';

  

const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042', '#8884D8', '#82ca9d', '#ffc658', '#ff7c7c', '#8dd1e1'];

  

type ViewMode = 'asset_type' | 'sector' | 'holdings';

  

export default function PortfolioChart() {

const [viewMode, setViewMode] = useState<ViewMode>('asset_type');

const { data, isLoading, error } = useAssetAllocation();

  

if (isLoading) {

return (

<Card title="Portfolio Allocation" className={UI_CONSTANTS.spacing.section}>

<div className="flex justify-center py-8">

<LoadingSpinner size="lg" />

</div>

</Card>

);

}

  

if (error) {

return (

<Card title="Portfolio Allocation" className={UI_CONSTANTS.spacing.section}>

<ErrorMessage message="Failed to load allocation data" />

</Card>

);

}

  

// Prepare chart data based on view mode

const getChartData = () => {

if (!data) return [];

switch (viewMode) {

case 'asset_type':

return (data.by_asset_type || []).map((item) => ({

name: item.asset_type || 'Unknown',

value: item.value || 0, // API uses 'value' not 'total_value'

percentage: item.percentage || 0,

count: item.count || 0,

}));

case 'sector':

return (data.by_sector || []).map((item) => ({

name: item.sector || 'Unknown',

value: item.value || 0, // API uses 'value' not 'total_value'

percentage: item.percentage || 0,

count: item.count || 0,

}));

case 'holdings':

// Show top 8 holdings, group the rest as "Others"

const topHoldings = (data.top_holdings || []).slice(0, 8);

const others = (data.top_holdings || []).slice(8);

const chartData = topHoldings.map((holding) => ({

name: holding.symbol || 'Unknown',

value: holding.total_value || 0, // Holdings use 'total_value'

percentage: holding.percentage || 0,

count: 1,

}));

if (others.length > 0) {

const othersValue = others.reduce((sum, holding) => sum + (holding.total_value || 0), 0);

const othersPercentage = others.reduce((sum, holding) => sum + (holding.percentage || 0), 0);

chartData.push({

name: `Others (${others.length})`,

value: othersValue,

percentage: othersPercentage,

count: others.length,

});

}

return chartData;

default:

return [];

}

};

  

const chartData = getChartData();

  

const renderCustomTooltip = ({ active, payload }: any) => {

if (active && payload && payload.length) {

const data = payload[0].payload;

return (

<div className="bg-white p-3 border border-gray-200 rounded-lg shadow-lg">

<p className="font-semibold text-gray-900">{data.name}</p>

<p className="text-blue-600">

Value: <span className="font-medium">${data.value.toLocaleString()}</span>

</p>

<p className="text-green-600">

Percentage: <span className="font-medium">{data.percentage.toFixed(1)}%</span>

</p>

{viewMode !== 'holdings' && (

<p className="text-gray-600">

Holdings: <span className="font-medium">{data.count}</span>

</p>

)}

</div>

);

}

return null;

};

  

const getViewModeTitle = () => {

switch (viewMode) {

case 'asset_type':

return 'By Asset Type';

case 'sector':

return 'By Sector';

case 'holdings':

return 'By Holdings';

default:

return '';

}

};

  

return (

<Card className={UI_CONSTANTS.spacing.section}>

{/* Custom Title with View Mode Selector */}

<div className="flex justify-between items-center mb-6">

<h3 className="text-xl font-medium text-gray-700">

Portfolio Allocation {getViewModeTitle()}

</h3>

<div className="flex space-x-1">

<button

onClick={() => setViewMode('asset_type')}

className={`px-3 py-1 text-sm rounded-md transition-colors ${

viewMode === 'asset_type'

? 'bg-blue-100 text-blue-700 font-medium'

: 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'

}`}

>

Asset Type

</button>

<button

onClick={() => setViewMode('sector')}

className={`px-3 py-1 text-sm rounded-md transition-colors ${

viewMode === 'sector'

? 'bg-blue-100 text-blue-700 font-medium'

: 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'

}`}

>

Sector

</button>

<button

onClick={() => setViewMode('holdings')}

className={`px-3 py-1 text-sm rounded-md transition-colors ${

viewMode === 'holdings'

? 'bg-blue-100 text-blue-700 font-medium'

: 'text-gray-600 hover:text-gray-900 hover:bg-gray-100'

}`}

>

Holdings

</button>

</div>

</div>

{/* Portfolio Summary */}

{data && data.allocation_summary && (

<div className="mb-4 p-3 bg-gray-50 rounded-lg">

<div className="grid grid-cols-2 md:grid-cols-3 gap-4 text-sm">

<div>

<span className="text-gray-600">Total Value:</span>

<div className="font-semibold text-lg text-gray-900">

${(data.allocation_summary.total_portfolio_value || 0).toLocaleString()}

</div>

</div>

<div>

<span className="text-gray-600">Total Holdings:</span>

<div className="font-semibold text-lg text-gray-900">

{data.allocation_summary.total_holdings || 0}

</div>

</div>

<div>

<span className="text-gray-600">Last Updated:</span>

<div className="font-semibold text-sm text-gray-900">

{data.allocation_summary.allocation_date === 'current'

? 'Real-time'

: data.allocation_summary.allocation_date

? new Date(data.allocation_summary.allocation_date).toLocaleDateString()

: 'N/A'

}

</div>

</div>

</div>

</div>

)}

  

<div className="h-[400px]">

<ResponsiveContainer width="100%" height="100%">

<PieChart>

<Pie

data={chartData}

cx="50%"

cy="50%"

labelLine={false}

outerRadius={120}

fill="#8884d8"

dataKey="value"

label={({ name, percentage }) => `${name}: ${percentage.toFixed(1)}%`}

>

{chartData.map((entry: any, index: number) => (

<Cell key={`cell-${index}`} fill={COLORS[index % COLORS.length]} />

))}

</Pie>

<Tooltip content={renderCustomTooltip} />

<Legend

wrapperStyle={{ paddingTop: '20px' }}

formatter={(value, entry: any) => (

<span style={{ color: entry.color }}>

{value} ({entry.payload.percentage.toFixed(1)}%)

</span>

)}

/>

</PieChart>

</ResponsiveContainer>

</div>

  

{/* Additional Details */}

{data && (

<div className="mt-4 space-y-3">

{viewMode === 'asset_type' && data.by_asset_type && (

<div>

<h4 className="text-sm font-medium text-gray-900 mb-2">Asset Type Breakdown</h4>

<div className="space-y-1">

{data.by_asset_type.map((asset, index) => (

<div key={asset.asset_type || index} className="flex justify-between items-center text-sm">

<div className="flex items-center">

<div

className="w-3 h-3 rounded-full mr-2"

style={{ backgroundColor: COLORS[index % COLORS.length] }}

/>

<span>{asset.asset_type || 'Unknown'}</span>

</div>

<div className="text-right">

<div className="font-medium">${(asset.value || 0).toLocaleString()}</div>

<div className="text-gray-500">{asset.count || 0} holdings</div>

</div>

</div>

))}

</div>

</div>

)}

  

{viewMode === 'sector' && data.by_sector && (

<div>

<h4 className="text-sm font-medium text-gray-900 mb-2">Sector Breakdown</h4>

<div className="space-y-1">

{data.by_sector.map((sector, index) => (

<div key={sector.sector || index} className="flex justify-between items-center text-sm">

<div className="flex items-center">

<div

className="w-3 h-3 rounded-full mr-2"

style={{ backgroundColor: COLORS[index % COLORS.length] }}

/>

<span>{sector.sector || 'Unknown'}</span>

</div>

<div className="text-right">

<div className="font-medium">${(sector.value || 0).toLocaleString()}</div>

<div className="text-gray-500">{sector.count || 0} holdings</div>

</div>

</div>

))}

</div>

</div>

)}

  

{viewMode === 'holdings' && data.top_holdings && (

<div>

<h4 className="text-sm font-medium text-gray-900 mb-2">Top Holdings</h4>

<div className="space-y-1">

{data.top_holdings.slice(0, 8).map((holding, index) => (

<div key={holding.symbol || index} className="flex justify-between items-center text-sm">

<div className="flex items-center">

<div

className="w-3 h-3 rounded-full mr-2"

style={{ backgroundColor: COLORS[index % COLORS.length] }}

/>

<div>

<div className="font-medium">{holding.symbol || 'Unknown'}</div>

<div className="text-gray-500 text-xs">{holding.name || 'N/A'}</div>

</div>

</div>

<div className="text-right">

<div className="font-medium">${(holding.total_value || 0).toLocaleString()}</div>

<div className="text-gray-500">{holding.quantity || 0} shares</div>

</div>

</div>

))}

</div>

</div>

)}

</div>

)}

</Card>

);

}