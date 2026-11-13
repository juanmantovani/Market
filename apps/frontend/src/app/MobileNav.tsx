import { useState } from 'react'
import { Popover } from '@base-ui/react/popover'
import { Ellipsis } from 'lucide-react'
import { matchPath, NavLink, useLocation } from 'react-router'
import { cn } from '../shared/lib/utils'
import { NAVIGATION } from './navigation'

const barItemClass = 'flex min-h-14 w-full flex-col items-center justify-center gap-1 px-2 text-xs font-medium transition-colors'

function barItemColor(isActive: boolean) {
  return isActive ? 'text-primary' : 'text-muted-foreground hover:text-foreground'
}

export default function MobileNav() {
  const [moreOpen, setMoreOpen] = useState(false)
  const { pathname } = useLocation()
  const barItems = NAVIGATION.filter((item) => item.inMobileBar)
  const moreItems = NAVIGATION.filter((item) => !item.inMobileBar)
  const isInMore = moreItems.some((item) => matchPath({ path: item.to, end: Boolean(item.end) }, pathname))

  return (
    <nav
      aria-label="Navegación móvil"
      className="fixed inset-x-0 bottom-0 z-20 border-t bg-background/95 pb-[env(safe-area-inset-bottom)] backdrop-blur md:hidden"
    >
      <ul className="mx-auto grid max-w-md grid-cols-4">
        {barItems.map(({ to, label, icon: Icon, end }) => (
          <li key={to}>
            <NavLink to={to} end={end} className={({ isActive }) => cn(barItemClass, barItemColor(isActive))}>
              <Icon className="size-5" aria-hidden />
              <span>{label}</span>
            </NavLink>
          </li>
        ))}
        <li>
          <Popover.Root open={moreOpen} onOpenChange={setMoreOpen}>
            <Popover.Trigger className={cn(barItemClass, barItemColor(isInMore || moreOpen))}>
              <Ellipsis className="size-5" aria-hidden />
              <span>Más</span>
            </Popover.Trigger>
            <Popover.Portal>
              <Popover.Positioner side="top" align="end" sideOffset={8} className="z-30">
                <Popover.Popup className="min-w-52 rounded-xl border bg-popover p-1.5 text-popover-foreground shadow-lg outline-none">
                  <Popover.Title className="sr-only">Más secciones</Popover.Title>
                  <ul className="flex flex-col gap-0.5">
                    {moreItems.map(({ to, label, icon: Icon, end }) => (
                      <li key={to}>
                        <NavLink
                          to={to}
                          end={end}
                          onClick={() => setMoreOpen(false)}
                          className={({ isActive }) =>
                            cn(
                              'flex min-h-11 items-center gap-3 rounded-lg px-3 text-sm font-medium transition-colors',
                              isActive ? 'bg-accent text-accent-foreground' : 'hover:bg-muted',
                            )
                          }
                        >
                          <Icon className="size-5 text-muted-foreground" aria-hidden />
                          {label}
                        </NavLink>
                      </li>
                    ))}
                  </ul>
                </Popover.Popup>
              </Popover.Positioner>
            </Popover.Portal>
          </Popover.Root>
        </li>
      </ul>
    </nav>
  )
}
