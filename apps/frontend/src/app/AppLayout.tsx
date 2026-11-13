import { Keyboard } from 'lucide-react'
import { Suspense, useEffect, useRef, useState } from 'react'
import { Outlet } from 'react-router'
import { Button } from '../shared/ui/button'
import BrandMark from './BrandMark'
import MobileNav from './MobileNav'
import PageSkeleton from './PageSkeleton'
import ShortcutsDialog from './ShortcutsDialog'
import { isHelpKey } from './shortcuts'
import Sidebar from './Sidebar'
import { readSidebarCollapsed, storeSidebarCollapsed } from './sidebar-preference'
import ThemeToggle from './ThemeToggle'
import { useFocusHeadingOnNavigation } from './use-focus-heading-on-navigation'
import { usePageTitle } from './use-page-title'

export default function AppLayout() {
  const [collapsed, setCollapsed] = useState(readSidebarCollapsed)
  const [helpOpen, setHelpOpen] = useState(false)
  const mainRef = useRef<HTMLElement>(null)
  usePageTitle()
  useFocusHeadingOnNavigation(mainRef)

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (!isHelpKey(event)) return
      event.preventDefault()
      setHelpOpen(true)
    }

    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [])

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
            <Button
              type="button"
              variant="ghost"
              size="icon-lg"
              className="hidden size-9 text-muted-foreground md:inline-flex"
              aria-label="Atajos de teclado"
              title="Atajos de teclado (?)"
              onClick={() => setHelpOpen(true)}
            >
              <Keyboard className="size-5" aria-hidden />
            </Button>
            <ThemeToggle />
          </div>
        </header>

        <main ref={mainRef} id="main" className="flex-1 px-4 pt-6 pb-24 md:px-8 md:pb-10">
          <div className="mx-auto w-full max-w-6xl">
            <Suspense fallback={<PageSkeleton />}>
              <Outlet />
            </Suspense>
          </div>
        </main>
      </div>

      <MobileNav />
      <ShortcutsDialog open={helpOpen} onOpenChange={setHelpOpen} />
    </div>
  )
}
