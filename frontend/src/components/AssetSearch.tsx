import React, { useState, useRef, useEffect } from 'react';
import { useAssetSearch, useTrendingAssets } from '@/hooks/useMarketData';
import { LoadingSpinner } from './ui';
import AssetDetailsModal from './AssetDetailsModal';

interface AssetSearchProps {
  onAssetSelect: (symbol: string, name: string) => void;
  placeholder?: string;
  className?: string;
  showTrending?: boolean;
  enableDetailsModal?: boolean;
}

const AssetSearch: React.FC<AssetSearchProps> = ({
  onAssetSelect,
  placeholder = "Search for stocks, ETFs, crypto...",
  className = "",
  showTrending = true,
  enableDetailsModal = true
}) => {
  const [query, setQuery] = useState('');
  const [isOpen, setIsOpen] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(-1);
  const [selectedAsset, setSelectedAsset] = useState<string | null>(null);
  const [showAssetModal, setShowAssetModal] = useState(false);
  const searchRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const { data: searchResults = [], isLoading } = useAssetSearch(query, query.length >= 2);
  const { data: trendingAssets = [] } = useTrendingAssets();

  const displayResults = query.length >= 2 ? searchResults : (showTrending ? trendingAssets.slice(0, 8) : []);

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
          if (enableDetailsModal) {
            handleViewDetails(asset.symbol);
          } else {
            handleAssetSelect(asset.symbol, asset.name);
          }
        }
        break;
      case 'Escape':
        setIsOpen(false);
        setSelectedIndex(-1);
        inputRef.current?.blur();
        break;
    }
  };

  const handleAssetSelect = (symbol: string, name: string) => {
    onAssetSelect(symbol, name);
    setQuery('');
    setIsOpen(false);
    setSelectedIndex(-1);
  };

  const handleViewDetails = (symbol: string) => {
    setSelectedAsset(symbol);
    setShowAssetModal(true);
    setIsOpen(false);
  };

  const handleAssetFromModal = (symbol: string, name: string) => {
    onAssetSelect(symbol, name);
    setShowAssetModal(false);
    setSelectedAsset(null);
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setQuery(value);
    setSelectedIndex(-1);
    setIsOpen(true);
  };

  const getAssetTypeIcon = (type: string) => {
    switch (type) {
      case 'stock':
        return (
          <div className="w-8 h-8 bg-blue-100 rounded-full flex items-center justify-center">
            <svg className="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z" />
            </svg>
          </div>
        );
      case 'etf':
        return (
          <div className="w-8 h-8 bg-green-100 rounded-full flex items-center justify-center">
            <svg className="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 11H5m14 0a2 2 0 012 2v6a2 2 0 01-2 2H5a2 2 0 01-2-2v-6a2 2 0 012-2m14 0V9a2 2 0 00-2-2M5 11V9a2 2 0 012-2m0 0V5a2 2 0 012-2h6a2 2 0 012 2v2M7 7h10" />
            </svg>
          </div>
        );
      case 'crypto':
        return (
          <div className="w-8 h-8 bg-orange-100 rounded-full flex items-center justify-center">
            <svg className="w-4 h-4 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1" />
            </svg>
          </div>
        );
      default:
        return (
          <div className="w-8 h-8 bg-gray-100 rounded-full flex items-center justify-center">
            <svg className="w-4 h-4 text-gray-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M13 16h-1v-4h-1m1-4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
          </div>
        );
    }
  };

  return (
    <div className={`relative ${className}`} ref={searchRef}>
      {/* Search Input */}
      <div className="relative">
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={handleInputChange}
          onFocus={() => setIsOpen(true)}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          className="w-full px-4 py-2 pl-10 pr-4 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
        />
        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
          <svg className="h-5 w-5 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
          </svg>
        </div>
        {isLoading && (
          <div className="absolute inset-y-0 right-0 pr-3 flex items-center">
            <LoadingSpinner size="sm" />
          </div>
        )}
      </div>

      {/* Search Results Dropdown */}
      {isOpen && (
        <div className="absolute z-50 w-full mt-1 bg-white rounded-md shadow-lg border border-gray-200 max-h-96 overflow-y-auto">
          {query.length < 2 && showTrending && trendingAssets.length > 0 && (
            <div className="px-4 py-2 text-sm text-gray-500 border-b border-gray-100">
              Trending Assets
            </div>
          )}
          
          {displayResults.length === 0 ? (
            <div className="px-4 py-8 text-center text-gray-500">
              {query.length >= 2 ? (
                isLoading ? (
                  <div className="flex items-center justify-center">
                    <LoadingSpinner size="sm" />
                    <span className="ml-2">Searching...</span>
                  </div>
                ) : (
                  <>
                    <svg className="mx-auto h-8 w-8 text-gray-400 mb-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                      <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
                    </svg>
                    <p>No assets found for "{query}"</p>
                    <p className="text-xs mt-1">Try searching for a stock symbol or company name</p>
                  </>
                )
              ) : (
                showTrending ? "Start typing to search for assets..." : "Type at least 2 characters to search"
              )}
            </div>
          ) : (
            displayResults.map((asset, index) => (
              <div
                key={asset.symbol}
                className={`px-4 py-3 border-b border-gray-100 last:border-b-0 transition-colors ${
                  selectedIndex === index ? 'bg-blue-50' : 'hover:bg-gray-50'
                }`}
              >
                <div className="flex items-center">
                  {getAssetTypeIcon(asset.type)}
                  
                  <div 
                    className="ml-3 flex-1 min-w-0 cursor-pointer"
                    onClick={() => enableDetailsModal ? handleViewDetails(asset.symbol) : handleAssetSelect(asset.symbol, asset.name)}
                    title={enableDetailsModal ? "Click to view details" : "Click to add to portfolio"}
                  >
                    <div className="flex items-center justify-between">
                      <p className="text-sm font-medium text-gray-900 truncate">
                        {asset.symbol}
                      </p>
                      <span className={`text-xs px-2 py-1 rounded-full ${
                        asset.type === 'stock' ? 'bg-blue-100 text-blue-800' :
                        asset.type === 'etf' ? 'bg-green-100 text-green-800' :
                        asset.type === 'crypto' ? 'bg-orange-100 text-orange-800' :
                        'bg-gray-100 text-gray-800'
                      }`}>
                        {asset.type.toUpperCase()}
                      </span>
                    </div>
                    <p className="text-sm text-gray-600 truncate">{asset.name}</p>
                    <div className="flex items-center mt-1">
                      <span className="text-xs text-gray-500">{asset.exchange}</span>
                      {asset.sector && (
                        <>
                          <span className="text-xs text-gray-400 mx-1">•</span>
                          <span className="text-xs text-gray-500">{asset.sector}</span>
                        </>
                      )}
                    </div>
                  </div>

                  {enableDetailsModal && (
                    <div className="ml-2 flex space-x-1">
                      <button
                        onClick={(e) => {
                          e.stopPropagation();
                          handleAssetSelect(asset.symbol, asset.name);
                        }}
                        className="p-1.5 text-gray-400 hover:text-green-600 hover:bg-green-50 rounded transition-colors"
                        title="Add to Portfolio"
                      >
                        <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                          <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 4v16m8-8H4" />
                        </svg>
                      </button>
                    </div>
                  )}
                </div>
              </div>
            ))
          )}
        </div>
      )}

      {/* Asset Details Modal */}
      {selectedAsset && (
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

export default AssetSearch;
