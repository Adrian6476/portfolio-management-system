import { render, screen, waitFor, fireEvent } from '@testing-library/react'

import '@testing-library/jest-dom'

import { QueryClient, QueryClientProvider } from 'react-query'

import HoldingsTable from '../HoldingsTable'

  

// Mock usePortfolio hooks

jest.mock('../../hooks/usePortfolio', () => ({

usePortfolioHoldings: jest.fn(),

useUpdateHolding: jest.fn(),

useDeleteHolding: jest.fn(),

useCurrentPrice: jest.fn(),

}))

  

// Helper function to render with QueryClient provider

const renderWithQueryClient = (component: React.ReactElement) => {

const queryClient = new QueryClient({

defaultOptions: {

queries: {

retry: false,

cacheTime: 0,

},

},

})

  

return render(

<QueryClientProvider client={queryClient}>

{component}

</QueryClientProvider>

)

}

  

describe('HoldingsTable', () => {

const mockHoldings = [

{

id: '1',

symbol: 'AAPL',

name: 'Apple Inc.',

quantity: 10,

average_cost: 150.25,

purchase_date: '2023-01-15',

},

{

id: '2',

symbol: 'GOOGL',

name: 'Alphabet Inc.',

quantity: 5,

average_cost: 1200.50,

purchase_date: '2023-02-20',

},

]

  

const mockPriceData = {

current_price: 175.30,

change: 5.25,

change_percent: 3.1,

}

  

beforeEach(() => {

jest.clearAllMocks()

// Mock usePortfolioHoldings

require('../../hooks/usePortfolio').usePortfolioHoldings.mockReturnValue({

data: { holdings: mockHoldings },

isLoading: false,

error: null,

})

  

// Mock useCurrentPrice

require('../../hooks/usePortfolio').useCurrentPrice.mockReturnValue({

data: mockPriceData,

isLoading: false,

error: null,

})

  

// Mock mutation hooks

require('../../hooks/usePortfolio').useUpdateHolding.mockReturnValue({

mutate: jest.fn(),

})

  

require('../../hooks/usePortfolio').useDeleteHolding.mockReturnValue({

mutate: jest.fn(),

})

})

  

it('renders loading state with accessibility attributes', async () => {

require('../../hooks/usePortfolio').usePortfolioHoldings.mockReturnValue({

isLoading: true,

})

  

renderWithQueryClient(<HoldingsTable />)

const loadingElement = screen.getByRole('status')

expect(loadingElement).toBeInTheDocument()

expect(loadingElement).toHaveAttribute('aria-label', 'Loading holdings data')

expect(screen.getByText('Loading...')).toBeInTheDocument()

})

  

it('handles empty holdings state', () => {

require('../../hooks/usePortfolio').usePortfolioHoldings.mockReturnValue({

data: { holdings: [] },

isLoading: false,

error: null

})

  

renderWithQueryClient(<HoldingsTable />)

expect(screen.getByText('No holdings found')).toBeInTheDocument()

})

  

it('handles edit modal loading state', async () => {

const mockMutate = jest.fn()

require('../../hooks/usePortfolio').useUpdateHolding.mockReturnValue({

mutate: mockMutate,

isLoading: false

})

  

renderWithQueryClient(<HoldingsTable />)

fireEvent.click(screen.getByTestId('edit-btn-1'))

expect(screen.getByText('Edit Holding')).toBeInTheDocument()

// Test form submission

fireEvent.change(screen.getByLabelText('Quantity'), {target: {value: '20'}})

fireEvent.click(screen.getByText('Save'))

expect(mockMutate).toHaveBeenCalled()

})

  

it('renders error state', () => {

require('../../hooks/usePortfolio').usePortfolioHoldings.mockReturnValue({

error: { message: 'Failed to load' },

})

  

renderWithQueryClient(<HoldingsTable />)

expect(screen.getByText(/Failed to load/)).toBeInTheDocument()

})

  

it('renders empty state', () => {

require('../../hooks/usePortfolio').usePortfolioHoldings.mockReturnValue({

data: { holdings: [] },

})

  

renderWithQueryClient(<HoldingsTable />)

expect(screen.getByText(/No holdings found/)).toBeInTheDocument()

})

  

it('opens edit modal when edit button clicked', async () => {

renderWithQueryClient(<HoldingsTable />)

  

await waitFor(() => {

fireEvent.click(screen.getByTestId('edit-btn-1'))

expect(screen.getByText('Edit Holding')).toBeInTheDocument()

})

})

  

it('opens delete confirmation when delete button clicked', async () => {

renderWithQueryClient(<HoldingsTable />)

  

await waitFor(() => {

fireEvent.click(screen.getByTestId('delete-btn-1'))

expect(screen.getByText('Delete Holding')).toBeInTheDocument()

expect(screen.getByText(/Are you sure you want to delete/)).toBeInTheDocument()

})

})

})