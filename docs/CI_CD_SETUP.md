# CI/CD Pipeline Setup Guide

## 🎯 Overview

This document provides detailed instructions for setting up Continuous Integration and Continuous Deployment (CI/CD) pipeline for the Portfolio Management System.

## 🔧 GitHub Actions Setup

### 1. Create CI/CD Workflow File

Create `.github/workflows/ci.yml` with the following content:

```yaml
name: CI/CD Pipeline

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main, develop ]

env:
  NODE_VERSION: '18'
  GO_VERSION: '1.21'
  PNPM_VERSION: '8'

jobs:
  # Frontend Testing
  frontend-test:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Node.js
        uses: actions/setup-node@v4
        with:
          node-version: ${{ env.NODE_VERSION }}

      - name: Setup pnpm
        uses: pnpm/action-setup@v2
        with:
          version: ${{ env.PNPM_VERSION }}

      - name: Install dependencies
        run: |
          cd frontend
          pnpm install --frozen-lockfile

      - name: Type check
        run: |
          cd frontend
          pnpm run type-check

      - name: Lint
        run: |
          cd frontend
          pnpm run lint

      - name: Run tests
        run: |
          cd frontend
          pnpm run test:coverage

      - name: Upload coverage to Codecov
        uses: codecov/codecov-action@v3
        with:
          directory: ./frontend/coverage

  # Backend Testing
  backend-test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        service: [api-gateway, portfolio-service, market-data-service, analytics-service, notification-service]
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: ${{ env.GO_VERSION }}

      - name: Run tests
        run: |
          if [ -d "services/${{ matrix.service }}" ]; then
            cd services/${{ matrix.service }}
            go mod download
            go test -v -cover ./...
          fi

  # Integration Testing
  integration-test:
    runs-on: ubuntu-latest
    needs: [frontend-test, backend-test]
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Create .env file for testing
        run: |
          cat > .env << EOF
          POSTGRES_URL=postgres://portfolio_user:portfolio_pass@localhost:5432/portfolio_db?sslmode=disable
          REDIS_URL=localhost:6379
          NATS_URL=nats://localhost:4222
          PORT=8080
          ENVIRONMENT=test
          EOF

      - name: Start services
        run: |
          docker-compose -f docker-compose.test.yml up -d
          sleep 30  # Wait for services to be ready

      - name: Run integration tests
        run: |
          # Add integration test commands here
          echo "Integration tests placeholder"

      - name: Stop services
        run: |
          docker-compose -f docker-compose.test.yml down

  # Build and Deploy (on main branch)
  deploy:
    runs-on: ubuntu-latest
    needs: [frontend-test, backend-test, integration-test]
    if: github.ref == 'refs/heads/main' && github.event_name == 'push'
    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Build services
        run: |
          docker-compose build

      - name: Deploy to staging
        run: |
          echo "Deploy to staging environment"
          # Add deployment commands here
```

### 2. Create Test Docker Compose File

Create `docker-compose.test.yml`:

```yaml
version: '3.8'

services:
  postgres-test:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: portfolio_test_db
      POSTGRES_USER: portfolio_user
      POSTGRES_PASSWORD: portfolio_pass
    ports:
      - "5433:5432"
    volumes:
      - ./scripts/init-db.sql:/docker-entrypoint-initdb.d/init-db.sql

  redis-test:
    image: redis:7-alpine
    ports:
      - "6380:6379"

  nats-test:
    image: nats:2.10-alpine
    ports:
      - "4223:4222"
```

## 🔍 Quality Gates

### Code Quality Checks

1. **Linting Rules**
   - ESLint for TypeScript/JavaScript
   - gofmt for Go code
   - Prettier for formatting

2. **Testing Requirements**
   - Minimum 70% code coverage
   - All unit tests must pass
   - Integration tests must pass

3. **Security Scanning**
   - Dependency vulnerability scanning
   - Static code analysis

### Branch Protection Rules

Configure these rules in GitHub:

```yaml
main branch:
  - Require pull request reviews (minimum 1)
  - Require status checks to pass
  - Require branches to be up to date
  - Include administrators
  - Restrict force pushes

develop branch:
  - Require pull request reviews (minimum 1)
  - Require status checks to pass
  - Require branches to be up to date
```

## 🚀 Deployment Strategy

### Environments

1. **Development (Local)**
   - Developer machines
   - Docker Compose setup
   - Hot reloading enabled

