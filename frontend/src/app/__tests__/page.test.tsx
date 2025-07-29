import { render, screen } from '@testing-library/react'
import '@testing-library/jest-dom'
import { QueryClient, QueryClientProvider } from 'react-query'
import Page from '../page'

// Mock Next.js router
jest.mock('next/navigation', () => ({
  useRouter: () => ({
    push: jest.fn(),
    replace: jest.fn(),
    prefetch: jest.fn(),
  }),
}))

// Mock DashboardLayout component
jest.mock('../../components/DashboardLayout', () => {
  return function MockDashboardLayout({ children, activeView, setActiveView }: any) {
    return (
      <div data-testid="dashboard-layout">
        <div data-testid="active-view">{activeView}</div>
        {children}
      </div>
    )
  }
})

// Mock PortfolioSummary component to avoid React Query issues
jest.mock('../../components/PortfolioSummary', () => {
  return function MockPortfolioSummary() {
    return (
      <div data-testid="portfolio-summary">
        <h3>Portfolio Summary</h3>
        <p>Developer A will implement this component</p>
        <div>API: GET /api/v1/portfolio/summary</div>
      </div>
    )
  }
})

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

describe('Dashboard Page', () => {
  beforeEach(() => {
    // Clear any previous mocks
    jest.clearAllMocks()
  })

  it('renders the dashboard page with layout', () => {
    renderWithQueryClient(<Page />)
    
    // Check if the dashboard layout is rendered
    expect(screen.getByTestId('dashboard-layout')).toBeInTheDocument()
  })

  it('starts with overview as default view', () => {
    renderWithQueryClient(<Page />)
    
    // Check if default view is overview
    expect(screen.getByTestId('active-view')).toHaveTextContent('overview')
  })

  it('renders PortfolioSummary in overview', () => {
    renderWithQueryClient(<Page />)
    expect(screen.getByTestId('portfolio-summary')).toBeInTheDocument()
  })

  it('does not show Developer C components in overview view', () => {
    renderWithQueryClient(<Page />)
    expect(screen.queryByTestId('add-holding-form')).not.toBeInTheDocument()
    expect(screen.queryByTestId('portfolio-chart')).not.toBeInTheDocument()
  })

  it('renders development status section in overview', () => {
    renderWithQueryClient(<Page />)
    
    // Check for development status
    expect(screen.getByText('Development Status')).toBeInTheDocument()
    expect(screen.getByText('Developer A Tasks')).toBeInTheDocument()
    expect(screen.getByText('Developer B Tasks')).toBeInTheDocument()
    expect(screen.getByText('Developer C Tasks')).toBeInTheDocument()
  })

  it('renders without console errors', () => {
    const consoleSpy = jest.spyOn(console, 'error').mockImplementation(() => {})
    
    renderWithQueryClient(<Page />)
    
    expect(consoleSpy).not.toHaveBeenCalled()
    consoleSpy.mockRestore()
  })
})