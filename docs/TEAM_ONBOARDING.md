# Team Onboarding Guide

## 🎯 Welcome to Portfolio Management System

Welcome to the team! This guide will help you get set up and productive quickly.

## 📋 Pre-Onboarding Checklist

### Required Accounts & Access
- [ ] GitHub account created
- [ ] Added to repository as collaborator
- [ ] Git configured with your name and email
- [ ] Communication channel access (Discord/Slack/Teams)

### Development Environment
- [ ] Node.js 18+ installed
- [ ] pnpm 8+ installed
- [ ] Go 1.21+ installed
- [ ] Docker & Docker Compose installed
- [ ] VS Code or preferred IDE configured

## 🚀 Quick Start (30 minutes)

### 1. Clone Repository
```bash
git clone https://github.com/Adrian6476/portfolio-management-system.git
cd portfolio-management-system
```

### 2. Install Dependencies
```bash
# Install all dependencies
pnpm install

# Set up Git hooks and commit template
pnpm run setup:git
```

### 3. Environment Setup
```bash
# Copy environment template
cp .env.example .env

# Edit .env with your local configuration
# Most defaults should work for development
```

### 4. Start Development Environment
```bash
# Start all services (this may take a few minutes first time)
pnpm run setup

# Once setup is complete, start development
pnpm run dev
```

### 5. Verify Setup
- Frontend: http://localhost:3000
- API Gateway: http://localhost:8080
- Database: localhost:5432 (portfolio_db)
- Redis: localhost:6379

## 👥 Your Role & Responsibilities

### Backend Developer 1 (Go Services Focus)
**Primary Services:**
- [`services/portfolio-service/`](../services/portfolio-service/) - Portfolio CRUD operations
- [`services/market-data-service/`](../services/market-data-service/) - Market data integration

**Key Responsibilities:**
- Implement portfolio holdings management
- Integrate Yahoo Finance API for market data
- Database schema design and optimization
- Write comprehensive unit tests

**Daily Tasks:**
- Check assigned GitHub issues
- Update issue status and comments
- Participate in daily standup
- Code review for other team members

### Backend Developer 2 (API Gateway & Analytics)
**Primary Services:**
- [`services/api-gateway/`](../services/api-gateway/) - Main API entry point
- [`services/analytics-service/`](../services/analytics-service/) - Performance calculations
- [`services/notification-service/`](../services/notification-service/) - Real-time notifications

**Key Responsibilities:**
- API routing and middleware
- Performance analytics calculations
- Real-time WebSocket connections
- Event-driven architecture with NATS

### Frontend Developer (Next.js)
**Primary Focus:**
- [`frontend/`](../frontend/) - Complete frontend application

**Key Responsibilities:**
- Portfolio dashboard UI/UX
- Interactive charts and data visualization
- Real-time data updates
- Responsive design implementation

**Tech Stack:**
- Next.js 14 with App Router
- TypeScript for type safety
- Tailwind CSS for styling
- Zustand for state management
- D3.js/Recharts for charts
- React Query for API calls

### Team Lead (Full-Stack)
**Primary Focus:**
- Overall project coordination
- Architecture decisions
- Code review and integration
- DevOps and deployment

**Key Responsibilities:**
- Sprint planning and task assignment
- Daily standup facilitation
- Cross-service integration
- Final code review and merging

## 🔧 Development Workflow

### 1. Pick Your First Task
1. Go to [GitHub Issues](../../issues)
2. Look for issues labeled with your focus area:
   - `backend` for backend developers
   - `frontend` for frontend developer
   - `good-first-issue` for first-time contributors
3. Assign yourself to an issue
4. Move it to "In Progress" on the project board

### 2. Create Feature Branch
```bash
# Always start from main
git checkout main
git pull origin main

# Create feature branch
git checkout -b feature/your-feature-name

# Example branch names:
# feature/portfolio-crud-operations
# feature/market-data-integration
# feature/dashboard-charts
```

### 3. Development Process
```bash
# Make your changes
# Run tests frequently
pnpm test

# Commit following conventions
git add .
git commit  # This will show commit template

# Example good commit:
# feat(portfolio): add CRUD operations for holdings
```

### 4. Submit Pull Request
```bash
# Push your branch
git push origin feature/your-feature-name

# Create PR on GitHub
# Use the PR template that appears
# Request review from team lead
```

## 📚 Key Documentation

### Must Read First
1. [**Team Collaboration Guide**](./TEAM_COLLABORATION_GUIDE.md) - Complete collaboration setup
2. [**Project Planning Documents**](../Project%20Planning/) - Project requirements and architecture
3. [**Contributing Guidelines**](../CONTRIBUTING.md) - Git commit standards

### Reference Documentation
4. [**CI/CD Setup**](./CI_CD_SETUP.md) - Automated testing and deployment
5. [**Project Management Templates**](./PROJECT_MANAGEMENT_TEMPLATES.md) - Sprint planning and templates
6. [**Main README**](../README.md) - Project overview and commands

