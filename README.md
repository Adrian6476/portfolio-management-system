# Portfolio Management System

A comprehensive, real-time portfolio management platform built with Next.js, Go, PostgreSQL, Redis, and NATS. Features live market data integration, advanced analytics, risk assessment, and WebSocket-powered real-time updates.

## 🏗️ Architecture

This system implements a modern architecture with real-time capabilities:

- **Frontend**: Next.js 14 with TypeScript, Tailwind CSS, React Query for state management, and Recharts for data visualization
- **API Gateway**: Go-based unified service with Gin framework handling routing, WebSocket connections, and API orchestration
- **Database**: PostgreSQL for persistent data storage with comprehensive schema
- **Market Data**: Finnhub API integration for live stock prices and market data
- **Real-time Updates**: WebSocket implementation for live portfolio and price updates
- **Containerization**: Docker & Docker Compose for development and deployment
- **Monorepo Structure**: pnpm workspace configuration for coordinated development

## 🚀 Quick Start

### Prerequisites

- Node.js >= 18.0.0
- pnpm >= 8.0.0
- Go >= 1.21
- Docker & Docker Compose
- Git
- Finnhub API Key (for market data)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/Adrian6476/portfolio-management-system.git
   cd portfolio-management-system
   ```

2. **Set up environment variables**
   ```bash
   cp .env.example .env
   # Edit .env and add your Finnhub API key
   ```

3. **Install dependencies**
   ```bash
   pnpm install
   ```

4. **Start the development environment**
   ```bash
   # Complete setup (install + databases + services + git hooks)
   pnpm run setup
   
   # Or start services individually
   pnpm run setup:db  # Start databases only
   pnpm run dev       # Start all services
   ```

5. **Access the application**
   - Frontend: http://localhost:3000
   - API Gateway: http://localhost:8080
   - Alternative Dashboard: http://localhost:3000/dashboard
   - Market Data Page: http://localhost:3000/market-data

## 📋 Available Scripts

### Root Level Commands
```bash
pnpm dev                    # Start all services (frontend + backend)
pnpm dev:frontend          # Start frontend only
pnpm dev:services          # Start backend services only
pnpm build                 # Build all services
pnpm build:frontend        # Build frontend only
pnpm build:services        # Build backend services
pnpm clean                 # Clean all build artifacts
pnpm setup                 # Complete setup (install + databases + services + git hooks)
pnpm setup:db              # Start databases only
pnpm setup:git             # Configure Git hooks and commit template
pnpm test                  # Run all tests
pnpm test:frontend         # Run frontend tests only
pnpm test:services         # Run backend tests only
```

### Frontend Commands
```bash
cd frontend
pnpm dev                   # Start development server
pnpm build                 # Build for production
pnpm start                 # Start production server
pnpm lint                  # Run ESLint
pnpm test                  # Run Jest tests
pnpm test:watch            # Run tests in watch mode
pnpm test:coverage         # Run tests with coverage
```

### Backend Commands
```bash
# In services/api-gateway directory
go mod download            # Download dependencies
go run main.go            # Start API gateway
go build                  # Build service
go test ./...             # Run tests
```

## 🐳 Docker Commands

```bash
# Start all services
docker-compose up --build

# Start specific services
docker-compose up postgres redis nats

# View logs
docker-compose logs -f [service-name]

# Stop all services
docker-compose down

# Clean up (remove volumes)
docker-compose down -v --remove-orphans
```

## 🔧 Configuration

### Environment Variables

Create a `.env` file in the root directory based on `.env.example`:

```env
# Database Configuration
POSTGRES_URL=postgres://portfolio_user:portfolio_pass@localhost:5432/portfolio_db?sslmode=disable
REDIS_URL=localhost:6379
NATS_URL=nats://localhost:4222

# API Gateway Configuration
PORT=8080
ENVIRONMENT=development
LOG_LEVEL=info
JWT_SECRET=your-secret-key-change-in-production

# Frontend Configuration
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8084

# External APIs
FINNHUB_API_KEY=your_finnhub_api_key_here

