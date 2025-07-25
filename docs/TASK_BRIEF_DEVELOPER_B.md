# Task Brief: Developer B - Holdings Data Table

**Objective**: Build a comprehensive and interactive data table to display, update, and delete portfolio holdings, including real-time price data.

**⚠️ DEPENDENCY**: Wait for Developer A to complete TypeScript interfaces before starting Step 2.

**📋 Read First**: 
- `docs/TEAM_COORDINATION_GUIDE.md` for shared resources and query keys
- `docs/UI_CONSISTENCY_GUIDE.md` for standardized UI components and patterns

---

## 1. Your Files

You are responsible for the following files. They have already been created for you:

-   **UI Component**: `frontend/src/components/HoldingsTable.tsx`
-   **Data Fetching Hooks**: `frontend/src/hooks/usePortfolio.ts`

---

## 2. Step-by-Step Implementation Guide

### **Step 1: Create Data Fetching Hooks**

-   **File**: `frontend/src/hooks/usePortfolio.ts`
-   **Tooling**: Use `react-query` for all data fetching and mutations.
-   **Action**:
    1.  **Fetch Holdings (`usePortfolioHoldings`)**:
        -   Create a `useQuery` hook to fetch data from `GET /api/v1/portfolio`.
        -   This hook will provide the list of all holdings to your table component.
    2.  **Update Holding (`useUpdateHolding`)**:
        -   Create a `useMutation` hook that sends a `PUT` request to `PUT /api/v1/portfolio/holdings/:id`.
        -   On success, it should invalidate the `portfolioHoldings` query to refetch the data.
    3.  **Delete Holding (`useDeleteHolding`)**:
        -   Create a `useMutation` hook that sends a `DELETE` request to `DELETE /api/v1/portfolio/holdings/:id`.
        -   On success, it should also invalidate the `portfolioHoldings` query.

### **Step 2: Build the Holdings Table Component**

-   **File**: `frontend/src/components/HoldingsTable.tsx`
-   **Action**:
    1.  **Data Display**:
        -   Use your `usePortfolioHoldings` hook to get the holdings data, loading, and error states.
        -   Render a table displaying columns like `Symbol`, `Name`, `Quantity`, `Average Cost`, and `Current Value`.
        -   Use TailwindCSS for styling to make the table responsive.
    2.  **Real-time Price Integration**:
        -   For each row, use a separate `useQuery` hook (with `staleTime` set to 1 minute) to fetch the current price from `GET /api/v1/market/prices/:symbol`.
        -   Calculate and display the `Current Value` (Quantity \* Current Price).
    3.  **Edit Functionality**:
        -   Add an "Edit" button to each row.
        -   Clicking it should open a modal (or an inline form) allowing the user to change the `Quantity` and `Average Cost`.
        -   On submit, call the `mutate` function from your `useUpdateHolding` hook.
    4.  **Delete Functionality**:
        -   Add a "Delete" button to each row.
        -   Show a confirmation dialog before proceeding.
        -   On confirmation, call the `mutate` function from your `useDeleteHolding` hook.

### **Step 3: Handle UI States**

-   **Action**:
    -   Display a single, table-level loading skeleton while the main holdings list is loading.
    -   For real-time prices, you can show a smaller, inline spinner.
    -   Disable form buttons while mutations are in progress to prevent double-submission.
    -   Use a library like `react-hot-toast` to show success or error notifications after an update or delete operation.

---

## 3. Acceptance Criteria

-   [ ] The table correctly fetches and displays all portfolio holdings.
-   [ ] Each holding's current price is fetched and displayed.
-   [ ] Users can successfully update a holding's quantity and average cost.
-   [ ] Users can successfully delete a holding after a confirmation prompt.
-   [ ] The table is responsive and usable on mobile devices.
-   [ ] All operations provide clear feedback to the user (loading, success, error).
