import { useId, type ReactNode } from 'react'
import { ShoppingCart } from 'lucide-react'
import { formatMoney } from '../../../shared/lib/money'
import { cn } from '../../../shared/lib/utils'
import { Button } from '../../../shared/ui/button'
import type { CartLine } from '../types'
import { CartLineRow } from './CartLineRow'

type CartPanelProps = {
  lines: CartLine[]
  total: number
  onQuantityChange: (itemId: string, quantity: number) => void
  onRemove: (itemId: string) => void
  onClear: () => void
  // Extra controls for the header, such as the close button of the sheet.
  headerAction?: ReactNode
  // What goes under the total.
  children?: ReactNode
  className?: string
}

export function CartPanel({ lines, total, onQuantityChange, onRemove, onClear, headerAction, children, className }: CartPanelProps) {
  const titleId = useId()

  return (
    <section aria-labelledby={titleId} className={cn('flex min-h-0 flex-col rounded-2xl border bg-card text-card-foreground', className)}>
      <header className="flex items-center justify-between gap-2 border-b px-4 py-2">
        <h2 id={titleId} className="py-2 text-base font-semibold">
          Venta actual
        </h2>
        <div className="flex items-center gap-1">
          {lines.length > 0 && (
            <Button type="button" variant="ghost" size="lg" className="h-11 text-muted-foreground md:h-9" onClick={onClear}>
              Vaciar
            </Button>
          )}
          {headerAction}
        </div>
      </header>

      {lines.length === 0 ? (
        <div className="flex flex-col items-center gap-2 px-6 py-10 text-center">
          <ShoppingCart className="size-8 text-muted-foreground" aria-hidden />
          <p className="text-sm font-medium">Todavía no hay artículos en la venta</p>
          <p className="text-sm text-muted-foreground">Buscar o escanear un artículo para empezar.</p>
        </div>
      ) : (
        <ul aria-label="Artículos de la venta" className="min-h-0 flex-1 divide-y overflow-y-auto px-4">
          {lines.map((line) => (
            <CartLineRow key={line.itemId} line={line} onQuantityChange={onQuantityChange} onRemove={onRemove} />
          ))}
        </ul>
      )}

      <footer className="flex flex-col gap-3 border-t px-4 py-4">
        <dl>
          <div className="flex items-baseline justify-between gap-4">
            <dt className="text-base font-medium">Total</dt>
            <dd className="text-2xl font-semibold tabular-nums">{formatMoney(total)}</dd>
          </div>
        </dl>
        {children}
      </footer>
    </section>
  )
}