2. **Staging (CI/CD)**
   - Automatic deployment on main branch
   - Integration testing environment
   - Mirror of production setup

3. **Production (Manual)**
   - Manual deployment approval
   - Health checks required
   - Rollback strategy in place

### Deployment Commands

```bash
# Development
pnpm run dev

# Staging deployment
docker-compose -f docker-compose.staging.yml up -d

# Production deployment
docker-compose -f docker-compose.prod.yml up -d
```

## 📊 Monitoring & Alerts

### Health Checks

Each service should expose health endpoints:

```go
// Example health check endpoint
func healthHandler(w http.ResponseWriter, r *http.Request) {
    health := map[string]string{
        "status": "healthy",
        "timestamp": time.Now().UTC().String(),
        "service": "api-gateway",
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(health)
}
```

### Logging Strategy

```yaml
# Structured logging format
{
  "timestamp": "2024-01-15T10:30:00Z",
  "level": "INFO",
  "service": "api-gateway",
  "message": "Request processed",
  "request_id": "req-123",
  "duration_ms": 45
}
```

## 🔧 Local Development Setup

### Pre-commit Hooks

Already configured with Husky:

```bash
# Install and configure
pnpm install
pnpm run setup:git

# Manual hook test
npx husky install
git add .
git commit -m "feat: test commit message validation"
```

### Environment Variables

Create `.env` file (use `.env.example` as template):

```env
# Required for development
POSTGRES_URL=postgres://portfolio_user:portfolio_pass@localhost:5432/portfolio_db?sslmode=disable
REDIS_URL=localhost:6379
NATS_URL=nats://localhost:4222
PORT=8080
ENVIRONMENT=development
LOG_LEVEL=debug

# Optional API keys
YAHOO_FINANCE_API_KEY=your-api-key-here
```

## 🧪 Testing Guidelines

### Frontend Testing

```bash
# Run specific test categories
cd frontend

# Unit tests
pnpm test -- --testPathPattern="unit"

# Integration tests
pnpm test -- --testPathPattern="integration"

# Component tests
pnpm test -- --testPathPattern="components"

# E2E tests (if implemented)
pnpm test:e2e
```

### Backend Testing

```bash
# Test specific service
cd services/portfolio-service
go test ./...

# Test with coverage
go test -cover ./...

# Benchmark tests
go test -bench=.

# Race condition detection
go test -race ./...
```

### Database Testing

```bash
# Run database tests
docker-compose exec postgres psql -U portfolio_user -d portfolio_db -f /docker-entrypoint-initdb.d/test-data.sql

# Reset test database
docker-compose down -v
docker-compose up -d postgres
```

## 🔒 Security Considerations

### Secrets Management

1. **Never commit secrets to Git**
2. **Use GitHub Secrets for CI/CD**
3. **Rotate secrets regularly**
4. **Use environment-specific secrets**

### GitHub Secrets Setup

Add these secrets in GitHub repository settings:

```yaml
# Database
POSTGRES_URL_STAGING
POSTGRES_URL_PRODUCTION

# API Keys
YAHOO_FINANCE_API_KEY

# Deployment
STAGING_SERVER_SSH_KEY
PRODUCTION_SERVER_SSH_KEY
```

## 📈 Performance Monitoring

### Metrics to Track

1. **Application Performance**
   - Response times
   - Error rates
   - Throughput

2. **Infrastructure Performance**
   - CPU usage
   - Memory usage
   - Database connections

3. **Business Metrics**
   - User engagement
   - Feature usage
   - Error frequency

### Monitoring Tools (Future Implementation)

- **Application Monitoring:** Prometheus + Grafana
- **Log Aggregation:** ELK Stack (Elasticsearch, Logstash, Kibana)
- **Error Tracking:** Sentry
- **Uptime Monitoring:** Simple HTTP checks

## 🔄 Continuous Improvement

### Weekly Reviews

1. **Performance Metrics Review**
2. **Test Coverage Analysis**
3. **Security Scan Results**
4. **Deployment Success Rate**
5. **Developer Experience Feedback**

### Optimization Areas

1. **Build Time Optimization**
2. **Test Execution Speed**
3. **Deployment Frequency**
4. **Code Quality Metrics**
5. **Documentation Coverage**

---

**Implementation Priority:**
1. Set up basic CI pipeline first
2. Add comprehensive testing
3. Implement automated deployment
4. Add monitoring and alerts
5. Optimize performance and security