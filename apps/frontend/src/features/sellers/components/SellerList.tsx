import { Pencil, Trash2 } from 'lucide-react'
import { formatPercent } from '../../../shared/lib/percent'
import { Button } from '../../../shared/ui/button'
import { ResponsiveList, type Column } from '../../../shared/ui/responsive-list'
import type { Seller } from '../types'

type SellerListProps = {
  sellers: readonly Seller[]
  isLoading: boolean
  onEdit: (seller: Seller) => void
  onDelete: (seller: Seller) => void
}

function CodeBadge({ code }: { code: string }) {
  return <span className="rounded-md bg-muted px-2 py-0.5 font-mono text-xs font-semibold">{code}</span>
}

function commission(seller: Seller) {
  return seller.commissionRate === null ? null : formatPercent(seller.commissionRate)
}

export function SellerList({ sellers, isLoading, onEdit, onDelete }: SellerListProps) {
  const columns: Column<Seller>[] = [
    { header: 'Código', cell: (seller) => <CodeBadge code={seller.code} /> },
    {
      header: 'Nombre',
      className: 'w-full max-w-0',
      cell: (seller) => <span className="block truncate font-medium">{seller.name}</span>,
    },
    {
      header: 'Comisión',
      className: 'text-right tabular-nums whitespace-nowrap',
      cell: (seller) => commission(seller) ?? <span className="text-muted-foreground">—</span>,
    },
    {
      header: 'Acciones',
      className: 'text-right',
      cell: (seller) => (
        <div className="flex justify-end gap-1">
          <Button type="button" variant="ghost" size="icon" aria-label={`Editar ${seller.name}`} title="Editar" onClick={() => onEdit(seller)}>
            <Pencil aria-hidden />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="text-destructive hover:bg-destructive/10 hover:text-destructive"
            aria-label={`Eliminar ${seller.name}`}
            title="Eliminar"
            onClick={() => onDelete(seller)}
          >
            <Trash2 aria-hidden />
          </Button>
        </div>
      ),
    },
  ]

  return (
    <ResponsiveList
      label="Vendedores"
      items={sellers}
      columns={columns}
      isLoading={isLoading}
      getKey={(seller) => seller.id}
      renderCard={(seller) => (
        <div className="flex flex-col gap-3">
          <div className="flex items-start justify-between gap-3">
            <div className="flex min-w-0 flex-col gap-1">
              <span className="font-medium">{seller.name}</span>
              <span className="text-sm text-muted-foreground">
                {commission(seller) ? `Comisión ${commission(seller)}` : 'Sin comisión'}
              </span>
            </div>
            <CodeBadge code={seller.code} />
          </div>
          <div className="flex gap-2">
            <Button type="button" variant="outline" className="h-11 flex-1" aria-label={`Editar ${seller.name}`} onClick={() => onEdit(seller)}>
              <Pencil aria-hidden />
              Editar
            </Button>
            <Button
              type="button"
              variant="outline"
              className="h-11 flex-1 text-destructive hover:bg-destructive/10 hover:text-destructive"
              aria-label={`Eliminar ${seller.name}`}
              onClick={() => onDelete(seller)}
            >
              <Trash2 aria-hidden />
              Eliminar
            </Button>
          </div>
        </div>
      )}
    />
  )
}
