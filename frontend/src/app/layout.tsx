import React from 'react'
import type { Metadata, Viewport } from 'next'
import { Inter } from 'next/font/google'
import './globals.css'
import { Providers } from '@/components/providers'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'Portfolio Management System',
  description: 'Advanced microservices-based portfolio management platform',
  keywords: ['portfolio', 'investment', 'finance', 'trading', 'analytics'],
  authors: [{ name: 'Portfolio Management Team' }],
}

export const viewport: Viewport = {
  width: 'device-width',
  initialScale: 1,
  themeColor: [
    { media: '(prefers-color-scheme: light)', color: 'white' },
    { media: '(prefers-color-scheme: dark)', color: 'black' },
  ],
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en" suppressHydrationWarning>
      <body className={inter.className}>
        <div id="root" className="min-h-screen bg-background">
          <Providers>
            {children}
          </Providers>
        </div>
      </body>
    </html>
  )
}