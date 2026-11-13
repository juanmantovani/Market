import { History, SlidersHorizontal } from 'lucide-react'
import { formatUnitQuantity } from '../../../shared/lib/quantity'
import { Button } from '../../../shared/ui/button'
import { LowStockBadge } from '../../../shared/ui/low-stock-badge'
import { ResponsiveList, type Column } from '../../../shared/ui/responsive-list'
import type { StockLevel } from '../types'

type StockListProps = {
  levels: readonly StockLevel[]
  categoryName: (categoryId: string) => string
  isLoading: boolean
  onAdjust: (level: StockLevel) => void
  onShowMovements: (level: StockLevel) => void
}

function minimum(level: StockLevel) {
  return level.minStock > 0 ? formatUnitQuantity(level.minStock, level.unit) : '—'
}

export function StockList({ levels, categoryName, isLoading, onAdjust, onShowMovements }: StockListProps) {
  const columns: Column<StockLevel>[] = [
    { header: 'Código', cell: (level) => <span className="font-mono text-xs text-muted-foreground">{level.code}</span> },
    {
      header: 'Artículo',
      className: 'w-full max-w-0',
      cell: (level) => <span className="block truncate font-medium">{level.name}</span>,
    },
    { header: 'Categoría', cell: (level) => <span className="text-muted-foreground">{categoryName(level.categoryId)}</span> },
    {
      header: 'Stock',
      className: 'text-right tabular-nums',
      cell: (level) => (
        <span className="inline-flex items-center justify-end gap-2 whitespace-nowrap">
          {level.lowStock && <LowStockBadge />}
          <span className="font-medium">{formatUnitQuantity(level.quantity, level.unit)}</span>
        </span>
      ),
    },
    { header: 'Mínimo', className: 'text-right tabular-nums whitespace-nowrap text-muted-foreground', cell: minimum },
    {
      header: 'Acciones',
      className: 'text-right',
      cell: (level) => (
        <div className="flex justify-end gap-1">
          <Button type="button" variant="ghost" size="icon" aria-label={`Ajustar ${level.name}`} title="Ajustar" onClick={() => onAdjust(level)}>
            <SlidersHorizontal aria-hidden />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label={`Movimientos de ${level.name}`}
            title="Movimientos"
            onClick={() => onShowMovements(level)}
          >
            <History aria-hidden />
          </Button>
        </div>
      ),
    },
  ]

  return (
    <ResponsiveList
      label="Stock"
      items={levels}
      columns={columns}
      isLoading={isLoading}
      getKey={(level) => level.itemId}
      renderCard={(level) => (
        <div className="flex flex-col gap-3">
          <div className="flex items-start justify-between gap-3">
            <div className="flex min-w-0 flex-col gap-1">
              <span className="font-medium">{level.name}</span>
              <span className="text-sm text-muted-foreground">
                <span className="font-mono text-xs">{level.code}</span> · {categoryName(level.categoryId)}
              </span>
            </div>
            <div className="flex shrink-0 flex-col items-end gap-1">
              <span className="text-base font-semibold tabular-nums">{formatUnitQuantity(level.quantity, level.unit)}</span>
              <span className="text-xs text-muted-foreground">Mínimo: {minimum(level)}</span>
            </div>
          </div>
          {level.lowStock && (
            <div>
              <LowStockBadge />
            </div>
          )}
          <div className="flex gap-2">
            <Button type="button" variant="outline" className="h-11 flex-1" aria-label={`Ajustar ${level.name}`} onClick={() => onAdjust(level)}>
              <SlidersHorizontal aria-hidden />
              Ajustar
            </Button>
            <Button
              type="button"
              variant="outline"
              className="h-11 flex-1"
              aria-label={`Movimientos de ${level.name}`}
              onClick={() => onShowMovements(level)}
            >
              <History aria-hidden />
              Movimientos
            </Button>
          </div>
        </div>
      )}
    />
  )
}
