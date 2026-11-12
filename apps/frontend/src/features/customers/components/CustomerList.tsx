import { Pencil, Trash2 } from 'lucide-react'
import { Button } from '../../../shared/ui/button'
import { ContactLines } from '../../../shared/ui/contact-lines'
import { ResponsiveList, type Column } from '../../../shared/ui/responsive-list'
import { formatDocument, taxConditionLabel, type Customer } from '../types'

type CustomerListProps = {
  customers: readonly Customer[]
  isLoading: boolean
  onEdit: (customer: Customer) => void
  onDelete: (customer: Customer) => void
}

export function CustomerList({ customers, isLoading, onEdit, onDelete }: CustomerListProps) {
  const columns: Column<Customer>[] = [
    {
      header: 'Nombre',
      className: 'w-full max-w-0',
      cell: (customer) => <span className="block truncate font-medium">{customer.name}</span>,
    },
    {
      header: 'Documento',
      className: 'whitespace-nowrap tabular-nums',
      cell: (customer) => formatDocument(customer) ?? <span className="text-muted-foreground">—</span>,
    },
    {
      header: 'Condición IVA',
      className: 'whitespace-nowrap',
      cell: (customer) => <span className="text-muted-foreground">{taxConditionLabel(customer.taxCondition)}</span>,
    },
    {
      header: 'Contacto',
      className: 'max-w-64',
      cell: (customer) =>
        customer.phone || customer.email ? (
          <ContactLines phone={customer.phone} email={customer.email} className="text-muted-foreground" />
        ) : (
          <span className="text-muted-foreground">—</span>
        ),
    },
    {
      header: 'Acciones',
      className: 'text-right',
      cell: (customer) => (
        <div className="flex justify-end gap-1">
          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label={`Editar ${customer.name}`}
            title="Editar"
            onClick={() => onEdit(customer)}
          >
            <Pencil aria-hidden />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="text-destructive hover:bg-destructive/10 hover:text-destructive"
            aria-label={`Eliminar ${customer.name}`}
            title="Eliminar"
            onClick={() => onDelete(customer)}
          >
            <Trash2 aria-hidden />
          </Button>
        </div>
      ),
    },
  ]

  return (
    <ResponsiveList
      label="Clientes"
      items={customers}
      columns={columns}
      isLoading={isLoading}
      getKey={(customer) => customer.id}
      renderCard={(customer) => (
        <div className="flex flex-col gap-3">
          <div className="flex min-w-0 flex-col gap-1">
            <span className="font-medium">{customer.name}</span>
            <span className="text-sm text-muted-foreground">
              {[formatDocument(customer), taxConditionLabel(customer.taxCondition)].filter(Boolean).join(' · ')}
            </span>
            {(customer.phone || customer.email) && (
              <ContactLines phone={customer.phone} email={customer.email} className="text-sm text-muted-foreground" />
            )}
          </div>
          <div className="flex gap-2">
            <Button
              type="button"
              variant="outline"
              className="h-11 flex-1"
              aria-label={`Editar ${customer.name}`}
              onClick={() => onEdit(customer)}
            >
              <Pencil aria-hidden />
              Editar
            </Button>
            <Button
              type="button"
              variant="outline"
              className="h-11 flex-1 text-destructive hover:bg-destructive/10 hover:text-destructive"
              aria-label={`Eliminar ${customer.name}`}
              onClick={() => onDelete(customer)}
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
