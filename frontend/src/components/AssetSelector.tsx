import React, { useState, useRef, useEffect } from 'react';
import { useAssetSearch, useTrendingAssets, AssetSearchResult, TrendingAsset } from '@/hooks/useMarketData';
import { LoadingSpinner } from './ui';
import AssetDetailsModal from './AssetDetailsModal';

type SelectableAsset = AssetSearchResult | TrendingAsset;

interface AssetSelectorProps {
  value: string;
  onChange: (symbol: string) => void;
  onAssetSelect?: (asset: SelectableAsset) => void;
  placeholder?: string;
  className?: string;
  error?: string;
  showTrending?: boolean;
  enableDetailsModal?: boolean;
  disabled?: boolean;
  id?: string;
}

const AssetSelector: React.FC<AssetSelectorProps> = ({
  value,
  onChange,
  onAssetSelect,
  placeholder = "Search for stocks, ETFs, crypto...",
  className = "",
  error,
  showTrending = true,
  enableDetailsModal = true,
  disabled = false,
  id
}) => {
  const [query, setQuery] = useState(value);
  const [isOpen, setIsOpen] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(-1);
  const [selectedAsset, setSelectedAsset] = useState<string | null>(null);
  const [showAssetModal, setShowAssetModal] = useState(false);
  const searchRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const { data: searchResults = [], isLoading } = useAssetSearch(query, query.length >= 2);
  const { data: trendingAssets = [] } = useTrendingAssets();

  const displayResults = query.length >= 2 ? searchResults : (showTrending ? trendingAssets.slice(0, 8) : []);

  // Sync internal query with external value
  useEffect(() => {
    setQuery(value);
  }, [value]);

  // Close dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (searchRef.current && !searchRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, []);

  // Handle keyboard navigation
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!isOpen) return;

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setSelectedIndex(prev => 
          prev < displayResults.length - 1 ? prev + 1 : prev
        );
        break;
      case 'ArrowUp':
        e.preventDefault();
        setSelectedIndex(prev => prev > 0 ? prev - 1 : -1);
        break;
      case 'Enter':
        e.preventDefault();
        if (selectedIndex >= 0 && displayResults[selectedIndex]) {
          const asset = displayResults[selectedIndex];
          handleAssetSelect(asset);
        }
        break;
      case 'Escape':
        setIsOpen(false);
        setSelectedIndex(-1);
        inputRef.current?.blur();
        break;
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const newValue = e.target.value;
    setQuery(newValue);
    onChange(newValue);
    setIsOpen(true);
    setSelectedIndex(-1);
  };

  const handleAssetSelect = (asset: SelectableAsset) => {
    setQuery(asset.symbol);
    onChange(asset.symbol);
    onAssetSelect?.(asset);
    setIsOpen(false);
    setSelectedIndex(-1);
  };

  const handleViewDetails = (symbol: string) => {
    setSelectedAsset(symbol);
    setShowAssetModal(true);
    setIsOpen(false);
  };

  const handleAssetFromModal = (symbol: string, name: string) => {
    const asset = displayResults.find(a => a.symbol === symbol);
    if (asset) {
      handleAssetSelect(asset);
    }
    setShowAssetModal(false);
    setSelectedAsset(null);
  };

  const formatPrice = (price: number | undefined) => {
    if (price === undefined) return 'N/A';
    return new Intl.NumberFormat('en-US', {
      style: 'currency',
      currency: 'USD',
      minimumFractionDigits: 2,
      maximumFractionDigits: 2,
    }).format(price);
  };

  const formatChange = (change: number | undefined, changePercent: number | undefined) => {
    if (change === undefined || changePercent === undefined) return '';
    const sign = change >= 0 ? '+' : '';
    return `${sign}${change.toFixed(2)} (${sign}${changePercent.toFixed(2)}%)`;
  };

  const formatChangeForAsset = (asset: SelectableAsset) => {
    if ('change' in asset && 'change_percent' in asset) {
      // AssetSearchResult or MarketMover with both change and change_percent
      return formatChange(asset.change as number, asset.change_percent as number);
    } else if ('change_percent' in asset) {
      // TrendingAsset - only has change_percent
      const changePercent = asset.change_percent as number;
      if (changePercent === undefined) return '';
      const sign = changePercent >= 0 ? '+' : '';
      return `${sign}${changePercent.toFixed(2)}%`;
    }
    return '';
  };

  return (
    <div ref={searchRef} className={`relative ${className}`}>
      <input
        ref={inputRef}
        id={id}
        type="text"
        value={query}
        onChange={handleInputChange}
        onFocus={() => setIsOpen(true)}
        onKeyDown={handleKeyDown}
        placeholder={placeholder}
        disabled={disabled}
        className={`w-full p-2 border rounded ${error ? 'border-red-500' : 'border-gray-300'} ${
          disabled ? 'bg-gray-100 cursor-not-allowed' : 'bg-white'
        }`}
      />
      
      {error && (
        <p className="text-red-500 text-sm mt-1">{error}</p>
      )}

      {isOpen && !disabled && (
        <div className="absolute z-50 w-full mt-1 bg-white border border-gray-300 rounded-md shadow-lg max-h-96 overflow-y-auto">
          {isLoading && query.length >= 2 ? (
            <div className="flex items-center justify-center p-4">
              <LoadingSpinner size="sm" className="mr-2" />
              <span className="text-gray-600">Searching...</span>
            </div>
          ) : displayResults.length > 0 ? (
            <>
              {query.length < 2 && showTrending && (
                <div className="px-3 py-2 bg-gray-50 text-sm text-gray-600 font-medium border-b">
                  Trending Assets
                </div>
              )}
              {displayResults.map((asset, index) => (
                <div
                  key={asset.symbol}
                  className={`px-3 py-3 cursor-pointer border-b border-gray-100 last:border-b-0 hover:bg-gray-50 ${
                    index === selectedIndex ? 'bg-blue-50' : ''
                  }`}
                  onClick={() => handleAssetSelect(asset)}
                >
                  <div className="flex items-center justify-between">
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center space-x-2">
                        <span className="font-medium text-gray-900">
                          {asset.symbol}
                        </span>
                        <span className="text-xs bg-gray-100 text-gray-600 px-1.5 py-0.5 rounded">
                          {asset.exchange}
                        </span>
                        {asset.type && (
                          <span className="text-xs bg-blue-100 text-blue-600 px-1.5 py-0.5 rounded">
                            {asset.type.toUpperCase()}
                          </span>
                        )}
                      </div>
                      <p className="text-sm text-gray-600 truncate mt-0.5">
                        {asset.name}
                      </p>
                      {asset.sector && (
                        <p className="text-xs text-gray-500 mt-0.5">
                          {asset.sector}
                        </p>
                      )}
                    </div>
                    <div className="flex items-center space-x-2 ml-2">
                      {/* Price info for trending assets */}
                      {'current_price' in asset && (
                        <div className="text-right">
                          <div className="text-sm font-medium text-gray-900">
                            {formatPrice(asset.current_price as number)}
                          </div>
                          {'change_percent' in asset && (
                            <div className={`text-xs ${
                              ((asset as any).change_percent || 0) >= 0 ? 'text-green-600' : 'text-red-600'
                            }`}>
                              {formatChangeForAsset(asset)}
                            </div>
                          )}
                        </div>
                      )}
                      {enableDetailsModal && (
                        <button
                          onClick={(e) => {
                            e.stopPropagation();
                            handleViewDetails(asset.symbol);
                          }}
                          className="p-1.5 text-gray-400 hover:text-blue-600 hover:bg-blue-50 rounded-full transition-colors"
                          title="View Details"
                        >
                          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
                          </svg>
                        </button>
                      )}
                    </div>
                  </div>
                </div>
              ))}
            </>
          ) : query.length >= 2 ? (
            <div className="p-4 text-center text-gray-500">
              No assets found for "{query}"
            </div>
          ) : (
            <div className="p-4 text-center text-gray-500">
              Start typing to search for assets
            </div>
          )}
        </div>
      )}

      {/* Asset Details Modal */}
      {showAssetModal && selectedAsset && (
        <AssetDetailsModal
          symbol={selectedAsset}
          isOpen={showAssetModal}
          onClose={() => {
            setShowAssetModal(false);
            setSelectedAsset(null);
          }}
          onAddToPortfolio={handleAssetFromModal}
        />
      )}
    </div>
  );
};

export default AssetSelector;
