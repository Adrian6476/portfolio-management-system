# UI Consistency Guide - Portfolio Dashboard

## 🎨 **Design System Overview**

This guide ensures all developers create a cohesive, professional-looking dashboard by using standardized components and patterns.

---

## 📦 **Shared UI Components**

**Location**: `frontend/src/components/ui/index.tsx`

### **✅ MANDATORY: Use These Components**

All developers MUST use these shared components instead of creating custom ones:

```typescript
// For components in the same directory level as 'ui' folder:
import {
  Card,
  Button,
  Input,
  LoadingSpinner,
  ErrorMessage,
  UI_CONSTANTS
} from './ui';

// For components in app directory:
import {
  Card,
  Button,
  Input,
  LoadingSpinner,
  ErrorMessage,
  UI_CONSTANTS
} from '../../components/ui';
```

### **Component Usage Examples**

#### **Cards (Developer A - Summary, Developer C - Chart)**
```typescript
// ✅ Correct
<Card title="Portfolio Summary" className="mb-4">
  <div>Summary content here</div>
</Card>

// ❌ Wrong - Don't create custom cards
<div className="bg-white p-4 rounded shadow">
```

#### **Buttons (All Developers)**
```typescript
// ✅ Correct
<Button variant="primary" onClick={handleSubmit}>
  Add Holding
</Button>

<Button variant="danger" size="sm" onClick={handleDelete}>
  Delete
</Button>

// ❌ Wrong - Don't create custom buttons
<button className="bg-blue-500 p-2 rounded">
```

#### **Form Inputs (Developer C - Form)**
```typescript
// ✅ Correct
<Input
  label="Symbol"
  placeholder="e.g., AAPL"
  value={symbol}
  onChange={(e) => setSymbol(e.target.value)}
  error={errors.symbol}
  required
/>

// ❌ Wrong - Don't create custom inputs
<input className="border p-2" />
```

#### **Loading States (All Developers)**
```typescript
// ✅ Correct
{isLoading && <LoadingSpinner size="lg" />}

// ❌ Wrong - Don't create custom spinners
{isLoading && <div className="animate-spin">Loading...</div>}
```

---

## 🎯 **Layout Patterns**

### **Developer A - Dashboard Layout**
```typescript
// ✅ Standard dashboard grid
<div className="container mx-auto px-4 py-8">
  <h1 className={UI_CONSTANTS.typography.heading1}>Portfolio Dashboard</h1>
  
  <div className="grid grid-cols-1 lg:grid-cols-3 gap-6 mt-8">
    {/* Summary card - spans 2 columns on large screens */}
    <div className="lg:col-span-2">
      <PortfolioSummary />
    </div>
    
    {/* Chart - spans 1 column */}
    <div className="lg:col-span-1">
      <PortfolioChart />
    </div>
    
    {/* Table - spans full width */}
    <div className="lg:col-span-3">
      <HoldingsTable />
    </div>
    
    {/* Form - spans 1 column */}
    <div className="lg:col-span-1">
      <AddHoldingForm />
    </div>
  </div>
</div>
```

### **Developer B - Table Layout**
```typescript
// ✅ Standard responsive table pattern
<Card title="Holdings">
  {isLoading ? (
    <div className="flex justify-center py-8">
      <LoadingSpinner size="lg" />
    </div>
  ) : error ? (
    <ErrorMessage message="Failed to load holdings" />
  ) : (
    <div className="overflow-x-auto">
      <table className="min-w-full divide-y divide-gray-200">
        {/* Table content */}
      </table>
    </div>
  )}
</Card>
```

### **Developer C - Form Layout**
```typescript
// ✅ Standard form pattern
<Card title="Add New Holding">
  <form onSubmit={handleSubmit} className="space-y-4">
    <Input
      label="Symbol"
      value={symbol}
      onChange={(e) => setSymbol(e.target.value)}
      error={errors.symbol}
      required
    />
    
    <div className="flex justify-end space-x-2 pt-4">
      <Button variant="secondary" onClick={onCancel}>
        Cancel
      </Button>
      <Button type="submit" disabled={isSubmitting}>
        {isSubmitting ? 'Adding...' : 'Add Holding'}
      </Button>
    </div>
  </form>
</Card>
```

---

## 🎨 **Color & Styling Standards**