## 🔍 Code Standards & Guidelines

### Frontend Code Standards (Next.js/TypeScript)
```typescript
// Use TypeScript for all new code
interface Portfolio {
  id: string;
  holdings: Holding[];
  totalValue: number;
}

// Use proper component structure
export const PortfolioDashboard: React.FC = () => {
  return (
    <div className="p-6">
      <h1 className="text-2xl font-bold">Portfolio Dashboard</h1>
    </div>
  );
};

// Use Zustand for state management
interface PortfolioStore {
  portfolio: Portfolio | null;
  setPortfolio: (portfolio: Portfolio) => void;
}
```

### Backend Code Standards (Go)
```go
// Use proper package structure
package handlers

import (
    "encoding/json"
    "net/http"
    "github.com/gorilla/mux"
)

// Use proper error handling
func (h *Handler) GetPortfolio(w http.ResponseWriter, r *http.Request) {
    portfolio, err := h.portfolioService.GetPortfolio(r.Context())
    if err != nil {
        http.Error(w, "Failed to get portfolio", http.StatusInternalServerError)
        return
    }
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(portfolio)
}

// Use structured logging
log.Info("Processing portfolio request", 
    "user_id", userID,
    "request_id", requestID)
```

## 🧪 Testing Guidelines

### Frontend Testing
```bash
cd frontend

# Run tests
pnpm test

# Run tests in watch mode
pnpm test:watch

# Generate coverage report
pnpm test:coverage

# Type checking
pnpm run type-check
```

### Backend Testing
```bash
cd services/portfolio-service

# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run specific test
go test -run TestGetPortfolio
```

## 🐛 Debugging & Troubleshooting

### Common Issues

**Port already in use:**
```bash
# Find and kill process using port
lsof -ti:3000 | xargs kill -9
```

**Docker issues:**
```bash
# Reset Docker environment
pnpm run clean
docker system prune -f
pnpm run setup
```

**pnpm installation issues:**
```bash
# Clear pnpm cache
pnpm store prune
rm -rf node_modules pnpm-lock.yaml
pnpm install
```

**Git commit rejected:**
```bash
# Your commit message doesn't follow convention
# Use the template:
git commit  # Shows template
# Or format manually:
git commit -m "feat(service): description of change"
```

### Getting Help

1. **Check Documentation**: Most answers are in our docs
2. **Search Issues**: Someone might have had the same problem
3. **Ask in Team Chat**: Don't hesitate to ask questions
4. **Create Issue**: For bugs or unclear documentation
5. **Daily Standup**: Bring up blockers

## 📞 Communication Protocols

### Daily Standup (15 minutes)
**Time:** [Set by team] - same time every day
**Format:**
- What I completed yesterday
- What I'm working on today  
- Any blockers or help needed

### Code Reviews
- All PRs need at least one approval
- Review within 24 hours
- Be constructive and helpful
- Test the changes locally when possible

### Emergency Issues
- Critical bugs: Immediate team notification
- Blockers: Mention in team chat + daily standup
- Questions: Team chat first, then create issue if needed

## 🎯 Learning Resources

### Project-Specific
- [Go by Example](https://gobyexample.com/) - Go language basics
- [Next.js Documentation](https://nextjs.org/docs) - Next.js features
- [Docker Documentation](https://docs.docker.com/) - Container basics
- [PostgreSQL Tutorial](https://www.postgresql.org/docs/) - Database queries

### Team Skills
- [Git Workflow](https://www.atlassian.com/git/tutorials/comparing-workflows) - Git collaboration
- [Code Review Best Practices](https://smartbear.com/learn/code-review/best-practices-for-peer-code-review/) - Effective reviews
- [Microservices Patterns](https://microservices.io/patterns/) - Architecture patterns

## ✅ First Week Checklist

### Day 1: Setup & Familiarization
- [ ] Complete environment setup
- [ ] Run all services successfully
- [ ] Read team documentation
- [ ] Introduce yourself to the team
- [ ] Pick your first issue

### Day 2-3: First Contribution
- [ ] Complete first small task
- [ ] Submit first pull request
- [ ] Participate in code review process
- [ ] Attend daily standup

### Day 4-5: Full Productivity
- [ ] Working independently on assigned tasks
- [ ] Helping review others' code
- [ ] Contributing to team discussions
- [ ] Comfortable with workflow

## 🎉 Welcome to the Team!

You're now ready to contribute effectively to the Portfolio Management System. Remember:

- **Ask questions** - Better to ask than assume
- **Follow the process** - It helps everyone stay coordinated
- **Share knowledge** - Help others when you learn something new
- **Have fun** - This is a learning experience!

**Next Steps:**
1. Complete setup checklist
2. Read key documentation  
3. Pick your first issue
4. Join the team chat
5. Start coding!

---

**Need help?** Contact your team lead or ask in the team chat. We're here to help you succeed! 🚀