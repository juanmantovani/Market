import { formatUnitQuantity, isLowStock } from '../../../shared/lib/quantity'
import { LowStockBadge } from '../../../shared/ui/low-stock-badge'
import { Skeleton } from '../../../shared/ui/skeleton'
import type { Item, StockQuantities } from '../types'

type ItemStockProps = {
  item: Item
  stock: StockQuantities
}

// The stock of an item, right aligned: its quantity in the item's unit,
// preceded by a "Stock bajo" mark (text, not only color) when it reached the
// minimum. In the table, below xl, the mark sits above the quantity so the
// column stays narrow.
export function ItemStock({ item, stock }: ItemStockProps) {
  const quantity = stock.quantities.get(item.id)

  if (quantity === undefined) {
    return (
      <div className="flex items-center justify-end">
        {stock.isLoading ? (
          <>
            <Skeleton className="h-4 w-14" />
            <span className="sr-only">Cargando stock</span>
          </>
        ) : (
          <>
            <span aria-hidden className="text-muted-foreground">
              —
            </span>
            <span className="sr-only">Stock no disponible</span>
          </>
        )}
      </div>
    )
  }

  return (
    <div className="flex items-center justify-end gap-2 whitespace-nowrap md:flex-col md:items-end md:gap-1 xl:flex-row xl:items-center xl:gap-2">
      {isLowStock(quantity, item.minStock) && <LowStockBadge />}
      <span className="font-medium tabular-nums">{formatUnitQuantity(quantity, item.unit)}</span>
    </div>
  )
}