### **Color Usage**
- **Primary Actions**: `variant="primary"` (Blue)
- **Destructive Actions**: `variant="danger"` (Red) 
- **Secondary Actions**: `variant="secondary"` (Gray)
- **Success States**: `variant="success"` (Green)

### **Typography Hierarchy**
```typescript
// ✅ Use these exact classes
<h1 className={UI_CONSTANTS.typography.heading1}>Page Title</h1>
<h2 className={UI_CONSTANTS.typography.heading2}>Section Title</h2>
<h3 className={UI_CONSTANTS.typography.heading3}>Card Title</h3>
<p className={UI_CONSTANTS.typography.body}>Body text</p>
<span className={UI_CONSTANTS.typography.caption}>Small text</span>
```

### **Spacing Standards**
```typescript
// ✅ Use these standard spacing classes
className={UI_CONSTANTS.spacing.section}    // mb-8 (between major sections)
className={UI_CONSTANTS.spacing.element}    // mb-4 (between elements)
className={UI_CONSTANTS.spacing.card}       // p-6 (card padding)
```

---

## 📱 **Responsive Design Rules**

### **Mobile-First Approach**
All components must work on mobile first, then enhance for larger screens:

```typescript
// ✅ Correct responsive pattern
<div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
  {/* Mobile: 1 column, Tablet: 2 columns, Desktop: 3 columns */}
</div>

// ✅ Hide/show based on screen size
<div className="hidden md:block">Desktop only content</div>
<div className="md:hidden">Mobile only content</div>
```

### **Table Responsiveness (Developer B)**
```typescript
// ✅ Standard responsive table wrapper
<div className="overflow-x-auto">
  <table className="min-w-full">
    {/* Table content */}
  </table>
</div>
```

---

## 🔄 **State Management Patterns**

### **Loading States**
```typescript
// ✅ Standard loading pattern
{isLoading ? (
  <div className="flex justify-center items-center py-8">
    <LoadingSpinner size="lg" />
  </div>
) : (
  <div>Content here</div>
)}
```

### **Error States**
```typescript
// ✅ Standard error pattern
{error && (
  <ErrorMessage message={error.message || 'Something went wrong'} />
)}
```

### **Empty States**
```typescript
// ✅ Standard empty state pattern
{data.length === 0 && (
  <div className="text-center py-8">
    <p className={UI_CONSTANTS.typography.body}>No holdings found</p>
    <Button variant="primary" onClick={onAddFirst} className="mt-4">
      Add Your First Holding
    </Button>
  </div>
)}
```

---

## 🚨 **Common Mistakes to Avoid**

### **❌ DON'T DO THESE:**

1. **Custom CSS Classes**
   ```typescript
   // ❌ Wrong
   <div className="my-custom-card">
   
   // ✅ Correct
   <Card>
   ```

2. **Inconsistent Colors**
   ```typescript
   // ❌ Wrong
   <button className="bg-blue-500">
   
   // ✅ Correct
   <Button variant="primary">
   ```

3. **Hardcoded Spacing**
   ```typescript
   // ❌ Wrong
   <div className="mb-5 p-7">
   
   // ✅ Correct
   <div className={`${UI_CONSTANTS.spacing.element} ${UI_CONSTANTS.spacing.card}`}>
   ```

4. **Custom Loading Spinners**
   ```typescript
   // ❌ Wrong
   <div className="animate-spin h-6 w-6 border-2 border-blue-600">
   
   // ✅ Correct
   <LoadingSpinner size="md" />
   ```

---

## ✅ **Quality Checklist**

Before submitting your code, verify:

- [ ] Used shared components from `ui/index.tsx`
- [ ] Applied consistent spacing using `UI_CONSTANTS`
- [ ] Used standard color variants for buttons
- [ ] Implemented proper loading states
- [ ] Added error handling with `ErrorMessage`
- [ ] Made design responsive (mobile-first)
- [ ] Used typography hierarchy correctly
- [ ] No custom CSS classes or inline styles

---

## 🔍 **Review Process**

1. **Self-Review**: Check against this guide before submitting PR
2. **Peer Review**: Another developer verifies UI consistency
3. **Integration Review**: Final check when merging branches

---

**Remember**: Consistency is more important than creativity. A cohesive, professional-looking dashboard builds user trust and makes the product feel polished.
