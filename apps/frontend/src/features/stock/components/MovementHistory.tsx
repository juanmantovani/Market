import { useState } from 'react'
import { CircleAlert, History } from 'lucide-react'
import { messageFromError } from '../../../shared/api/client'
import { formatDateTime } from '../../../shared/lib/date'
import { formatMoney } from '../../../shared/lib/money'
import { formatUnitQuantity, unitShort } from '../../../shared/lib/quantity'
import { cn } from '../../../shared/lib/utils'
import { Button } from '../../../shared/ui/button'
import { EmptyState } from '../../../shared/ui/empty-state'
import { Pagination } from '../../../shared/ui/pagination'
import { Skeleton } from '../../../shared/ui/skeleton'
import { useStockMovements } from '../hooks/use-stock'
import { MOVEMENT_TYPE_LABELS, type StockLevel, type StockMovement } from '../types'

const PAGE_SIZE = 20

type MovementHistoryProps = {
  level: StockLevel
  supplierName: (supplierId: string) => string
}

function details(movement: StockMovement, level: StockLevel, supplierName: (supplierId: string) => string): string[] {
  const parts: string[] = []
  if (movement.supplierId) parts.push(supplierName(movement.supplierId))
  if (movement.reference) parts.push(movement.reference)
  if (movement.unitCost !== null) {
    const unit = level.unit === 'unit' ? '' : ` / ${unitShort(level.unit)}`
    parts.push(`Costo ${formatMoney(movement.unitCost)}${unit}`)
  }
  if (movement.reason) parts.push(`Motivo: ${movement.reason}`)
  return parts
}

export function MovementHistory({ level, supplierName }: MovementHistoryProps) {
  const [page, setPage] = useState(1)
  const movements = useStockMovements({ itemId: level.itemId, page, limit: PAGE_SIZE })

  if (movements.isPending) {
    return (
      <div role="status" aria-label="Cargando movimientos" className="flex flex-col gap-3">
        {Array.from({ length: 3 }, (_, index) => (
          <Skeleton key={index} className="h-16 w-full rounded-xl" />
        ))}
      </div>
    )
  }

  if (movements.isError) {
    return (
      <EmptyState
        icon={CircleAlert}
        title="No se pudieron cargar los movimientos"
        description={messageFromError(movements.error)}
        action={
          <Button type="button" variant="outline" size="lg" onClick={() => void movements.refetch()}>
            Reintentar
          </Button>
        }
      />
    )
  }

  if (movements.data.items.length === 0) {
    return (
      <EmptyState
        icon={History}
        title="Todavía no hay movimientos"
        description="El stock cambia con cada ingreso de mercadería, ajuste o venta."
      />
    )
  }

  return (
    <div className="flex flex-col gap-4">
      <ol aria-label={`Movimientos de ${level.name}`} className="flex flex-col divide-y rounded-xl border">
        {movements.data.items.map((movement) => {
          const extra = details(movement, level, supplierName)
          const sign = movement.quantity > 0 ? '+' : ''
          return (
            <li key={movement.id} className="flex flex-col gap-1 px-4 py-3">
              <div className="flex items-baseline justify-between gap-3">
                <span className="font-medium">{MOVEMENT_TYPE_LABELS[movement.type]}</span>
                <span className={cn('font-semibold tabular-nums', movement.quantity < 0 && 'text-destructive')}>
                  {sign}
                  {formatUnitQuantity(movement.quantity, level.unit)}
                </span>
              </div>
              <div className="flex flex-wrap items-baseline justify-between gap-x-3 gap-y-1 text-sm text-muted-foreground">
                <time dateTime={movement.createdAt}>{formatDateTime(movement.createdAt)}</time>
                <span className="tabular-nums">Queda: {formatUnitQuantity(movement.balanceAfter, level.unit)}</span>
              </div>
              {extra.length > 0 && <p className="text-sm text-muted-foreground">{extra.join(' · ')}</p>}
            </li>
          )
        })}
      </ol>
      {movements.data.total > movements.data.limit && (
        <Pagination page={movements.data.page} limit={movements.data.limit} total={movements.data.total} onPageChange={setPage} />
      )}
    </div>
  )
}
