import { ChevronsLeft, ChevronsRight } from 'lucide-react'
import { NavLink } from 'react-router'
import { cn } from '../shared/lib/utils'
import { Button } from '../shared/ui/button'
import BrandMark from './BrandMark'
import { NAVIGATION } from './navigation'

type SidebarProps = {
  collapsed: boolean
  onToggle: () => void
}

export default function Sidebar({ collapsed, onToggle }: SidebarProps) {
  return (
    <aside
      className={cn(
        'sticky top-0 hidden h-dvh shrink-0 flex-col border-r border-sidebar-border bg-sidebar text-sidebar-foreground transition-[width] duration-200 md:flex',
        collapsed ? 'w-18' : 'w-64',
      )}
    >
      <div className={cn('flex h-16 items-center gap-3 px-4', collapsed && 'justify-center px-0')}>
        <BrandMark />
        {!collapsed && <span className="text-lg font-semibold tracking-tight">Market</span>}
      </div>

      <nav aria-label="Principal" className="flex-1 overflow-y-auto px-3 py-2">
        <ul className="flex flex-col gap-1">
          {NAVIGATION.map(({ to, label, icon: Icon, end }) => (
            <li key={to}>
              <NavLink
                to={to}
                end={end}
                aria-label={collapsed ? label : undefined}
                title={collapsed ? label : undefined}
                className={({ isActive }) =>
                  cn(
                    'flex h-10 items-center gap-3 rounded-lg px-3 text-sm font-medium transition-colors',
                    isActive
                      ? 'bg-sidebar-accent text-sidebar-accent-foreground'
                      : 'text-sidebar-foreground/70 hover:bg-sidebar-accent/60 hover:text-sidebar-foreground',
                    collapsed && 'justify-center px-0',
                  )
                }
              >
                <Icon className="size-5 shrink-0" aria-hidden />
                {!collapsed && <span>{label}</span>}
              </NavLink>
            </li>
          ))}
        </ul>
      </nav>

      <div className="border-t border-sidebar-border p-3">
        <Button
          type="button"
          variant="ghost"
          size="lg"
          className={cn(
            'w-full text-muted-foreground aria-expanded:bg-transparent aria-expanded:text-muted-foreground aria-expanded:hover:bg-muted aria-expanded:hover:text-foreground',
            collapsed ? 'justify-center' : 'justify-start',
          )}
          aria-expanded={!collapsed}
          aria-label={collapsed ? 'Expandir menú' : 'Contraer menú'}
          onClick={onToggle}
        >
          {collapsed ? <ChevronsRight aria-hidden /> : <ChevronsLeft aria-hidden />}
          {!collapsed && <span>Contraer</span>}
        </Button>
      </div>
    </aside>
  )
}
