import { useState } from 'react'
import { CircleAlert, Plus, SearchX, Users } from 'lucide-react'
import { messageFromError } from '../../../shared/api/client'
import { useDebouncedValue } from '../../../shared/hooks/use-debounced-value'
import { Button } from '../../../shared/ui/button'
import { ConfirmDialog } from '../../../shared/ui/confirm-dialog'
import { EmptyState } from '../../../shared/ui/empty-state'
import { FormDialog } from '../../../shared/ui/form-dialog'
import { PageHeader } from '../../../shared/ui/page-header'
import { Pagination } from '../../../shared/ui/pagination'
import { SearchField } from '../../../shared/ui/search-field'
import { useToast } from '../../../shared/ui/use-toast'
import { CustomerForm } from '../components/CustomerForm'
import { CustomerList } from '../components/CustomerList'
import { useCustomers, useDeleteCustomer, useSaveCustomer } from '../hooks/use-customers'
import { toFormValues, type Customer, type CustomerRequest } from '../types'

const PAGE_SIZE = 25

// `open` is kept apart from what the overlay shows, so its content stays in
// place while the closing animation runs instead of flashing empty.
type FormState = { open: boolean; customer: Customer | null; formKey: number }
type ConfirmState = { open: boolean; customer: Customer | null }

export default function CustomersPage() {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [form, setForm] = useState<FormState>({ open: false, customer: null, formKey: 0 })
  const [confirm, setConfirm] = useState<ConfirmState>({ open: false, customer: null })
  const debouncedSearch = useDebouncedValue(search)

  const customers = useCustomers({ search: debouncedSearch, page, limit: PAGE_SIZE })
  const saveCustomer = useSaveCustomer()
  const deleteCustomer = useDeleteCustomer()
  const toast = useToast()

  const rows = customers.data?.items ?? []
  const isSearching = debouncedSearch.trim() !== ''
  const isEditing = form.customer !== null

  function openForm(customer: Customer | null) {
    setForm((current) => ({ open: true, customer, formKey: current.formKey + 1 }))
  }

  function closeForm() {
    setForm((current) => ({ ...current, open: false }))
  }

  function closeConfirm() {
    setConfirm((current) => ({ ...current, open: false }))
  }

  async function save(request: CustomerRequest) {
    const id = form.customer?.id
    const saved = await saveCustomer.mutateAsync({ id, request })
    closeForm()
    toast.success(id ? 'Cliente actualizado' : 'Cliente creado', `«${saved.name}» ya está guardado.`)
  }

  async function remove(customer: Customer) {
    try {
      await deleteCustomer.mutateAsync(customer.id)
      if (rows.length === 1 && page > 1) {
        setPage(page - 1)
      }
      toast.success('Cliente eliminado', `«${customer.name}» ya no existe.`)
    } catch (error) {
      toast.error('No se pudo eliminar el cliente', messageFromError(error))
    } finally {
      closeConfirm()
    }
  }

  function renderContent() {
    if (customers.isError) {
      return (
        <EmptyState
          icon={CircleAlert}
          title="No se pudieron cargar los clientes"
          description={messageFromError(customers.error)}
          action={
            <Button type="button" variant="outline" size="lg" onClick={() => void customers.refetch()}>
              Reintentar
            </Button>
          }
        />
      )
    }

    if (!customers.isPending && rows.length === 0) {
      return isSearching ? (
        <EmptyState
          icon={SearchX}
          title="No se encontraron clientes"
          description="Probar con otra parte del nombre, el documento, el email o el teléfono."
        />
      ) : (
        <EmptyState
          icon={Users}
          title="Todavía no hay clientes"
          description="Cargar a quienes compran seguido o necesitan factura, con sus datos."
          action={
            <Button type="button" size="lg" onClick={() => openForm(null)}>
              <Plus aria-hidden />
              Crear el primero
            </Button>
          }
        />
      )
    }

    return (
      <>
        <CustomerList
          customers={rows}
          isLoading={customers.isPending}
          onEdit={openForm}
          onDelete={(customer) => setConfirm({ open: true, customer })}
        />
        {customers.data && (
          <Pagination
            page={customers.data.page}
            limit={customers.data.limit}
            total={customers.data.total}
            onPageChange={setPage}
          />
        )}
      </>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Clientes"
        actions={
          <Button type="button" size="lg" className="h-11 md:h-9" onClick={() => openForm(null)}>
            <Plus aria-hidden />
            Nuevo cliente
          </Button>
        }
      />

      <SearchField
        label="Buscar clientes"
        hint="Por nombre, documento, email o teléfono"
        value={search}
        onChange={(value) => {
          setSearch(value)
          setPage(1)
        }}
        className="sm:max-w-sm"
      />

      {renderContent()}

      <FormDialog
        open={form.open}
        onOpenChange={(open) => {
          if (!open) closeForm()
        }}
        size="wide"
        title={isEditing ? 'Editar cliente' : 'Nuevo cliente'}
        description={form.customer?.name}
      >
        <CustomerForm
          key={form.formKey}
          initialValues={toFormValues(form.customer)}
          submitLabel={isEditing ? 'Guardar cambios' : 'Crear cliente'}
          onSubmit={save}
        />
      </FormDialog>

      <ConfirmDialog
        open={confirm.open}
        onOpenChange={(open) => {
          if (!open) closeConfirm()
        }}
        title={`¿Eliminar «${confirm.customer?.name ?? ''}»?`}
        description="Se borra definitivamente y no se puede deshacer."
        confirmLabel="Eliminar"
        isPending={deleteCustomer.isPending}
        onConfirm={() => {
          if (confirm.customer) void remove(confirm.customer)
        }}
      />
    </div>
  )
}
