// Shared UI Components Library
// All developers must use these components for consistency

import React from 'react';

// Design System Constants
export const UI_CONSTANTS = {
  // Colors (using Tailwind classes)
  colors: {
    primary: 'bg-blue-600 hover:bg-blue-700',
    secondary: 'bg-gray-600 hover:bg-gray-700',
    success: 'bg-green-600 hover:bg-green-700',
    danger: 'bg-red-600 hover:bg-red-700',
    warning: 'bg-yellow-600 hover:bg-yellow-700',
  },
  
  // Spacing
  spacing: {
    card: 'p-6',
    section: 'mb-8',
    element: 'mb-4',
  },
  
  // Border radius
  radius: {
    card: 'rounded-lg',
    button: 'rounded-md',
    input: 'rounded-md',
  },
  
  // Shadows
  shadows: {
    card: 'shadow-md hover:shadow-lg',
    modal: 'shadow-xl',
  },
  
  // Typography
  typography: {
    heading1: 'text-3xl font-bold text-gray-900',
    heading2: 'text-2xl font-semibold text-gray-800',
    heading3: 'text-xl font-medium text-gray-700',
    body: 'text-gray-600',
    caption: 'text-sm text-gray-500',
  },
};

// Standard Card Component
interface CardProps {
  children: React.ReactNode;
  className?: string;
  title?: string;
}

export function Card({ children, className = '', title }: CardProps) {
  return (
    <div className={`bg-white ${UI_CONSTANTS.radius.card} ${UI_CONSTANTS.shadows.card} ${UI_CONSTANTS.spacing.card} ${className}`}>
      {title && (
        <h3 className={`${UI_CONSTANTS.typography.heading3} ${UI_CONSTANTS.spacing.element}`}>
          {title}
        </h3>
      )}
      {children}
    </div>
  );
}

// Standard Button Component
interface ButtonProps {
  children: React.ReactNode;
  onClick?: () => void;
  variant?: 'primary' | 'secondary' | 'success' | 'danger' | 'warning';
  size?: 'sm' | 'md' | 'lg';
  disabled?: boolean;
  type?: 'button' | 'submit' | 'reset';
  className?: string;
}

export function Button({
  children,
  onClick,
  variant = 'primary',
  size = 'md',
  disabled = false,
  type = 'button',
  className = '',
}: ButtonProps) {
  const sizeClasses = {
    sm: 'px-3 py-1.5 text-sm',
    md: 'px-4 py-2 text-base',
    lg: 'px-6 py-3 text-lg',
  };

  return (
    <button
      type={type}
      onClick={onClick}
      disabled={disabled}
      className={`
        ${UI_CONSTANTS.colors[variant]}
        ${sizeClasses[size]}
        ${UI_CONSTANTS.radius.button}
        text-white font-medium
        transition-colors duration-200
        disabled:opacity-50 disabled:cursor-not-allowed
        focus:outline-none focus:ring-2 focus:ring-blue-500 focus:ring-offset-2
        ${className}
      `}
    >
      {children}
    </button>
  );
}

// Standard Input Component
interface InputProps {
  label?: string;
  placeholder?: string;
  value?: string | number;
  onChange?: (e: React.ChangeEvent<HTMLInputElement>) => void;
  type?: 'text' | 'number' | 'email' | 'password';
  error?: string;
  required?: boolean;
  className?: string;
}

export function Input({
  label,
  placeholder,
  value,
  onChange,
  type = 'text',
  error,
  required = false,
  className = '',
}: InputProps) {
  return (
    <div className={`${UI_CONSTANTS.spacing.element} ${className}`}>
      {label && (
        <label className={`block ${UI_CONSTANTS.typography.caption} font-medium mb-1`}>
          {label}
          {required && <span className="text-red-500 ml-1">*</span>}
        </label>
      )}
      <input
        type={type}
        placeholder={placeholder}
        value={value}
        onChange={onChange}
        className={`
          w-full px-3 py-2 border border-gray-300 ${UI_CONSTANTS.radius.input}
          focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent
          ${error ? 'border-red-500' : 'border-gray-300'}
          ${UI_CONSTANTS.typography.body}
        `}
      />
      {error && (
        <p className="mt-1 text-sm text-red-500">{error}</p>
      )}
    </div>
  );
}

// Standard Loading Spinner
interface LoadingSpinnerProps {
  size?: 'sm' | 'md' | 'lg';
  className?: string;
}

export function LoadingSpinner({ size = 'md', className = '' }: LoadingSpinnerProps) {
  const sizeClasses = {
    sm: 'h-4 w-4',
    md: 'h-8 w-8',
    lg: 'h-12 w-12',
  };

  return (
    <div className={`animate-spin rounded-full border-b-2 border-blue-600 ${sizeClasses[size]} ${className}`} />
  );
}

// Standard Error Message
interface ErrorMessageProps {
  message: string;
  className?: string;
}

export function ErrorMessage({ message, className = '' }: ErrorMessageProps) {
  return (
    <div className={`bg-red-50 border border-red-200 text-red-700 px-4 py-3 ${UI_CONSTANTS.radius.card} ${className}`}>
      <p className="text-sm">{message}</p>
    </div>
  );
}
