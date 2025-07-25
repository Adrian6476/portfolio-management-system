# Task Brief: Developer A - Dashboard Foundation & Portfolio Summary

**Objective**: Build the foundational layout for the portfolio dashboard and implement the summary component that displays key portfolio metrics.

**⚠️ CRITICAL**: You are the foundation developer. Other team members depend on your TypeScript interfaces. Complete Step 1 FIRST and notify the team immediately.

**📋 Read First**: 
- `docs/TEAM_COORDINATION_GUIDE.md` for shared resources and query keys
- `docs/UI_CONSISTENCY_GUIDE.md` for standardized UI components and patterns

---

## 1. Your Files

You are responsible for the following files. They have already been created for you:

-   **Page Layout**: `frontend/src/app/dashboard/page.tsx`
-   **UI Component**: `frontend/src/components/PortfolioSummary.tsx`
-   **Shared API Client**: `frontend/src/lib/api.ts`
-   **Shared Types**: `frontend/src/types/portfolio.ts`

---

## 2. Step-by-Step Implementation Guide

### **Step 1: Define Shared TypeScript Interfaces**

-   **File**: `frontend/src/types/portfolio.ts`
-   **Action**: Define the TypeScript interfaces for the data you will receive from the API. The backend provides the following structure for the summary:
    ```json
    {
      "summary": {
        "total_holdings": 10,
        "total_cost": 50000.00,
        "total_shares": 120.5
      },
      "asset_allocation": [
        {
          "asset_type": "STOCK",
          "count": 5,
          "total_value": 30000.00,
          "percentage": 60.0
        }
      ]
    }
    ```
-   **Implementation**: Create `PortfolioSummaryResponse`, `Summary`, and `AssetAllocation` interfaces based on the JSON above.

### **Step 2: Configure the API Client**

-   **File**: `frontend/src/lib/api.ts`
-   **Action**: The file already contains a basic Axios instance. Ensure the `baseURL` points to the correct API gateway address (`http://localhost:8080/api/v1`). This client will be used by the entire team.

### **Step 3: Implement the Portfolio Summary Component**

-   **File**: `frontend/src/components/PortfolioSummary.tsx`
-   **Tooling**: Use `react-query`'s `useQuery` hook to fetch data.
-   **API Endpoint**: `GET /api/v1/portfolio/summary`
-   **Action**:
    1.  Import `useQuery` from `react-query` and the `apiClient` from `../lib/api`.
    2.  Create a function to fetch the data, e.g., `const fetchSummary = async () => { const { data } = await apiClient.get('/portfolio/summary'); return data; }`.
    3.  Call `useQuery('portfolioSummary', fetchSummary)` to get the `data`, `isLoading`, and `isError` states.
    4.  Display a loading spinner while `isLoading` is true.
    5.  Display an error message if `isError` is true.
    6.  Once data is available, display the key metrics (`total_holdings`, `total_cost`, `total_shares`) in a visually appealing way using TailwindCSS cards.

### **Step 4: Build the Dashboard Page Layout**

-   **File**: `frontend/src/app/dashboard/page.tsx`
-   **Action**:
    1.  Import the `PortfolioSummary` component.
    2.  Create a responsive layout using CSS Grid or Flexbox (e.g., `grid grid-cols-1 md:grid-cols-3 gap-4`).
    3.  Place the `PortfolioSummary` component in the layout.
    4.  Leave placeholders for the other components (`HoldingsTable`, `AddHoldingForm`, `PortfolioChart`) that your teammates will build.

---

## 3. Acceptance Criteria

-   [ ] The dashboard page renders without errors.
-   [ ] The `PortfolioSummary` component correctly fetches and displays data from the API.
-   [ ] The component shows a clear loading state while fetching.
-   [ ] The component shows a clear error state if the API call fails.
-   [ ] The layout is responsive and looks good on both desktop and mobile screens.
-   [ ] All code is strongly typed using the interfaces defined in `portfolio.ts`.
