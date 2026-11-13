import { Eye } from 'lucide-react'
import { formatDateTime } from '../../../shared/lib/date'
import { formatMoney } from '../../../shared/lib/money'
import { Button } from '../../../shared/ui/button'
import { ResponsiveList, type Column } from '../../../shared/ui/responsive-list'
import type { SaleRecord } from '../types'
import { SaleStatusBadge } from './SaleStatusBadge'

type SaleListProps = {
  sales: readonly SaleRecord[]
  isLoading: boolean
  onOpen: (sale: SaleRecord) => void
}

function sellerName(sale: SaleRecord) {
  return sale.seller?.name ?? 'Sin vendedor'
}

export function SaleList({ sales, isLoading, onOpen }: SaleListProps) {
  const columns: Column<SaleRecord>[] = [
    { header: 'Número', className: 'whitespace-nowrap font-medium tabular-nums', cell: (sale) => `N.º ${sale.number}` },
    { header: 'Fecha', className: 'whitespace-nowrap tabular-nums', cell: (sale) => formatDateTime(sale.createdAt) },
    {
      header: 'Cliente',
      className: 'w-full max-w-0',
      cell: (sale) => <span className="block truncate">{sale.customer.name}</span>,
    },
    { header: 'Vendedor', className: 'whitespace-nowrap text-muted-foreground', cell: sellerName },
    { header: 'Total', className: 'whitespace-nowrap text-right font-medium tabular-nums', cell: (sale) => formatMoney(sale.total) },
    { header: 'Estado', className: 'whitespace-nowrap', cell: (sale) => <SaleStatusBadge status={sale.status} /> },
    {
      header: 'Acciones',
      className: 'text-right',
      cell: (sale) => (
        <Button
          type="button"
          variant="ghost"
          size="icon"
          aria-label={`Ver la venta N.º ${sale.number}`}
          title="Ver detalle"
          onClick={() => onOpen(sale)}
        >
          <Eye aria-hidden />
        </Button>
      ),
    },
  ]

  return (
    <ResponsiveList
      label="Ventas"
      items={sales}
      columns={columns}
      isLoading={isLoading}
      getKey={(sale) => sale.id}
      renderCard={(sale) => (
        <div className="flex flex-col gap-3">
          <div className="flex items-start justify-between gap-3">
            <div className="flex min-w-0 flex-col gap-1">
              <span className="font-medium">N.º {sale.number}</span>
              <span className="text-sm text-muted-foreground">{formatDateTime(sale.createdAt)}</span>
              <span className="truncate text-sm">{sale.customer.name}</span>
              <span className="text-sm text-muted-foreground">{sellerName(sale)}</span>
            </div>
            <div className="flex shrink-0 flex-col items-end gap-2">
              <span className="font-semibold tabular-nums">{formatMoney(sale.total)}</span>
              <SaleStatusBadge status={sale.status} />
            </div>
          </div>
          <Button type="button" variant="outline" className="h-11" aria-label={`Ver la venta N.º ${sale.number}`} onClick={() => onOpen(sale)}>
            <Eye aria-hidden />
            Ver detalle
          </Button>
        </div>
      )}
    />
  )
}
