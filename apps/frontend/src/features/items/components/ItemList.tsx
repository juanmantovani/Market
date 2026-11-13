import { Pencil, Trash2 } from 'lucide-react'
import { formatMoney } from '../../../shared/lib/money'
import { Button } from '../../../shared/ui/button'
import { ResponsiveList, type Column } from '../../../shared/ui/responsive-list'
import { unitShort } from '../../../shared/lib/quantity'
import type { Item } from '../types'

type ItemListProps = {
  items: readonly Item[]
  categoryName: (categoryId: string) => string
  isLoading: boolean
  onEdit: (item: Item) => void
  onDelete: (item: Item) => void
}

function price(item: Item) {
  return item.unit === 'unit' ? formatMoney(item.salePrice) : `${formatMoney(item.salePrice)} / ${unitShort(item.unit)}`
}

export function ItemList({ items, categoryName, isLoading, onEdit, onDelete }: ItemListProps) {
  const columns: Column<Item>[] = [
    { header: 'Código', cell: (item) => <span className="font-mono text-xs text-muted-foreground">{item.code}</span> },
    {
      header: 'Nombre',
      className: 'w-full max-w-0',
      cell: (item) => <span className="block truncate font-medium">{item.name}</span>,
    },
    { header: 'Categoría', cell: (item) => <span className="text-muted-foreground">{categoryName(item.categoryId)}</span> },
    { header: 'Precio', className: 'text-right tabular-nums', cell: (item) => price(item) },
    {
      header: 'Acciones',
      className: 'text-right',
      cell: (item) => (
        <div className="flex justify-end gap-1">
          <Button type="button" variant="ghost" size="icon" aria-label={`Editar ${item.name}`} title="Editar" onClick={() => onEdit(item)}>
            <Pencil aria-hidden />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="text-destructive hover:bg-destructive/10 hover:text-destructive"
            aria-label={`Eliminar ${item.name}`}
            title="Eliminar"
            onClick={() => onDelete(item)}
          >
            <Trash2 aria-hidden />
          </Button>
        </div>
      ),
    },
  ]

  return (
    <ResponsiveList
      label="Artículos"
      items={items}
      columns={columns}
      isLoading={isLoading}
      getKey={(item) => item.id}
      renderCard={(item) => (
        <div className="flex flex-col gap-3">
          <div className="flex items-start justify-between gap-3">
            <div className="flex min-w-0 flex-col gap-1">
              <span className="font-medium">{item.name}</span>
              <span className="text-sm text-muted-foreground">
                <span className="font-mono text-xs">{item.code}</span> · {categoryName(item.categoryId)}
              </span>
            </div>
            <span className="shrink-0 text-base font-semibold tabular-nums">{price(item)}</span>
          </div>
          <div className="flex gap-2">
            <Button type="button" variant="outline" className="h-11 flex-1" aria-label={`Editar ${item.name}`} onClick={() => onEdit(item)}>
              <Pencil aria-hidden />
              Editar
            </Button>
            <Button
              type="button"
              variant="outline"
              className="h-11 flex-1 text-destructive hover:bg-destructive/10 hover:text-destructive"
              aria-label={`Eliminar ${item.name}`}
              onClick={() => onDelete(item)}
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
