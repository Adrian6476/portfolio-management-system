# Portfolio Management System

A modern, microservices-based portfolio management platform built with Next.js, Go, PostgreSQL, Redis, and NATS.

## 🏗️ Architecture

This system follows a microservices architecture with event-driven communication:

- **Frontend**: Next.js 14 with TypeScript, Tailwind CSS, and Zustand for state management
- **API Gateway**: Go-based gateway handling routing, authentication, and API orchestration
- **Microservices**: Portfolio Service, Market Data Service, Analytics Service, Notification Service
- **Databases**: PostgreSQL for persistent data, Redis for caching
- **Message Broker**: NATS for event-driven communication
- **Containerization**: Docker & Docker Compose for development and deployment

## 🚀 Quick Start

### Prerequisites

- Node.js >= 18.0.0
- pnpm >= 8.0.0
- Go >= 1.21
- Docker & Docker Compose
- Git

### Installation

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd portfolio-management-system
   ```

2. **Install dependencies**
   ```bash
   pnpm install
   ```

3. **Start the development environment**
   ```bash
   # Start all services with Docker Compose
   pnpm run setup
   
   # Or start services individually
   pnpm run setup:db  # Start databases only
   pnpm run dev       # Start all services
   ```

4. **Access the application**
   - Frontend: http://localhost:3000
   - API Gateway: http://localhost:8080
   - API Documentation: http://localhost:8080/swagger (Coming Soon)

## 📋 Available Scripts

### Root Level Commands
```bash
pnpm dev                    # Start all services (frontend + backend)
pnpm dev:frontend          # Start frontend only
pnpm dev:services          # Start backend services only
pnpm build                 # Build all services
pnpm clean                 # Clean all build artifacts
pnpm setup                 # Complete setup (install + databases + services)
pnpm test                  # Run all tests
```

### Frontend Commands
```bash
cd frontend
pnpm dev                   # Start development server
pnpm build                 # Build for production
pnpm start                 # Start production server
pnpm lint                  # Run ESLint
pnpm test                  # Run tests
```

### Backend Commands
```bash
# In each service directory
go mod download            # Download dependencies
go run main.go            # Start service
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

Create a `.env` file in the root directory:

```env
# Database
POSTGRES_URL=postgres://portfolio_user:portfolio_pass@localhost:5432/portfolio_db?sslmode=disable
REDIS_URL=localhost:6379
NATS_URL=nats://localhost:4222

# API Gateway
PORT=8080
ENVIRONMENT=development
LOG_LEVEL=info
JWT_SECRET=your-secret-key-change-in-production

# Frontend
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_WS_URL=ws://localhost:8084

# External APIs
YAHOO_FINANCE_API_KEY=your-api-key
```

## 📁 Project Structure

```
portfolio-management-system/
├── frontend/                    # Next.js frontend application
│   ├── src/
│   │   ├── app/                # App Router pages
│   │   ├── components/         # Reusable UI components
│   │   ├── lib/               # Utility libraries
│   │   ├── hooks/             # Custom React hooks
│   │   ├── store/             # Zustand stores
│   │   └── types/             # TypeScript type definitions
│   ├── public/                # Static assets
│   └── package.json
├── services/                   # Backend microservices
│   ├── api-gateway/           # API Gateway service
│   ├── portfolio-service/     # Portfolio management service
│   ├── market-data-service/   # Market data service
│   ├── analytics-service/     # Analytics and calculations
│   └── notification-service/  # Real-time notifications
├── scripts/                   # Database and utility scripts
│   └── init-db.sql           # Database initialization
├── docker-compose.yml         # Docker services configuration
└── package.json              # Root package.json (monorepo)
```

## 🎯 Key Features

### Current Implementation
- ✅ Microservices architecture setup
- ✅ API Gateway with routing
- ✅ Database schema and initialization
- ✅ Docker containerization
- ✅ Development environment configuration
- ✅ Frontend structure with Next.js 14

### Planned Features
- 🔄 Portfolio CRUD operations
- 🔄 Real-time market data integration
- 🔄 Performance analytics and charts
- 🔄 Risk management calculations
- 🔄 WebSocket real-time updates
- 🔄 AI-powered insights
- 🔄 What-if analysis
- 🔄 Automated rebalancing suggestions

## 🔗 API Endpoints

### Portfolio Management
- `GET /api/v1/portfolio` - Get user portfolio
- `GET /api/v1/portfolio/summary` - Get portfolio summary
- `POST /api/v1/portfolio/holdings` - Add new holding
- `PUT /api/v1/portfolio/holdings/:id` - Update holding
- `DELETE /api/v1/portfolio/holdings/:id` - Remove holding

### Market Data
- `GET /api/v1/market/assets` - Get available assets
- `GET /api/v1/market/prices/:symbol` - Get current price
- `GET /api/v1/market/prices/:symbol/history` - Get price history

### Analytics
- `GET /api/v1/analytics/performance` - Get performance metrics
- `GET /api/v1/analytics/risk` - Get risk analysis
- `POST /api/v1/analytics/whatif` - What-if analysis

## 🧪 Testing

```bash
# Run all tests
pnpm test

# Frontend tests
cd frontend && pnpm test

# Backend tests
cd services/api-gateway && go test ./...

# Integration tests
docker-compose -f docker-compose.test.yml up --abort-on-container-exit
```

## 📊 Monitoring & Health Checks

- Health Check: `GET /health`
- Service Status: Each service exposes health endpoints
- Logging: Structured logging with request tracing
- Metrics: Prometheus-compatible metrics (Coming Soon)

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

### Database Migrations
- Initial schema is in `scripts/init-db.sql`
- Future migrations should be versioned
- Use database transactions for complex migrations

### Adding New Services
1. Create service directory in `services/`
2. Add Go module with `go mod init`
3. Implement service with health checks
4. Add to `docker-compose.yml`
5. Update API Gateway routing

### Frontend Development
- Use TypeScript for all new code
- Follow existing component patterns
- Add proper error handling
- Include loading states

## 📄 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙋‍♂️ Support

For questions and support:
- Open an issue on GitHub
- Check the documentation
- Review existing issues and PRs

---

**Status**: 🚧 Under Development - Basic environment setup complete, core features in progress.
