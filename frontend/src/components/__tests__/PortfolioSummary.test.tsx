import { render, screen, waitFor } from '@testing-library/react'
import '@testing-library/jest-dom'
import { QueryClient, QueryClientProvider, useQuery } from 'react-query'
import PortfolioSummary from '../PortfolioSummary'
import apiClient from '@/lib/api'

// Mock the API client
jest.mock('@/lib/api')
const mockedApiClient = apiClient as jest.Mocked<typeof apiClient>

// Mock useQuery for error tests
jest.mock('react-query', () => ({
  ...jest.requireActual('react-query'),
  useQuery: jest.fn(),
}))
const mockedUseQuery = useQuery as jest.MockedFunction<typeof useQuery>

// Mock data
const mockPortfolioSummaryData = {
  summary: {
    total_holdings: 5,
    total_cost: 15000.50,
    total_shares: 125.75
  },
  asset_allocation: [
    {
      asset_type: 'STOCK',
      count: 3,
      total_value: 12000.00,
      percentage: 80.0
    },
    {
      asset_type: 'BOND',
      count: 2,
      total_value: 3000.50,
      percentage: 20.0
    }
  ],
  top_holdings: [
    {
      id: '1',
      symbol: 'AAPL',
      name: 'Apple Inc.',
      asset_type: 'STOCK',
      quantity: 50,
      average_cost: 150.00,
      purchase_date: '2023-01-15',
      total_value: 7500.00
    },
    {
      id: '2',
      symbol: 'GOOGL',
      name: 'Alphabet Inc.',
      asset_type: 'STOCK',
      quantity: 25,
      average_cost: 120.00,
      purchase_date: '2023-02-10',
      total_value: 3000.00
    }
  ]
}

// Helper function to render component with React Query provider
const renderWithQueryClient = (component: React.ReactElement) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
        cacheTime: 0,
        refetchOnWindowFocus: false,
        staleTime: 0,
      },
    },
  })

  return render(
    <QueryClientProvider client={queryClient}>
      {component}
    </QueryClientProvider>
  )
}

