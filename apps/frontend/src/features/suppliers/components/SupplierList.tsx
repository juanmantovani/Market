import { Pencil, Trash2 } from 'lucide-react'
import { formatCuit } from '../../../shared/lib/document'
import { Button } from '../../../shared/ui/button'
import { ContactLines } from '../../../shared/ui/contact-lines'
import { ResponsiveList, type Column } from '../../../shared/ui/responsive-list'
import type { Supplier } from '../types'

type SupplierListProps = {
  suppliers: readonly Supplier[]
  isLoading: boolean
  onEdit: (supplier: Supplier) => void
  onDelete: (supplier: Supplier) => void
}

function Dash() {
  return <span className="text-muted-foreground">—</span>
}

export function SupplierList({ suppliers, isLoading, onEdit, onDelete }: SupplierListProps) {
  const columns: Column<Supplier>[] = [
    {
      header: 'Razón social',
      className: 'w-full max-w-0',
      cell: (supplier) => <span className="block truncate font-medium">{supplier.businessName}</span>,
    },
    {
      header: 'CUIT',
      className: 'whitespace-nowrap tabular-nums',
      cell: (supplier) => (supplier.cuit ? formatCuit(supplier.cuit) : <Dash />),
    },
    {
      header: 'Contacto',
      className: 'whitespace-nowrap',
      cell: (supplier) => supplier.contactName ?? <Dash />,
    },
    {
      header: 'Teléfono y email',
      className: 'max-w-64',
      cell: (supplier) =>
        supplier.phone || supplier.email ? (
          <ContactLines phone={supplier.phone} email={supplier.email} className="text-muted-foreground" />
        ) : (
          <Dash />
        ),
    },
    {
      header: 'Acciones',
      className: 'text-right',
      cell: (supplier) => (
        <div className="flex justify-end gap-1">
          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label={`Editar ${supplier.businessName}`}
            title="Editar"
            onClick={() => onEdit(supplier)}
          >
            <Pencil aria-hidden />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="text-destructive hover:bg-destructive/10 hover:text-destructive"
            aria-label={`Eliminar ${supplier.businessName}`}
            title="Eliminar"
            onClick={() => onDelete(supplier)}
          >
            <Trash2 aria-hidden />
          </Button>
        </div>
      ),
    },
  ]

  return (
    <ResponsiveList
      label="Proveedores"
      items={suppliers}
      columns={columns}
      isLoading={isLoading}
      getKey={(supplier) => supplier.id}
      renderCard={(supplier) => {
        const summary = [supplier.cuit && `CUIT ${formatCuit(supplier.cuit)}`, supplier.contactName].filter(Boolean).join(' · ')
        return (
          <div className="flex flex-col gap-3">
            <div className="flex min-w-0 flex-col gap-1">
              <span className="font-medium">{supplier.businessName}</span>
              {summary && <span className="text-sm text-muted-foreground">{summary}</span>}
              {(supplier.phone || supplier.email) && (
                <ContactLines phone={supplier.phone} email={supplier.email} className="text-sm text-muted-foreground" />
              )}
            </div>
            <div className="flex gap-2">
              <Button
                type="button"
                variant="outline"
                className="h-11 flex-1"
                aria-label={`Editar ${supplier.businessName}`}
                onClick={() => onEdit(supplier)}
              >
                <Pencil aria-hidden />
                Editar
              </Button>
              <Button
                type="button"
                variant="outline"
                className="h-11 flex-1 text-destructive hover:bg-destructive/10 hover:text-destructive"
                aria-label={`Eliminar ${supplier.businessName}`}
                onClick={() => onDelete(supplier)}
              >
                <Trash2 aria-hidden />
                Eliminar
              </Button>
            </div>
          </div>
        )
      }}
    />
  )
}
