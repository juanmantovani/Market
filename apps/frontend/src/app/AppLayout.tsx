import { useState } from 'react'
import { Outlet } from 'react-router'
import BrandMark from './BrandMark'
import MobileNav from './MobileNav'
import ServerStatus from './ServerStatus'
import Sidebar from './Sidebar'
import { readSidebarCollapsed, storeSidebarCollapsed } from './sidebar-preference'
import ThemeToggle from './ThemeToggle'

export default function AppLayout() {
  const [collapsed, setCollapsed] = useState(readSidebarCollapsed)

  function toggleSidebar() {
    const next = !collapsed
    storeSidebarCollapsed(next)
    setCollapsed(next)
  }

  return (
    <div className="flex min-h-dvh bg-background text-foreground">
      <a
        href="#main"
        className="sr-only z-50 rounded-lg bg-primary px-4 py-2 text-primary-foreground focus:not-sr-only focus:fixed focus:top-3 focus:left-3"
      >
        Saltar al contenido
      </a>

      <Sidebar collapsed={collapsed} onToggle={toggleSidebar} />

      <div className="flex min-w-0 flex-1 flex-col">
        <header className="sticky top-0 z-10 flex h-16 items-center gap-3 border-b bg-background/80 px-4 backdrop-blur md:px-8">
          <span className="flex items-center gap-2 text-lg font-semibold tracking-tight md:hidden">
            <BrandMark />
            Market
          </span>
          <div className="ml-auto flex items-center gap-2">
            <ServerStatus />
            <ThemeToggle />
          </div>
        </header>

        <main id="main" className="flex-1 px-4 pt-6 pb-24 md:px-8 md:pb-10">
          <div className="mx-auto w-full max-w-6xl">
            <Outlet />
          </div>
        </main>
      </div>

      <MobileNav />
    </div>
  )
}