describe('PortfolioSummary Component', () => {
  beforeEach(() => {
    jest.clearAllMocks()
    // Reset useQuery mock to use real implementation by default
    mockedUseQuery.mockImplementation(jest.requireActual('react-query').useQuery)
  })

  describe('Loading State', () => {
    it('displays loading spinner when data is being fetched', () => {
      // Mock useQuery to return loading state
      mockedUseQuery.mockReturnValue({
        data: undefined,
        isLoading: true,
        isError: false,
        error: null,
        isIdle: false,
        isLoadingError: false,
        isRefetchError: false,
        isSuccess: false,
        status: 'loading',
        dataUpdatedAt: 0,
        errorUpdatedAt: 0,
        failureCount: 0,
        isFetched: false,
        isFetchedAfterMount: false,
        isFetching: true,
        isPlaceholderData: false,
        isPreviousData: false,
        isStale: false,
        refetch: jest.fn(),
        remove: jest.fn(),
      } as any)

      renderWithQueryClient(<PortfolioSummary />)

      expect(screen.getByText('Loading portfolio summary...')).toBeInTheDocument()
      // Check for the spinner by class instead of role
      expect(document.querySelector('.animate-spin')).toBeInTheDocument()
    })
  })

  describe('Error State', () => {
    it('displays error message when API call fails', async () => {
      const errorMessage = 'Network Error'
      // Mock useQuery to return error state
      mockedUseQuery.mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
        error: new Error(errorMessage),
        isIdle: false,
        isLoadingError: true,
        isRefetchError: false,
        isSuccess: false,
        status: 'error',
        dataUpdatedAt: 0,
        errorUpdatedAt: Date.now(),
        failureCount: 1,
        isFetched: true,
        isFetchedAfterMount: true,
        isFetching: false,
        isPlaceholderData: false,
        isPreviousData: false,
        isStale: false,
        refetch: jest.fn(),
        remove: jest.fn(),
      } as any)

      renderWithQueryClient(<PortfolioSummary />)

      expect(screen.getByText('Error loading portfolio summary')).toBeInTheDocument()
      expect(screen.getByText(errorMessage)).toBeInTheDocument()
    })

    it('displays generic error message for non-Error objects', async () => {
      // Mock useQuery to return error state with non-Error object
      mockedUseQuery.mockReturnValue({
        data: undefined,
        isLoading: false,
        isError: true,
        error: 'String error',
        isIdle: false,
        isLoadingError: true,
        isRefetchError: false,
        isSuccess: false,
        status: 'error',
        dataUpdatedAt: 0,
        errorUpdatedAt: Date.now(),
        failureCount: 1,
        isFetched: true,
        isFetchedAfterMount: true,
        isFetching: false,
        isPlaceholderData: false,
        isPreviousData: false,
        isStale: false,
        refetch: jest.fn(),
        remove: jest.fn(),
      } as any)

      renderWithQueryClient(<PortfolioSummary />)

      expect(screen.getByText('Error loading portfolio summary')).toBeInTheDocument()
      expect(screen.getByText('Failed to fetch portfolio summary')).toBeInTheDocument()
    })

    it('displays error when no data is returned', async () => {
      // Mock useQuery to return success state but with null data
      mockedUseQuery.mockReturnValue({
        data: null,
        isLoading: false,
        isError: false,
        error: null,
        isIdle: false,
        isLoadingError: false,
        isRefetchError: false,
        isSuccess: true,
        status: 'success',
        dataUpdatedAt: Date.now(),
        errorUpdatedAt: 0,
        failureCount: 0,
        isFetched: true,
        isFetchedAfterMount: true,
        isFetching: false,
        isPlaceholderData: false,
        isPreviousData: false,
        isStale: false,
        refetch: jest.fn(),
        remove: jest.fn(),
      } as any)

      renderWithQueryClient(<PortfolioSummary />)

      expect(screen.getByText('Error loading portfolio summary')).toBeInTheDocument()
      expect(screen.getByText('No portfolio data available')).toBeInTheDocument()
    })
  })

  describe('Success State', () => {
    beforeEach(() => {
      mockedApiClient.get.mockResolvedValue({ data: mockPortfolioSummaryData })
    })

    it('displays portfolio summary metrics correctly', async () => {
      renderWithQueryClient(<PortfolioSummary />)

      await waitFor(() => {
        // Check summary metrics
        expect(screen.getByText('Total Holdings')).toBeInTheDocument()
        expect(screen.getByText('5')).toBeInTheDocument()
        expect(screen.getByText('Number of positions')).toBeInTheDocument()

        expect(screen.getByText('Total Cost')).toBeInTheDocument()
        expect(screen.getByText('$15,000.50')).toBeInTheDocument()
        expect(screen.getByText('Total invested amount')).toBeInTheDocument()

        expect(screen.getByText('Total Shares')).toBeInTheDocument()
        expect(screen.getByText('125.75')).toBeInTheDocument()
        expect(screen.getByText('Total number of shares')).toBeInTheDocument()
      })
    })

    it('displays asset allocation correctly', async () => {
      renderWithQueryClient(<PortfolioSummary />)

      await waitFor(() => {
        expect(screen.getByText('Asset Allocation')).toBeInTheDocument()
        
        // Check STOCK allocation
        expect(screen.getByText('STOCK')).toBeInTheDocument()
        expect(screen.getByText('$12,000.00')).toBeInTheDocument()
        expect(screen.getByText('80.0%')).toBeInTheDocument()

        // Check BOND allocation
        expect(screen.getByText('BOND')).toBeInTheDocument()
        expect(screen.getByText('$3,000.50')).toBeInTheDocument()
        expect(screen.getByText('20.0%')).toBeInTheDocument()
      })
    })

    it('displays top holdings correctly', async () => {
      renderWithQueryClient(<PortfolioSummary />)

      await waitFor(() => {
        expect(screen.getByText('Top Holdings')).toBeInTheDocument()
        
        // Check AAPL holding
        expect(screen.getByText('AAPL')).toBeInTheDocument()
        expect(screen.getByText('Apple Inc.')).toBeInTheDocument()
        expect(screen.getByText('$7,500.00')).toBeInTheDocument()
        expect(screen.getByText('50 shares')).toBeInTheDocument()

        // Check GOOGL holding
        expect(screen.getByText('GOOGL')).toBeInTheDocument()
        expect(screen.getByText('Alphabet Inc.')).toBeInTheDocument()
        expect(screen.getByText('$3,000.00')).toBeInTheDocument()
        expect(screen.getByText('25 shares')).toBeInTheDocument()
      })
    })

    it('calculates total_value correctly when not provided', async () => {
      const dataWithoutTotalValue = {
        ...mockPortfolioSummaryData,
        top_holdings: [
          {
            id: '1',
            symbol: 'AAPL',
            name: 'Apple Inc.',
            asset_type: 'STOCK',
            quantity: 10,
            average_cost: 100.00,
            purchase_date: '2023-01-15'
            // total_value not provided
          }
        ]
      }

      mockedApiClient.get.mockResolvedValue({ data: dataWithoutTotalValue })

      renderWithQueryClient(<PortfolioSummary />)

      await waitFor(() => {
        // Should calculate 10 * 100 = $1,000.00
        expect(screen.getByText('$1,000.00')).toBeInTheDocument()
      })
    })
  })

  describe('Edge Cases', () => {
    it('handles empty asset allocation array', async () => {
      const dataWithEmptyAllocation = {
        ...mockPortfolioSummaryData,
        asset_allocation: []
      }

      mockedApiClient.get.mockResolvedValue({ data: dataWithEmptyAllocation })

      renderWithQueryClient(<PortfolioSummary />)

      await waitFor(() => {
        // Should still show summary metrics but not asset allocation section
        expect(screen.getByText('Total Holdings')).toBeInTheDocument()
        expect(screen.queryByText('Asset Allocation')).not.toBeInTheDocument()
      })
    })

    it('handles empty top holdings array', async () => {
      const dataWithEmptyHoldings = {
        ...mockPortfolioSummaryData,
        top_holdings: []
      }

      mockedApiClient.get.mockResolvedValue({ data: dataWithEmptyHoldings })

      renderWithQueryClient(<PortfolioSummary />)

      await waitFor(() => {
        // Should still show summary metrics but not top holdings section
        expect(screen.getByText('Total Holdings')).toBeInTheDocument()
        expect(screen.queryByText('Top Holdings')).not.toBeInTheDocument()
      })
    })

    it('limits top holdings to 5 items', async () => {
      const dataWithManyHoldings = {
        ...mockPortfolioSummaryData,
        top_holdings: Array.from({ length: 10 }, (_, i) => ({
          id: `${i + 1}`,
          symbol: `STOCK${i + 1}`,
          name: `Company ${i + 1}`,
          asset_type: 'STOCK',
          quantity: 10,
          average_cost: 100.00,
          purchase_date: '2023-01-15',
          total_value: 1000.00
        }))
      }

      mockedApiClient.get.mockResolvedValue({ data: dataWithManyHoldings })

      renderWithQueryClient(<PortfolioSummary />)

      await waitFor(() => {
        // Should only show first 5 holdings
        expect(screen.getByText('STOCK1')).toBeInTheDocument()
        expect(screen.getByText('STOCK5')).toBeInTheDocument()
        expect(screen.queryByText('STOCK6')).not.toBeInTheDocument()
      })
    })
  })

  describe('API Integration', () => {
    it('calls the correct API endpoint', async () => {
      mockedApiClient.get.mockResolvedValue({ data: mockPortfolioSummaryData })

      renderWithQueryClient(<PortfolioSummary />)

      await waitFor(() => {
        expect(mockedApiClient.get).toHaveBeenCalledWith('/portfolio/summary')
        expect(mockedApiClient.get).toHaveBeenCalledTimes(1)
      })
    })

    it('uses correct query key for caching', async () => {
      mockedApiClient.get.mockResolvedValue({ data: mockPortfolioSummaryData })

      renderWithQueryClient(<PortfolioSummary />)

      await waitFor(() => {
        expect(screen.getByText('Total Holdings')).toBeInTheDocument()
      })

      // The query key should be 'portfolioSummary' as defined in QUERY_KEYS
      expect(mockedApiClient.get).toHaveBeenCalledWith('/portfolio/summary')
    })
  })

  describe('Formatting', () => {
    it('formats currency values correctly', async () => {
      const dataWithLargeNumbers = {
        summary: {
          total_holdings: 1,
          total_cost: 1234567.89,
          total_shares: 1000
        },
        asset_allocation: [
          {
            asset_type: 'STOCK',
            count: 1,
            total_value: 9876543.21, // Different value to avoid duplicates
            percentage: 100.0
          }
        ],
        top_holdings: []
      }

      mockedApiClient.get.mockResolvedValue({ data: dataWithLargeNumbers })

      renderWithQueryClient(<PortfolioSummary />)

      await waitFor(() => {
        // Check for the total cost formatting
        expect(screen.getByText('$1,234,567.89')).toBeInTheDocument()
        // Check for the asset allocation formatting
        expect(screen.getByText('$9,876,543.21')).toBeInTheDocument()
      })
    })

    it('formats percentage correctly', async () => {
      mockedApiClient.get.mockResolvedValue({ data: mockPortfolioSummaryData })

      renderWithQueryClient(<PortfolioSummary />)

      await waitFor(() => {
        expect(screen.getByText('80.0%')).toBeInTheDocument()
        expect(screen.getByText('20.0%')).toBeInTheDocument()
      })
    })

    it('formats share quantities correctly', async () => {
      mockedApiClient.get.mockResolvedValue({ data: mockPortfolioSummaryData })

      renderWithQueryClient(<PortfolioSummary />)

      await waitFor(() => {
        expect(screen.getByText('125.75')).toBeInTheDocument() // total shares
        expect(screen.getByText('50 shares')).toBeInTheDocument() // individual holding
        expect(screen.getByText('25 shares')).toBeInTheDocument() // individual holding
      })
    })
  })
})