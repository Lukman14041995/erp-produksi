import { useEffect, useState } from 'react'
import { Outlet, useLocation } from 'react-router-dom'
import { ErrorBoundary } from '@/components/ErrorBoundary'
import { Sidebar } from './Sidebar'
import { Header } from './Header'

export function AppShell() {
  const { pathname } = useLocation()
  const [mobileNavOpen, setMobileNavOpen] = useState(false)

  // Close the mobile drawer automatically on route change (covers
  // navigation triggered from outside the sidebar, e.g. the command menu).
  useEffect(() => {
    setMobileNavOpen(false)
  }, [pathname])

  return (
    <div className="flex h-screen overflow-hidden bg-[var(--color-background)]">
      <Sidebar mobileOpen={mobileNavOpen} onCloseMobile={() => setMobileNavOpen(false)} />
      <div className="flex flex-1 flex-col overflow-hidden">
        <Header onOpenSidebar={() => setMobileNavOpen(true)} />
        <main className="flex-1 overflow-y-auto p-4 md:p-6">
          <ErrorBoundary key={pathname}>
            <Outlet />
          </ErrorBoundary>
        </main>
      </div>
    </div>
  )
}
