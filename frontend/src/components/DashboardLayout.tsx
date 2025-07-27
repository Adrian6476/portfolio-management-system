'use client'

import React from 'react';
import { 
  BarChart3, 
  Table,
  Plus,
  PieChart,
  Settings,
  RefreshCw,
  DollarSign,
  TrendingUp
} from 'lucide-react';

interface DashboardLayoutProps {
  children: React.ReactNode;
  activeView: string;
  setActiveView: (view: string) => void;
}

export default function DashboardLayout({ children, activeView, setActiveView }: DashboardLayoutProps) {
  const getCurrentDate = () => {
    return new Date().toLocaleDateString('en-US', { 
      weekday: 'long', 
      year: 'numeric', 
      month: 'long', 
      day: 'numeric' 
    });
  };

  const navigationItems = [
    { id: 'overview', icon: BarChart3, label: 'Portfolio Overview', description: 'Main dashboard view' },
    { id: 'manage', icon: Table, label: 'Manage Holdings', description: 'View and edit holdings' },
    { id: 'add', icon: Plus, label: 'Add Holdings', description: 'Add new investments' },
    { id: 'analytics', icon: PieChart, label: 'Analytics', description: 'Charts and performance' },
  ];

  const getViewTitle = () => {
    const item = navigationItems.find(item => item.id === activeView);
    return item ? item.label : 'Portfolio Overview';
  };

  return (
    <div className="flex min-h-screen bg-gray-50">
      {/* Sidebar */}
      <aside className="bank-sidebar">
        <div className="w-12 h-12 rounded-xl bg-primary flex items-center justify-center mb-8">
          <TrendingUp className="w-6 h-6 text-white" />
        </div>
        
        <nav className="space-y-2">
          {navigationItems.map((item) => (
            <button
              key={item.id}
              onClick={() => setActiveView(item.id)}
              className={`bank-nav-item group relative ${
                activeView === item.id ? 'bg-primary' : ''
              }`}
              title={item.label}
            >
              <item.icon className={`w-5 h-5 ${
                activeView === item.id ? 'text-white' : 'text-gray-400'
              }`} />
              
              {/* Tooltip */}
              <div className="absolute left-16 top-1/2 transform -translate-y-1/2 bg-gray-900 text-white px-3 py-2 rounded-lg text-sm whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity duration-200 pointer-events-none z-50">
                <div className="font-medium">{item.label}</div>
                <div className="text-xs text-gray-300">{item.description}</div>
                <div className="absolute left-0 top-1/2 transform -translate-y-1/2 -translate-x-1 w-2 h-2 bg-gray-900 rotate-45"></div>
              </div>
            </button>
          ))}
        </nav>

        <div className="mt-auto">
          <button className="bank-nav-item group relative" title="Settings">
            <Settings className="w-5 h-5 text-gray-400" />
            
            {/* Tooltip */}
            <div className="absolute left-16 top-1/2 transform -translate-y-1/2 bg-gray-900 text-white px-3 py-2 rounded-lg text-sm whitespace-nowrap opacity-0 group-hover:opacity-100 transition-opacity duration-200 pointer-events-none z-50">
              <div className="font-medium">Settings</div>
              <div className="text-xs text-gray-300">App preferences</div>
              <div className="absolute left-0 top-1/2 transform -translate-y-1/2 -translate-x-1 w-2 h-2 bg-gray-900 rotate-45"></div>
            </div>
          </button>
        </div>
      </aside>

      {/* Main Content */}
      <main className="flex-1 flex flex-col">
        {/* Header */}
        <header className="bank-header">
          <div className="flex items-center space-x-6">
            <div>
              <h1 className="text-2xl font-bold text-gray-900">{getViewTitle()}</h1>
              <span className="text-sm text-gray-500">{getCurrentDate()}</span>
            </div>
          </div>
          
          <div className="flex items-center space-x-4">
            {/* Portfolio Value Summary */}
            <div className="hidden md:flex items-center space-x-6 px-4 py-2 bg-gray-50 rounded-lg">
              <div className="text-center">
                <div className="text-sm font-medium text-gray-900">Total Value</div>
                <div className="text-lg font-bold text-green-600">$--,---</div>
              </div>
              <div className="text-center">
                <div className="text-sm font-medium text-gray-900">Today</div>
                <div className="text-lg font-bold text-gray-600">--%</div>
              </div>
            </div>
            
            {/* Action Buttons */}
            <div className="flex items-center space-x-3">
              <button 
                className="p-2 rounded-lg border border-gray-200 hover:bg-gray-50 transition-colors"
                title="Refresh Data"
              >
                <RefreshCw className="w-4 h-4 text-gray-600" />
              </button>
              
              <button 
                className="p-2 rounded-lg border border-gray-200 hover:bg-gray-50 transition-colors"
                title="Portfolio Value"
              >
                <DollarSign className="w-4 h-4 text-gray-600" />
              </button>
              
              <button className="px-4 py-2 bg-slate-900 text-white rounded-lg hover:bg-slate-800 transition-colors">
                Export
              </button>
            </div>
          </div>
        </header>

        {/* Content Area */}
        <div className="bank-main-content">
          {children}
        </div>
      </main>
    </div>
  );
}
