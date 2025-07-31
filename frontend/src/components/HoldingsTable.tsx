import React, { useState, useEffect } from 'react';

import {

usePortfolioHoldings,

useUpdateHolding,

useDeleteHolding,

useCurrentPrice

} from '../hooks/usePortfolio';

import { Holding } from '../types/portfolio';

import { usePortfolioWebSocket } from '../hooks/useWebSocket';

  

// Enhanced price display component with WebSocket integration

const PriceDisplay: React.FC<{ symbol: string; isWebSocketConnected: boolean }> = ({ symbol, isWebSocketConnected }) => {

const { data, isLoading, error } = useCurrentPrice(symbol);

const [isUpdating, setIsUpdating] = useState(false);

  

// Show update animation when price changes

useEffect(() => {

if (data && isWebSocketConnected) {

setIsUpdating(true);

const timer = setTimeout(() => setIsUpdating(false), 1000);

return () => clearTimeout(timer);

}

}, [data?.current_price, data?.change, isWebSocketConnected]);

  

if (isLoading) return <span className="text-gray-400">Loading...</span>;

if (error) return <span className="text-red-500">Error</span>;

if (!data) return <span>-</span>;

  

const isPositive = data.change >= 0;

const arrow = isPositive ? '↑' : '↓';

const colorClass = isPositive ? 'text-green-600' : 'text-red-600';

  

return (

<div className={`flex flex-col transition-all duration-300 ${isUpdating ? 'bg-blue-50 rounded px-2 py-1' : ''}`}>

<div className="flex items-center space-x-2">

<span className="font-medium">${data.current_price.toFixed(2)}</span>

{isWebSocketConnected && (

<div className="flex items-center space-x-1">

<div className="w-1.5 h-1.5 rounded-full bg-green-500 animate-pulse" title="Live data"></div>

<span className="text-xs text-green-600 font-medium">LIVE</span>

</div>

)}

</div>

<div className="flex items-center space-x-1">

<span className={`text-xs font-medium ${colorClass}`}>

{arrow} ${Math.abs(data.change).toFixed(2)}

</span>

<span className={`text-xs ${colorClass}`}>

({isPositive ? '+' : ''}{data.change_percent.toFixed(2)}%)

</span>

</div>

</div>

);

};

  

// Enhanced value display with real-time calculations

const ValueDisplay: React.FC<{

holding: Holding;

symbol: string;

isWebSocketConnected: boolean

}> = ({ holding, symbol, isWebSocketConnected }) => {

const { data: priceData } = useCurrentPrice(symbol);

const [isUpdating, setIsUpdating] = useState(false);

const currentPrice = priceData?.current_price;

const costBasis = holding.quantity * holding.average_cost;

const currentValue = currentPrice ? holding.quantity * currentPrice : costBasis;

const gainLoss = currentValue - costBasis;

const gainLossPercent = costBasis > 0 ? (gainLoss / costBasis) * 100 : 0;

  

useEffect(() => {

if (currentPrice && isWebSocketConnected) {

setIsUpdating(true);

const timer = setTimeout(() => setIsUpdating(false), 1000);

return () => clearTimeout(timer);

}

}, [currentPrice, isWebSocketConnected]);

  

const isPositive = gainLoss >= 0;

const colorClass = isPositive ? 'text-green-600' : 'text-red-600';

  

return (

<div className={`flex flex-col transition-all duration-300 ${isUpdating ? 'bg-blue-50 rounded px-2 py-1' : ''}`}>

<span className="font-medium">${currentValue.toFixed(2)}</span>

{currentPrice && (

<div className="flex items-center space-x-1">

<span className={`text-xs font-medium ${colorClass}`}>

{isPositive ? '+' : ''}${gainLoss.toFixed(2)}

</span>

<span className={`text-xs ${colorClass}`}>

({isPositive ? '+' : ''}{gainLossPercent.toFixed(2)}%)

</span>

</div>

)}

</div>

);

};

  

// 确认对话框组件

const ConfirmModal: React.FC<{

title: string;

message: string;

onConfirm: () => void;

onCancel: () => void;

}> = ({ title, message, onConfirm, onCancel }) => (

<div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">

<div className="bg-white p-6 rounded-lg w-96">

<h2 className="text-xl font-bold mb-2">{title}</h2>

<p className="mb-6">{message}</p>

<div className="flex justify-end space-x-2">

<button className="px-4 py-2 border rounded" onClick={onCancel}>

Cancel

</button>

<button className="px-4 py-2 bg-red-600 text-white rounded" onClick={onConfirm}>

Confirm

</button>

</div>

</div>

</div>

);

  

// 编辑模态框组件