# Docker Environment
COMPOSE_PROJECT_NAME=portfolio-management
```

### Getting a Finnhub API Key

1. Visit [Finnhub.io](https://finnhub.io/)
2. Sign up for a free account
3. Get your API key from the dashboard
4. Add it to your `.env` file

## 📁 Project Structure

```
portfolio-management-system/
├── frontend/                    # Next.js frontend application
│   ├── src/
│   │   ├── app/                # App Router pages and layouts
│   │   │   ├── page.tsx        # Main portfolio page
│   │   │   ├── dashboard/      # Alternative dashboard view
│   │   │   ├── market-data/    # Market data and asset search
│   │   │   └── __tests__/      # Page-level tests
│   │   ├── components/         # Reusable UI components
│   │   │   ├── PortfolioSummary.tsx     # Portfolio overview
│   │   │   ├── HoldingsTable.tsx       # Holdings management
│   │   │   ├── PortfolioChart.tsx      # Data visualization
│   │   │   ├── AnalyticsDashboard.tsx  # Risk and performance analytics
│   │   │   ├── AssetSearch.tsx         # Market data search
│   │   │   ├── ExportModal.tsx         # Data export functionality
│   │   │   ├── WhatIfAnalysisModal.tsx # Scenario analysis
│   │   │   ├── TransactionForm.tsx     # Transaction management
│   │   │   ├── NotificationCenter.tsx  # Real-time notifications
│   │   │   └── __tests__/              # Component tests
│   │   ├── hooks/              # Custom React hooks
│   │   │   ├── usePortfolio.ts         # Portfolio data management
│   │   │   ├── useAnalytics.ts         # Analytics and risk metrics
│   │   │   ├── useWebSocket.ts         # Real-time updates
│   │   │   ├── useMarketData.ts        # Market data integration
│   │   │   └── usePortfolioExport.ts   # Data export functionality
│   │   ├── lib/                # Utility libraries
│   │   │   ├── api.ts          # API client configuration
│   │   │   └── exportUtils.ts  # Data export utilities
│   │   └── types/              # TypeScript type definitions
│   │       └── portfolio.ts    # Shared type definitions
│   ├── __mocks__/              # Jest mocks
│   ├── jest.config.js          # Jest configuration
│   └── package.json
├── services/                   # Backend services
│   └── api-gateway/           # Unified API Gateway service
│       ├── internal/
│       │   ├── handlers/      # HTTP request handlers
│       │   ├── services/      # Business logic and external integrations
│       │   │   ├── services.go        # Service initialization
│       │   │   ├── finnhub.go         # Finnhub API integration
│       │   │   ├── websocket.go       # WebSocket hub
│       │   │   └── market_updater.go  # Real-time market updates
│       │   ├── middleware/    # HTTP middleware
│       │   └── config/        # Configuration management
│       ├── main.go           # Application entry point
│       ├── go.mod            # Go module definition
│       └── Dockerfile        # Container configuration
├── scripts/                   # Database and utility scripts
│   └── init-db.sql           # Database schema initialization
├── docs/                     # Project documentation
├── .github/                  # GitHub workflows and templates
├── docker-compose.yml        # Docker services configuration
├── docker-compose.test.yml   # Testing environment
├── .env.example             # Environment variables template
├── .commitlintrc.js         # Commit message linting
├── .gitmessage              # Git commit template
└── package.json             # Root package.json (monorepo)
```

## 🎯 Key Features

### ✅ Implemented Features

#### Portfolio Management
- **Portfolio Overview**: Real-time portfolio summary with total value, cost basis, and P&L
- **Holdings Management**: Add, edit, and delete portfolio holdings with real-time price updates
- **Transaction Tracking**: Complete transaction history with buy/sell operations
- **Asset Search**: Search and discover stocks, ETFs, and cryptocurrencies
- **Multiple Dashboard Views**: Main portfolio view and alternative grid dashboard

#### Real-time Data & Analytics
- **Live Market Data**: Real-time price updates via Finnhub API integration
- **WebSocket Updates**: Live portfolio value and price change notifications
- **Performance Analytics**: Portfolio performance metrics and top/worst performers
- **Risk Assessment**: Comprehensive risk analysis including VaR, Sharpe ratio, and beta
- **Asset Allocation**: Visual breakdown by asset type, sector, and individual holdings
- **Diversification Analysis**: Sector diversification and concentration risk metrics

#### Data Visualization & Export
- **Interactive Charts**: Portfolio allocation pie charts with multiple view modes
- **Performance Dashboards**: Risk metrics, volatility analysis, and market indicators
- **Data Export**: Export portfolio data in CSV and JSON formats
- **What-If Analysis**: Scenario planning and portfolio optimization tools

#### Market Data Integration
- **Market Overview**: Trending assets and market movers (gainers/losers)
- **Asset Details**: Comprehensive asset information with key financial metrics
- **Price History**: Historical price data and technical indicators
- **Real-time Updates**: Live price feeds with WebSocket connectivity

#### Technical Features
- **Responsive Design**: Mobile-first design with Tailwind CSS
- **Type Safety**: Full TypeScript implementation across frontend and backend
- **Testing**: Comprehensive test suite with Jest and React Testing Library
- **Error Handling**: Robust error handling with user-friendly messages
- **Loading States**: Smooth loading indicators and skeleton screens
- **Accessibility**: WCAG-compliant components with proper ARIA labels

### 🔄 Future Enhancements

#### Architecture & Infrastructure
- **Redis Integration**: Implement caching layer for improved performance and reduced database load
- **NATS Messaging**: Add event-driven communication for microservices architecture
- **True Microservices**: Split unified API Gateway into dedicated services (Portfolio, Market Data, Analytics, Notifications)
- **API Rate Limiting**: Implement Redis-based rate limiting for external API calls
- **Distributed Caching**: Use Redis for session management and real-time data caching

#### Features & Functionality
- **AI-powered Insights**: Portfolio optimization recommendations using machine learning
- **Automated Rebalancing**: Smart rebalancing suggestions based on target allocations
- **Advanced Charting**: Technical indicators, candlestick charts, and drawing tools
- **Portfolio Backtesting**: Historical performance simulation with different strategies
- **Multi-currency Support**: International markets and currency conversion
- **Social Trading**: Community features, portfolio sharing, and social insights
- **Mobile Application**: React Native or Flutter mobile app
- **Advanced Analytics**: Monte Carlo simulations, stress testing, and scenario analysis
- **Integration Hub**: Connect with brokers, banks, and other financial services
- **Alerts & Notifications**: Advanced alerting system with multiple delivery channels

## 🔗 API Endpoints

### Portfolio Management
- `GET /api/v1/portfolio` - Get user portfolio holdings
- `GET /api/v1/portfolio/summary` - Get comprehensive portfolio summary
- `GET /api/v1/portfolio/performance` - Get portfolio performance metrics
- `POST /api/v1/portfolio/holdings` - Add new holding to portfolio
- `PUT /api/v1/portfolio/holdings/:id` - Update existing holding
- `DELETE /api/v1/portfolio/holdings/:id` - Remove holding from portfolio

### Transactions
- `GET /api/v1/transactions` - Get transaction history
- `POST /api/v1/transactions` - Create new transaction
- `GET /api/v1/transactions/:id` - Get specific transaction
- `PUT /api/v1/transactions/:id` - Update transaction
- `DELETE /api/v1/transactions/:id` - Delete transaction

### Market Data
- `GET /api/v1/market/assets` - Search and get available assets
- `GET /api/v1/market/assets/:symbol` - Get detailed asset information
- `GET /api/v1/market/prices/:symbol` - Get current price and basic metrics
- `GET /api/v1/market/prices/:symbol/history` - Get historical price data

### Analytics
- `GET /api/v1/analytics/performance` - Get detailed performance analytics
- `GET /api/v1/analytics/risk` - Get comprehensive risk assessment
- `GET /api/v1/analytics/allocation` - Get asset allocation breakdown
- `POST /api/v1/analytics/whatif` - Perform what-if scenario analysis

### Notifications
- `GET /api/v1/notifications` - Get user notifications
- `PUT /api/v1/notifications/:id/read` - Mark notification as read
- `POST /api/v1/notifications/settings` - Update notification preferences

### Real-time Updates
- `GET /api/v1/ws` - WebSocket endpoint for real-time updates

### Health & Development
- `GET /health` - Service health check
- `POST /dev/sample-data` - Create sample portfolio data (development only)

## 🧪 Testing

The project includes comprehensive testing across both frontend and backend:

### Frontend Testing
```bash
# Run all frontend tests
cd frontend && pnpm test

