import { cn } from '../../../shared/lib/utils'

export function CategoryStatusBadge({ active }: { active: boolean }) {
  return (
    <span
      className={cn(
        'inline-flex h-6 items-center gap-1.5 rounded-full px-2.5 text-xs font-medium',
        active ? 'bg-success/10 text-success' : 'bg-muted text-muted-foreground',
      )}
    >
      <span className={cn('size-1.5 rounded-full', active ? 'bg-success' : 'bg-muted-foreground')} aria-hidden />
      {active ? 'Activa' : 'Inactiva'}
    </span>
  )
}
