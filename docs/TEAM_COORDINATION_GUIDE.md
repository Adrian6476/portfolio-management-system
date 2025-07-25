# Team Coordination Guide - Critical Requirements

## 🔄 **Shared Resources Coordination**

### **React Query Keys (MUST USE EXACTLY)**
All developers must use these exact query keys to ensure proper cache coordination:

```typescript
// Developer A & C use this:
const PORTFOLIO_SUMMARY_KEY = 'portfolioSummary';

// Developer B uses these:
const PORTFOLIO_HOLDINGS_KEY = 'portfolioHoldings';
const CURRENT_PRICE_KEY = (symbol: string) => ['currentPrice', symbol];
```

### **Required TypeScript Interfaces (Developer A - PRIORITY 1)**
Developer A must complete these interfaces FIRST before others can proceed:

```typescript
// frontend/src/types/portfolio.ts
export interface Holding {
  id: string;
  symbol: string;
  name: string;
  asset_type: string;
  quantity: number;
  average_cost: number;
  purchase_date: string;
}

export interface PortfolioResponse {
  holdings: Holding[];
  total_holdings: number;
}

export interface Summary {
  total_holdings: number;
  total_cost: number;
  total_shares: number;
}

export interface AssetAllocation {
  asset_type: string;
  count: number;
  total_value: number;
  percentage: number;
}

export interface PortfolioSummaryResponse {
  summary: Summary;
  asset_allocation: AssetAllocation[];
  top_holdings: Holding[];
}

export interface CurrentPriceResponse {
  symbol: string;
  current_price: number;
  change: number;
  change_percent: number;
  high: number;
  low: number;
  open: number;
  previous_close: number;
  timestamp: number;
}
```

### **Environment Configuration (ALL DEVELOPERS)**
Create `.env.local` in frontend directory:
```
NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
```

## 📋 **Development Sequence (CRITICAL)**

### **Phase 1: Foundation (Developer A ONLY - Day 1 Morning)**
1. Complete `frontend/src/types/portfolio.ts` with ALL interfaces above
2. Configure `frontend/src/lib/api.ts` with proper base URL
3. Push to `feat/dashboard-layout` branch
4. **NOTIFY TEAM**: "TypeScript interfaces ready"

### **Phase 2: Parallel Development (Day 1 Afternoon - Day 2)**
1. Developer B & C pull Developer A's interfaces
2. All developers work on their components using the shared types
3. Use exact query keys specified above

### **Phase 3: Integration (Day 3)**
1. Merge all sub-branches into `feat/portfolio-dashboard`
2. Test cache invalidation flow
3. Verify all components update properly

## 🔒 **Cache Invalidation Protocol**

### **Developer B - After Mutations:**
```typescript
// After successful update/delete
queryClient.invalidateQueries([PORTFOLIO_HOLDINGS_KEY]);
queryClient.invalidateQueries([PORTFOLIO_SUMMARY_KEY]);
```

### **Developer C - After Form Submission:**
```typescript
// After successful add holding
queryClient.invalidateQueries([PORTFOLIO_HOLDINGS_KEY]);
queryClient.invalidateQueries([PORTFOLIO_SUMMARY_KEY]);
```

## ⚠️ **Conflict Prevention Rules**

1. **NO custom query keys** - use only the standardized ones above
2. **NO duplicate TypeScript interfaces** - import from `types/portfolio.ts`
3. **NO hardcoded API URLs** - use the apiClient from `lib/api.ts`
4. **NO CSS conflicts** - use only Tailwind utility classes
5. **NO environment variables** - except the one specified above

## 📞 **Communication Checkpoints**

- **9:00 AM Daily**: 15-minute standup with progress update
- **Developer A completes interfaces**: Immediate Slack notification
- **Any API issues**: Immediate team notification
- **Before merging**: Code review by at least one other developer

## 🚨 **Emergency Contacts**

If any developer gets blocked on shared resources, immediately contact the team lead for resolution.
