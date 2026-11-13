import type { KeyboardEvent } from 'react'
import { SearchX } from 'lucide-react'
import { formatMoney } from '../../../shared/lib/money'
import { unitShort } from '../../../shared/lib/quantity'
import { cn } from '../../../shared/lib/utils'
import { Button } from '../../../shared/ui/button'
import { EmptyState } from '../../../shared/ui/empty-state'
import { Skeleton } from '../../../shared/ui/skeleton'
import type { CatalogItem } from '../types'

type CatalogGridProps = {
  items: CatalogItem[] | undefined
  isPending: boolean
  isError: boolean
  onRetry: () => void
  onPick: (item: CatalogItem) => void
  // Escape goes back to the search field.
  onEscape: () => void
}

const NEXT_KEYS = new Set(['ArrowRight', 'ArrowDown'])
const PREVIOUS_KEYS = new Set(['ArrowLeft', 'ArrowUp'])

export function CatalogGrid({ items, isPending, isError, onRetry, onPick, onEscape }: CatalogGridProps) {
  if (isError && !items) {
    return (
      <EmptyState
        icon={SearchX}
        title="No se pudieron cargar los artículos"
        description="Revisar la conexión e intentar de nuevo."
        action={
          <Button type="button" variant="outline" size="lg" className="h-11 md:h-9" onClick={onRetry}>
            Reintentar
          </Button>
        }
      />
    )
  }

  if (isPending && !items) {
    return (
      <ul aria-label="Artículos" aria-busy="true" className="grid grid-cols-2 gap-2 sm:grid-cols-3 xl:grid-cols-4">
        {Array.from({ length: 8 }, (_, index) => (
          <li key={index}>
            <Skeleton className="h-24 rounded-xl" />
          </li>
        ))}
      </ul>
    )
  }

  if (!items || items.length === 0) {
    return (
      <EmptyState
        icon={SearchX}
        title="No hay artículos que coincidan"
        description="Probar con otra búsqueda o con otra categoría."
      />
    )
  }

  // Arrow keys walk the cards, so the catalog works without a mouse.
  function moveFocus(event: KeyboardEvent<HTMLUListElement>) {
    if (event.key === 'Escape') {
      event.preventDefault()
      onEscape()
      return
    }
    const step = NEXT_KEYS.has(event.key) ? 1 : PREVIOUS_KEYS.has(event.key) ? -1 : 0
    if (step === 0) return
    const buttons = [...event.currentTarget.querySelectorAll('button')]
    const current = buttons.indexOf(document.activeElement as HTMLButtonElement)
    const next = buttons[current + step]
    if (next) {
      event.preventDefault()
      next.focus()
    } else if (step < 0) {
      event.preventDefault()
      onEscape()
    }
  }

  return (
    <ul aria-label="Artículos" className="grid grid-cols-2 gap-2 sm:grid-cols-3 xl:grid-cols-4" onKeyDown={moveFocus}>
      {items.map((item) => (
        <li key={item.id}>
          <button
            type="button"
            onClick={() => onPick(item)}
            className={cn(
              'flex h-full min-h-24 w-full flex-col items-start gap-1 rounded-xl border bg-card p-3 text-left shadow-xs transition-colors',
              'hover:border-primary/40 hover:bg-accent/40 focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/50 focus-visible:outline-none active:translate-y-px',
            )}
          >
            <span className="line-clamp-2 text-sm font-medium">{item.name}</span>
            <span className="text-xs text-muted-foreground">{item.code}</span>
            <span className="mt-auto text-sm font-semibold tabular-nums">
              {formatMoney(item.salePrice)}
              {item.unit !== 'unit' && <span className="font-normal text-muted-foreground"> / {unitShort(item.unit)}</span>}
            </span>
          </button>
        </li>
      ))}
    </ul>
  )
}
