# Task Brief: Developer C - Add Holding Form & Portfolio Chart

**Objective**: Build the UI components for adding a new asset to the portfolio and visualizing the portfolio's composition with a chart.

**⚠️ DEPENDENCY**: Wait for Developer A to complete TypeScript interfaces before starting.

**📋 Read First**: 
- `docs/TEAM_COORDINATION_GUIDE.md` for shared resources and query keys
- `docs/UI_CONSISTENCY_GUIDE.md` for standardized UI components and patterns

---

## 1. Your Files

You are responsible for the following files. They have already been created for you:

-   **UI Component**: `frontend/src/components/AddHoldingForm.tsx`
-   **UI Component**: `frontend/src/components/PortfolioChart.tsx`

---

## 2. Step-by-Step Implementation Guide

### **Step 1: Implement the Add Holding Form**

-   **File**: `frontend/src/components/AddHoldingForm.tsx`
-   **Tooling**: Use `react-hook-form` for form state management and `zod` for validation. Use `react-query`'s `useMutation` for the API call.
-   **Action**:
    1.  **Define Schema**: Create a `zod` schema to validate the form fields: `symbol` (string, non-empty), `quantity` (number, positive), `average_cost` (number, positive).
    2.  **Build Form**: Use the `useForm` hook from `react-hook-form` with the `zodResolver`. Create input fields for `Symbol`, `Quantity`, and `Average Cost`.
    3.  **Handle Submission**:
        -   Create a `useMutation` hook that sends a `POST` request to `/api/v1/portfolio/holdings` with the form data.
        -   On success, the mutation should invalidate the `portfolioHoldings` and `portfolioSummary` queries so the rest of the dashboard updates automatically.
        -   In the form's `onSubmit` handler, call the `mutate` function from your `useMutation` hook.
    4.  **User Feedback**:
        -   Display validation errors next to each field.
        -   Disable the submit button while the mutation is in progress.
        -   Show a success/error toast notification after the submission attempt.
        -   Reset the form after a successful submission.

### **Step 2: Implement the Portfolio Chart Component**

-   **File**: `frontend/src/components/PortfolioChart.tsx`
-   **Tooling**: Use `recharts` for the chart and `react-query` to get the data.
-   **API Endpoint**: `GET /api/v1/portfolio/summary` (This is the same endpoint Developer A is using, but you will use a different part of the response).
-   **Action**:
    1.  **Fetch Data**: Use `useQuery` to fetch the portfolio summary data. You can use the same query key (`'portfolioSummary'`) as Developer A; `react-query` will automatically handle de-duplication.
    2.  **Process Data**: The data you need is in the `asset_allocation` array of the API response. You may need to format it to match what `recharts` expects (e.g., mapping `asset_type` to `name` and `total_value` to `value`).
    3.  **Render Chart**:
        -   Use the `ResponsiveContainer`, `PieChart`, `Pie`, `Cell`, and `Tooltip` components from `recharts`.
        -   Create a pie chart that visualizes the portfolio's composition by `asset_type`.
        -   Assign different colors to each slice of the pie.
        -   Configure the `Tooltip` to show the asset type and its percentage value on hover.
    4.  **Handle UI States**: Display a loading state while the data is being fetched and an error message if the API call fails.

---

## 3. Acceptance Criteria

-   [ ] The "Add Holding" form validates user input correctly.
-   [ ] The form successfully submits data to the API and adds a new holding.
-   [ ] The dashboard automatically updates after a new holding is added.
-   [ ] The pie chart correctly fetches and visualizes the asset allocation data.
-   [ ] The chart is responsive and includes interactive tooltips.
-   [ ] All operations provide clear feedback (loading, validation errors, success/error notifications).