const EditModal: React.FC<{

holding: Holding;

onClose: () => void;

onSave: (updated: Holding) => void;

}> = ({ holding, onClose, onSave }) => {

const [formData, setFormData] = useState({

quantity: holding.quantity,

average_cost: holding.average_cost,

purchase_date: holding.purchase_date,

});

  

const handleSubmit = (e: React.FormEvent) => {

e.preventDefault();

onSave({ ...holding, ...formData });

};

  

return (

<div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center">

<div className="bg-white p-6 rounded-lg w-96">

<h2 className="text-xl font-bold mb-4">Edit Holding</h2>

<form onSubmit={handleSubmit}>

<div className="mb-4">

<label htmlFor="edit-quantity" className="block text-sm font-medium text-gray-700 mb-1">Quantity</label>

<input

id="edit-quantity"

type="number"

className="w-full p-2 border rounded"

value={formData.quantity}

onChange={(e) => setFormData({...formData, quantity: Number(e.target.value)})}

min="1"

step="1"

required

/>

</div>

<div className="mb-4">

<label className="block text-sm font-medium text-gray-700 mb-1">Average Cost</label>

<input

type="number"

className="w-full p-2 border rounded"

value={formData.average_cost}

onChange={(e) => setFormData({...formData, average_cost: Number(e.target.value)})}

min="0.01"

step="0.01"

required

/>

</div>

<div className="flex justify-end space-x-2">

<button type="button" className="px-4 py-2 border rounded" onClick={onClose}>

Cancel

</button>

<button type="submit" className="px-4 py-2 bg-blue-600 text-white rounded">

Save

</button>

</div>

</form>

</div>

</div>

);

};

  

// 主表格组件

const HoldingsTable: React.FC = () => {

const { data, isLoading, error } = usePortfolioHoldings();

const [editingHolding, setEditingHolding] = useState<Holding | null>(null);

const [deletingHolding, setDeletingHolding] = useState<Holding | null>(null);

const updateHolding = useUpdateHolding();

const deleteHolding = useDeleteHolding();

// WebSocket connection for real-time updates

const webSocket = usePortfolioWebSocket(true);

  

if (isLoading) return (

<div className="flex justify-center items-center h-64" role="status" aria-label="Loading holdings data">

<div className="animate-spin rounded-full h-12 w-12 border-t-2 border-b-2 border-blue-500"></div>

<span className="sr-only">Loading...</span>

</div>

);

  

if (error) return (

<div className="bg-red-50 border-l-4 border-red-500 p-4">

<div className="flex">

<div className="flex-shrink-0">

<svg className="h-5 w-5 text-red-500" viewBox="0 0 20 20" fill="currentColor">

<path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />

</svg>

</div>

<div className="ml-3">

<p className="text-sm text-red-700">

Error loading holdings: {error.message}

</p>

</div>

</div>

</div>

);

  

if (!data?.holdings?.length) return <div className="p-4 text-gray-500">No holdings found</div>;

  

return (

<div className="space-y-4">

{/* Connection Status Header */}

<div className="flex items-center justify-between">

<h3 className="text-lg font-semibold text-gray-900">Holdings</h3>

<div className="flex items-center space-x-2">

<div className={`w-2 h-2 rounded-full ${webSocket.isConnected ? 'bg-green-500' : 'bg-gray-400'}`}></div>

<span className="text-sm text-gray-600">

{webSocket.isConnected ? 'Real-time Updates' : 'Static Data'}

</span>

</div>

</div>

  

<div className="overflow-x-auto" data-testid="holdings-table" data-test-version="1.0">

<table className="min-w-full divide-y divide-gray-200">

<thead className="bg-gray-50">

<tr>

<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Symbol</th>

<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Name</th>

<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Quantity</th>

<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Avg Price</th>

<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Current Price</th>

<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Market Value</th>

<th className="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase tracking-wider">Actions</th>

</tr>

</thead>

<tbody className="bg-white divide-y divide-gray-200">

{data.holdings.map((holding) => (

<tr key={holding.id} className="hover:bg-gray-50">

<td className="px-6 py-4 whitespace-nowrap text-sm font-medium text-gray-900" data-testid={`holding-symbol-${holding.id}`}>

{holding.symbol}

</td>

<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500" data-testid={`holding-name-${holding.id}`}>

{holding.name}

</td>

<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500" data-testid={`holding-quantity-${holding.id}`}>

{holding.quantity.toLocaleString()}

</td>

<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500" data-testid={`holding-avgprice-${holding.id}`}>

${holding.average_cost.toFixed(2)}

</td>

<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500" data-testid={`holding-currentprice-${holding.id}`}>

<PriceDisplay symbol={holding.symbol} isWebSocketConnected={webSocket.isConnected} />

</td>

<td className="px-6 py-4 whitespace-nowrap text-sm text-gray-500" data-testid={`holding-value-${holding.id}`}>

<ValueDisplay

holding={holding}

symbol={holding.symbol}

isWebSocketConnected={webSocket.isConnected}

/>

</td>

<td className="px-6 py-4 whitespace-nowrap text-sm font-medium">

<button

className="text-indigo-600 hover:text-indigo-900 mr-3 transition-colors"

data-testid={`edit-btn-${holding.id}`}

onClick={() => setEditingHolding(holding)}

>

Edit

</button>

<button

className="text-red-600 hover:text-red-900 transition-colors"

data-testid={`delete-btn-${holding.id}`}

onClick={() => setDeletingHolding(holding)}

>

Delete

</button>

</td>

</tr>

))}

</tbody>

</table>

</div>

  

{editingHolding && (

<EditModal

holding={editingHolding}

onClose={() => setEditingHolding(null)}

onSave={(updated) => {

updateHolding.mutate(updated);

setEditingHolding(null);

}}

/>

)}

  

{deletingHolding && (

<ConfirmModal

title="Delete Holding"

message={`Are you sure you want to delete ${deletingHolding.name} (${deletingHolding.symbol})?`}

onConfirm={() => {

deleteHolding.mutate(deletingHolding.id);

setDeletingHolding(null);

}}

onCancel={() => setDeletingHolding(null)}

/>

)}

</div>

);

};

  

export default HoldingsTable;