# Run tests in watch mode
pnpm test:watch

# Run tests with coverage report
pnpm test:coverage

# Run specific test file
pnpm test PortfolioSummary.test.tsx
```

**Test Coverage Includes:**
- Component rendering and user interactions
- Custom hooks functionality
- API integration and error handling
- WebSocket connection management
- Data export utilities

### Backend Testing
```bash
# Run all backend tests
cd services/api-gateway && go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific test package
go test ./internal/handlers
```

**Test Coverage Includes:**
- HTTP handlers and API endpoints
- Database operations and queries
- Market data integration
- WebSocket functionality
- Portfolio calculations and analytics

### Integration Testing
```bash
# Run integration tests with Docker
docker-compose -f docker-compose.test.yml up --abort-on-container-exit

# Run all tests (frontend + backend)
pnpm test
```

### Test Files Structure
```
frontend/src/
├── components/__tests__/
│   ├── PortfolioSummary.test.tsx
│   ├── HoldingsTable.test.tsx
│   └── AddHoldingForm.test.tsx
└── app/__tests__/
    └── page.test.tsx

services/api-gateway/internal/
├── handlers/
│   ├── handlers_test.go
│   ├── portfolio_test.go
│   ├── analytics_test.go
│   └── transactions_test.go
```

## 📊 Monitoring & Health Checks

### Health Monitoring
- **API Health**: `GET /health` - Service health and status
- **Database Health**: Automatic PostgreSQL connection monitoring
- **Redis Health**: Cache service connectivity checks
- **NATS Health**: Message broker status monitoring

### Logging & Observability
- **Structured Logging**: JSON-formatted logs with Zap logger
- **Request Tracing**: Request ID tracking across services
- **Error Tracking**: Comprehensive error logging and handling
- **Performance Monitoring**: Response time and throughput metrics

### Real-time Status
- **WebSocket Status**: Live connection status indicators in UI
- **Market Data Status**: Real-time API connectivity monitoring
- **Database Performance**: Query performance and connection pooling
- **Service Dependencies**: External service availability tracking

### Development Tools
- **Hot Reload**: Automatic frontend and backend reloading
- **Database Migrations**: Automated schema updates
- **Docker Health Checks**: Container health monitoring
- **Git Hooks**: Pre-commit validation and formatting

## 🤝 Contributing

We use strict Git commit conventions to maintain clear and consistent code history.

### Git Commit Guidelines

This project follows [Conventional Commits](https://www.conventionalcommits.org/) specification:

```bash
# Format: <type>(<scope>): <subject>
feat(frontend): add portfolio dashboard
fix(api): resolve database connection issue
docs: update setup instructions
```

### Development Setup

1. Fork and clone the repository
2. Install dependencies and configure Git hooks:
   ```bash
   pnpm install
   pnpm run setup:git  # Configure commit template and validation
   ```
3. Create a feature branch: `git checkout -b feat/your-feature`
4. Follow commit conventions: `git commit` (displays template)
5. Push and create Pull Request

### Automatic Validation

- ✅ Auto-validate commit message format on each commit
- ✅ Git template guides correct format
- ✅ Detailed commit types and examples

For more details, see [`CONTRIBUTING.md`](./CONTRIBUTING.md)

## 📝 Development Notes

### Database Schema
- **Initial Schema**: Complete schema in [`scripts/init-db.sql`](scripts/init-db.sql)
- **Sample Data**: Includes default user and sample assets
- **Indexes**: Optimized indexes for query performance
- **Relationships**: Proper foreign key constraints and cascading deletes

### Architecture Decisions
- **Unified API Gateway**: Single Go service handling all backend logic
- **Real-time Updates**: WebSocket integration for live data
- **Market Data**: Finnhub API for reliable financial data
- **State Management**: React Query for server state, local state for UI
- **Type Safety**: Shared TypeScript types between frontend and backend

### Code Organization
- **Component Structure**: Reusable UI components with proper separation
- **Custom Hooks**: Business logic abstracted into reusable hooks
- **Error Boundaries**: Comprehensive error handling throughout the app
- **Loading States**: Consistent loading indicators and skeleton screens

### Performance Optimizations
- **Dynamic Imports**: Code splitting for better initial load times
- **Query Caching**: React Query for efficient data fetching and caching
- **WebSocket Efficiency**: Selective subscriptions and message handling
- **Database Optimization**: Proper indexing and query optimization

### Development Workflow
- **Git Hooks**: Automated commit message validation
- **Code Formatting**: Consistent code style enforcement
- **Testing Strategy**: Unit tests for components and integration tests for APIs
- **Docker Development**: Containerized development environment

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙋‍♂️ Support

For questions and support:
- Open an issue on GitHub
- Check the documentation
- Review existing issues and PRs

---

## 🚀 Deployment

### Production Deployment
```bash
# Build all services
pnpm build

# Start production environment
docker-compose up -d

# Check service health
curl http://localhost:8080/health
```

### Environment-Specific Configuration
- **Development**: Hot reload, debug logging, sample data
- **Production**: Optimized builds, structured logging, health monitoring
- **Testing**: Isolated database, mock external services

---

**Status**: ✅ **Production Ready** - Full-featured portfolio management system with real-time updates, comprehensive analytics, and robust testing.

**Live Features**: Portfolio management, real-time market data, risk analytics, data export, WebSocket updates, responsive design, and comprehensive testing suite.
