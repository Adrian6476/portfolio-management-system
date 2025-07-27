// This file is now obsolete - dashboard functionality moved to root (/)
// Redirect to root if someone navigates to /dashboard

'use client'

import { useEffect } from 'react'
import { useRouter } from 'next/navigation'

export default function DashboardRedirect() {
  const router = useRouter()
  
  useEffect(() => {
    router.replace('/')
  }, [router])

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-50">
      <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary"></div>
    </div>
  )
}
