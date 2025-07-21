'use client'

import { useEffect, useState } from 'react'

export default function HomePage() {
  const [mounted, setMounted] = useState(false)

  useEffect(() => {
    setMounted(true)
  }, [])

  if (!mounted) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary"></div>
      </div>
    )
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <header className="text-center mb-12">
        <h1 className="text-4xl md:text-6xl font-bold text-gradient mb-4">
          Portfolio Management System
        </h1>
        <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
          Advanced microservices-based portfolio management platform with real-time analytics,
          automated insights, and comprehensive risk management.
        </p>
      </header>

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8 mb-12">
        <div className="metric-card card-hover">
          <h3 className="text-lg font-semibold mb-2">Real-time Portfolio Tracking</h3>
          <p className="text-muted-foreground">
            Monitor your investments with live market data and instant portfolio updates.
          </p>
        </div>

        <div className="metric-card card-hover">
          <h3 className="text-lg font-semibold mb-2">Advanced Analytics</h3>
          <p className="text-muted-foreground">
            Comprehensive performance analysis with risk metrics and predictive insights.
          </p>
        </div>

        <div className="metric-card card-hover">
          <h3 className="text-lg font-semibold mb-2">Microservices Architecture</h3>
          <p className="text-muted-foreground">
            Scalable, resilient system built with modern microservices and event-driven design.
          </p>
        </div>

        <div className="metric-card card-hover">
          <h3 className="text-lg font-semibold mb-2">Automated Rebalancing</h3>
          <p className="text-muted-foreground">
            Intelligent portfolio optimization with automated rebalancing suggestions.
          </p>
        </div>

        <div className="metric-card card-hover">
          <h3 className="text-lg font-semibold mb-2">AI-Powered Insights</h3>
          <p className="text-muted-foreground">
            Machine learning-driven analysis for personalized investment recommendations.
          </p>
        </div>

        <div className="metric-card card-hover">
          <h3 className="text-lg font-semibold mb-2">What-if Analysis</h3>
          <p className="text-muted-foreground">
            Simulate portfolio changes and see projected impacts before making decisions.
          </p>
        </div>
      </div>

      <div className="text-center">
        <div className="inline-flex items-center justify-center space-x-4 p-4 bg-muted rounded-lg">
          <div className="flex items-center space-x-2">
            <div className="w-3 h-3 bg-success rounded-full animate-pulse"></div>
            <span className="text-sm font-medium">System Status: Initializing</span>
          </div>
          <div className="text-sm text-muted-foreground">
            Services: API Gateway, Portfolio Service, Market Data, Analytics
          </div>
        </div>
      </div>
    </div>
  )
}