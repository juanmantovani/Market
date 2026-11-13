import { queryKeys } from '../../../shared/api/query-keys'
import { documentText, sellerByCode } from '../lib/parties'
import type { CustomerOption, SearchCustomers, SearchSellers, SellerOption } from '../types'
import { PartyPicker } from './PartyPicker'

type PickerProps<T, S> = {
  open: boolean
  onClose: () => void
  search: S
  current: T | null
  onPick: (option: T | null) => void
}

// Consumidor Final comes in the results like any customer; choosing it
// goes back to the default.
export function CustomerPicker({ open, onClose, search, current, onPick }: PickerProps<CustomerOption, SearchCustomers>) {
  return (
    <PartyPicker<CustomerOption>
      open={open}
      onClose={onClose}
      title="Elegir cliente"
      searchLabel="Buscar cliente"
      hint="Por nombre, documento, email o teléfono. Enter elige el primero."
      emptyText="No hay clientes que coincidan. Probar con otra búsqueda."
      queryKey={queryKeys.customers.list}
      search={search}
      getKey={(customer) => customer.id}
      isCurrent={(customer) => (current ? customer.id === current.id : customer.system)}
      renderOption={(customer) => (
        <span className="flex min-w-0 flex-col">
          <span className="truncate font-medium">{customer.name}</span>
          <span className="text-xs text-muted-foreground">
            {customer.system ? 'Cliente de las ventas sin cliente' : (documentText(customer) ?? 'Sin documento')}
          </span>
        </span>
      )}
      onPick={onPick}
    />
  )
}

export function SellerPicker({ open, onClose, search, current, onPick }: PickerProps<SellerOption, SearchSellers>) {
  return (
    <PartyPicker<SellerOption>
      open={open}
      onClose={onClose}
      title="Elegir vendedor"
      searchLabel="Buscar vendedor"
      hint="Por código o nombre. Enter elige el del código tipeado."
      emptyText="No hay vendedores que coincidan. Probar con otra búsqueda."
      queryKey={queryKeys.sellers.list}
      search={search}
      getKey={(seller) => seller.id}
      isCurrent={(seller) => seller.id === current?.id}
      renderOption={(seller) => (
        <span className="flex min-w-0 items-center gap-2">
          <span className="rounded bg-muted px-1.5 py-0.5 text-xs font-semibold tabular-nums">{seller.code}</span>
          <span className="truncate font-medium">{seller.name}</span>
        </span>
      )}
      pickOnEnter={sellerByCode}
      onPick={onPick}
      extra={{
        label: 'Sin vendedor',
        description: 'La venta no queda a nombre de nadie',
        isCurrent: current === null,
        onPick: () => onPick(null),
      }}
    />
  )
}
