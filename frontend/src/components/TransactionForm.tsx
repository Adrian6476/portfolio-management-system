import React, { useState, useEffect } from 'react';
import { useTransactionManagement } from '@/hooks/useTransactions';
import type { Transaction, CreateTransactionRequest } from '@/hooks/useTransactions';
import { Button, LoadingSpinner, ErrorMessage } from './ui';
import AssetSelector from './AssetSelector';
import { AssetSearchResult, TrendingAsset } from '@/hooks/useMarketData';

interface TransactionFormProps {
  isOpen: boolean;
  onClose: () => void;
  transaction?: Transaction | null;
}

const TransactionForm: React.FC<TransactionFormProps> = ({ isOpen, onClose, transaction }) => {
  const [formData, setFormData] = useState({
    symbol: '',
    transaction_type: 'BUY' as 'BUY' | 'SELL',
    quantity: '',
    price: '',
    fees: '',
    notes: '',
  });

  const [selectedAssetInfo, setSelectedAssetInfo] = useState<AssetSearchResult | TrendingAsset | null>(null);
  const { createTransaction, updateTransaction, deleteTransaction, isLoading, error } = useTransactionManagement();

  // Populate form when editing
  useEffect(() => {
    if (transaction) {
      setFormData({
        symbol: transaction.symbol,
        transaction_type: transaction.transaction_type,
        quantity: transaction.quantity.toString(),
        price: transaction.price.toString(),
        fees: transaction.fees?.toString() || '',
        notes: transaction.notes || '',
      });
    } else {
      // Reset form for new transaction
      setFormData({
        symbol: '',
        transaction_type: 'BUY',
        quantity: '',
        price: '',
        fees: '',
        notes: '',
      });
    }
  }, [transaction]);

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement | HTMLTextAreaElement>) => {
    const { name, value } = e.target;
    setFormData(prev => ({
      ...prev,
      [name]: value,
    }));
  };

  const handleAssetSelect = (asset: AssetSearchResult | TrendingAsset) => {
    setSelectedAssetInfo(asset);
    setFormData(prev => ({
      ...prev,
      symbol: asset.symbol,
      // Auto-populate price if available
      price: 'current_price' in asset ? asset.current_price.toString() : prev.price
    }));
  };

  const handleSymbolChange = (symbol: string) => {
    setFormData(prev => ({
      ...prev,
      symbol: symbol
    }));
    // Clear selected asset info if user types manually
    if (selectedAssetInfo && selectedAssetInfo.symbol !== symbol) {
      setSelectedAssetInfo(null);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.symbol || !formData.quantity || !formData.price) {
      return;
    }

    const transactionData: CreateTransactionRequest = {
      symbol: formData.symbol.toUpperCase(),
      transaction_type: formData.transaction_type,
      quantity: parseFloat(formData.quantity),
      price: parseFloat(formData.price),
      fees: formData.fees ? parseFloat(formData.fees) : undefined,
      notes: formData.notes || undefined,
    };

    try {
      if (transaction) {
        // Update existing transaction
        await updateTransaction.mutateAsync({
          id: transaction.id,
          ...transactionData,
        });
      } else {
        // Create new transaction
        await createTransaction.mutateAsync(transactionData);
      }
      onClose();
    } catch (error) {
      // Error is handled by the mutation hooks
      console.error('Transaction operation failed:', error);
    }
  };

  const handleDelete = async () => {
    if (!transaction || !window.confirm('Are you sure you want to delete this transaction?')) {
      return;
    }

    try {
      await deleteTransaction.mutateAsync(transaction.id);
      onClose();
    } catch (error) {
      console.error('Delete failed:', error);
    }
  };

  const resetForm = () => {
    setFormData({
      symbol: '',
      transaction_type: 'BUY',
      quantity: '',
      price: '',
      fees: '',
      notes: '',
    });
    setSelectedAssetInfo(null);
  };

  const handleClose = () => {
    resetForm();
    onClose();
  };

  if (!isOpen) return null;

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50 p-4">
      <div className="bg-white rounded-lg shadow-xl max-w-2xl w-full max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex justify-between items-center p-6 border-b border-gray-200">
          <h2 className="text-2xl font-semibold text-gray-800">
            {transaction ? 'Edit Transaction' : 'Add Transaction'}
          </h2>
          <button
            onClick={handleClose}
            className="text-gray-500 hover:text-gray-700 text-2xl font-bold"
          >
            ×
          </button>
        </div>

        {/* Content */}
        <div className="p-6">
          <form onSubmit={handleSubmit} className="space-y-6">
            {/* Transaction Type and Symbol */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label htmlFor="transaction_type" className="block text-sm font-medium text-gray-700 mb-1">
                  Transaction Type
                </label>
                <select
                  id="transaction_type"
                  name="transaction_type"
                  value={formData.transaction_type}
                  onChange={handleInputChange}
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                >
                  <option value="BUY">Buy</option>
                  <option value="SELL">Sell</option>
                </select>
              </div>

              <div>
                <label htmlFor="symbol" className="block text-sm font-medium text-gray-700 mb-1">
                  Symbol *
                </label>
                <AssetSelector
                  value={formData.symbol}
                  onChange={handleSymbolChange}
                  onAssetSelect={handleAssetSelect}
                  placeholder="Search for stocks, ETFs, crypto..."
                  showTrending={true}
                  enableDetailsModal={true}
                />
                {selectedAssetInfo && (
                  <div className="mt-2 p-2 bg-blue-50 rounded-md">
                    <p className="text-sm text-blue-800">
                      <strong>{selectedAssetInfo.symbol}</strong> - {selectedAssetInfo.name}
                      {selectedAssetInfo.exchange && (
                        <span className="text-blue-600 ml-2">({selectedAssetInfo.exchange})</span>
                      )}
                    </p>
                    {'current_price' in selectedAssetInfo && (
                      <p className="text-xs text-blue-600 mt-1">
                        Current Price: ${selectedAssetInfo.current_price?.toFixed(2)}
                      </p>
                    )}
                  </div>
                )}
              </div>
            </div>

            {/* Quantity and Price */}
            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              <div>
                <label htmlFor="quantity" className="block text-sm font-medium text-gray-700 mb-1">
                  Quantity *
                </label>
                <input
                  type="number"
                  id="quantity"
                  name="quantity"
                  value={formData.quantity}
                  onChange={handleInputChange}
                  placeholder="Number of shares"
                  min="0.01"
                  step="0.01"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  required
                />
              </div>

              <div>
                <label htmlFor="price" className="block text-sm font-medium text-gray-700 mb-1">
                  Price per Share ($) *
                </label>
                <input
                  type="number"
                  id="price"
                  name="price"
                  value={formData.price}
                  onChange={handleInputChange}
                  placeholder="Price per share"
                  min="0.01"
                  step="0.01"
                  className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
                  required
                />
              </div>
            </div>

            {/* Fees */}
            <div>
              <label htmlFor="fees" className="block text-sm font-medium text-gray-700 mb-1">
                Fees ($)
              </label>
              <input
                type="number"
                id="fees"
                name="fees"
                value={formData.fees}
                onChange={handleInputChange}
                placeholder="Transaction fees (optional)"
                min="0"
                step="0.01"
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>

            {/* Notes */}
            <div>
              <label htmlFor="notes" className="block text-sm font-medium text-gray-700 mb-1">
                Notes
              </label>
              <textarea
                id="notes"
                name="notes"
                value={formData.notes}
                onChange={handleInputChange}
                placeholder="Optional notes about this transaction"
                rows={3}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              />
            </div>

            {/* Transaction Summary */}
            {formData.quantity && formData.price && (
              <div className="bg-gray-50 p-4 rounded-lg border border-gray-200">
                <h4 className="text-sm font-medium text-gray-900 mb-2">Transaction Summary</h4>
                <div className="grid grid-cols-2 gap-4 text-sm">
                  <div>
                    <span className="text-gray-600">Total Value:</span>
                    <div className="font-semibold text-gray-900">
                      ${(parseFloat(formData.quantity) * parseFloat(formData.price)).toFixed(2)}
                    </div>
                  </div>
                  {formData.fees && (
                    <div>
                      <span className="text-gray-600">Total with Fees:</span>
                      <div className="font-semibold text-gray-900">
                        ${(parseFloat(formData.quantity) * parseFloat(formData.price) + parseFloat(formData.fees)).toFixed(2)}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}

            {/* Error Display */}
            {error && (
              <ErrorMessage message={error.message || 'Failed to save transaction'} />
            )}

            {/* Form Actions */}
            <div className="flex justify-between">
              <div>
                {transaction && (
                  <Button
                    type="button"
                    variant="danger"
                    onClick={handleDelete}
                    disabled={isLoading}
                  >
                    Delete Transaction
                  </Button>
                )}
              </div>
              
              <div className="flex space-x-3">
                <Button
                  type="button"
                  variant="secondary"
                  onClick={handleClose}
                  disabled={isLoading}
                >
                  Cancel
                </Button>
                
                <Button
                  type="submit"
                  disabled={isLoading || !formData.symbol || !formData.quantity || !formData.price}
                  className="flex items-center"
                >
                  {isLoading && <LoadingSpinner size="sm" className="mr-2" />}
                  {isLoading 
                    ? 'Saving...' 
                    : transaction 
                    ? 'Update Transaction' 
                    : 'Create Transaction'
                  }
                </Button>
              </div>
            </div>
          </form>
        </div>
      </div>
    </div>
  );
};

export default TransactionForm;
