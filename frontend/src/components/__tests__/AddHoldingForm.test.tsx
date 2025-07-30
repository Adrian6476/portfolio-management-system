import { render, screen, waitFor, fireEvent, act } from '@testing-library/react'
import '@testing-library/jest-dom'
import { QueryClient, QueryClientProvider } from 'react-query'
import AddHoldingForm from '../AddHoldingForm'
import apiClient from '@/lib/api'

// Mock API client
jest.mock('@/lib/api')
const mockedApiClient = apiClient as jest.Mocked<typeof apiClient>

// Mock useMutation
jest.mock('react-query', () => ({
  ...jest.requireActual('react-query'),
  useMutation: jest.fn(),
}))

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

describe('AddHoldingForm', () => {
  const mockMutate = jest.fn()
  
  beforeEach(() => {
    jest.clearAllMocks()
    require('react-query').useMutation.mockReturnValue({
      mutate: mockMutate,
      isLoading: false,
      isError: false,
      error: null,
    })
  })

  it('renders form with all fields', () => {
    renderWithQueryClient(<AddHoldingForm />)
    
    expect(screen.getByLabelText('Symbol')).toBeInTheDocument()
    expect(screen.getByLabelText('Quantity')).toBeInTheDocument()
    expect(screen.getByLabelText('Average Cost')).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Add Holding/i })).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Reset/i })).toBeInTheDocument()
  })

  it('shows validation errors for empty fields', async () => {
    renderWithQueryClient(<AddHoldingForm />)
    
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: /Add Holding/i }))
    })
    
    await waitFor(() => {
      expect(screen.getByText('Symbol is required')).toBeInTheDocument()
      expect(screen.getByText('Quantity must be a number')).toBeInTheDocument()
      expect(screen.getByText('Average cost must be a number')).toBeInTheDocument()
    })
  })

  it('validates number inputs', async () => {
    renderWithQueryClient(<AddHoldingForm />)
    
    await act(async () => {
      fireEvent.change(screen.getByLabelText('Quantity'), { target: { value: '-10' } })
      fireEvent.change(screen.getByLabelText('Average Cost'), { target: { value: '0' } })
      fireEvent.click(screen.getByRole('button', { name: /Add Holding/i }))
    })
    
    await waitFor(() => {
      expect(screen.getByText('Quantity must be positive')).toBeInTheDocument()
      expect(screen.getByText('Average cost must be positive')).toBeInTheDocument()
    })
  })

  it('submits valid form data', async () => {
    renderWithQueryClient(<AddHoldingForm />)
    
    await act(async () => {
      fireEvent.change(screen.getByLabelText('Symbol'), { target: { value: 'AAPL' } })
      fireEvent.change(screen.getByLabelText('Quantity'), { target: { value: '10' } })
      fireEvent.change(screen.getByLabelText('Average Cost'), { target: { value: '150.25' } })
      fireEvent.click(screen.getByRole('button', { name: /Add Holding/i }))
    })
    
    await waitFor(() => {
      expect(mockMutate).toHaveBeenCalledWith({
        symbol: 'AAPL',
        quantity: 10,
        average_cost: 150.25
      })
    })
  })

  it('shows loading state during submission', async () => {
    require('react-query').useMutation.mockReturnValue({
      mutate: mockMutate,
      isLoading: true,
    })

    renderWithQueryClient(<AddHoldingForm />)
    
    await act(async () => {
      fireEvent.change(screen.getByLabelText('Symbol'), { target: { value: 'AAPL' } })
      fireEvent.change(screen.getByLabelText('Quantity'), { target: { value: '10' } })
      fireEvent.change(screen.getByLabelText('Average Cost'), { target: { value: '150.25' } })
    })
    
    expect(screen.getByText(/Adding.../i)).toBeInTheDocument()
    expect(screen.getByRole('button', { name: /Adding.../i })).toBeDisabled()
  })

  it('resets form after successful submission', async () => {
    let onSuccessCallback: (() => void) | undefined
    
    // Mock useMutation to capture and trigger the onSuccess callback
    require('react-query').useMutation.mockImplementation((config: any) => {
      onSuccessCallback = config.onSuccess
      return {
        mutate: jest.fn(() => {
          // Simulate successful mutation by calling onSuccess
          if (onSuccessCallback) {
            onSuccessCallback()
          }
        }),
        isLoading: false,
        isError: false,
        error: null,
      }
    })
    
    renderWithQueryClient(<AddHoldingForm />)
    
    const symbolInput = screen.getByLabelText('Symbol') as HTMLInputElement
    const quantityInput = screen.getByLabelText('Quantity') as HTMLInputElement
    const costInput = screen.getByLabelText('Average Cost') as HTMLInputElement
    
    // Fill the form
    await act(async () => {
      fireEvent.change(symbolInput, { target: { value: 'AAPL' } })
      fireEvent.change(quantityInput, { target: { value: '10' } })
      fireEvent.change(costInput, { target: { value: '150.25' } })
    })
    
    // Submit the form
    await act(async () => {
      fireEvent.click(screen.getByRole('button', { name: /Add Holding/i }))
    })
    
    // Wait for form reset by checking input values
    await waitFor(() => {
      expect(symbolInput.value).toBe('')
      expect(quantityInput.value).toBe('')
      expect(costInput.value).toBe('')
    }, { timeout: 2000 })
  })

  it('handles API errors', async () => {
    const errorMessage = 'Failed to add holding'
    require('react-query').useMutation.mockReturnValue({
      mutate: mockMutate,
      isError: true,
      error: { message: errorMessage },
    })

    renderWithQueryClient(<AddHoldingForm />)
    
    // Submit form to trigger error display
    await act(async () => {
      fireEvent.change(screen.getByLabelText('Symbol'), { target: { value: 'AAPL' } })
      fireEvent.change(screen.getByLabelText('Quantity'), { target: { value: '10' } })
      fireEvent.change(screen.getByLabelText('Average Cost'), { target: { value: '150.25' } })
      fireEvent.click(screen.getByRole('button', { name: /Add Holding/i }))
    })
    
    // Error should be displayed in the UI
    await waitFor(() => {
      expect(screen.getByText(errorMessage)).toBeInTheDocument()
    })
  })

  it('resets form when reset button clicked', async () => {
    renderWithQueryClient(<AddHoldingForm />)
    
    const symbolInput = screen.getByLabelText('Symbol') as HTMLInputElement
    const quantityInput = screen.getByLabelText('Quantity') as HTMLInputElement
    const costInput = screen.getByLabelText('Average Cost') as HTMLInputElement
    
    await act(async () => {
      fireEvent.change(symbolInput, { target: { value: 'AAPL' } })
      fireEvent.change(quantityInput, { target: { value: '10' } })
      fireEvent.change(costInput, { target: { value: '150.25' } })
      fireEvent.click(screen.getByRole('button', { name: /Reset/i }))
    })
    
    await waitFor(() => {
      expect(symbolInput.value).toBe('')
      expect(quantityInput.value).toBe('')
      expect(costInput.value).toBe('')
    })
  })
})