import { useState } from 'react'
import { CircleAlert, Plus, SearchX, Truck } from 'lucide-react'
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
import { SupplierForm } from '../components/SupplierForm'
import { SupplierList } from '../components/SupplierList'
import { useSuppliers, useDeleteSupplier, useSaveSupplier } from '../hooks/use-suppliers'
import { toFormValues, type Supplier, type SupplierRequest } from '../types'

const PAGE_SIZE = 25

// `open` is kept apart from what the overlay shows, so its content stays in
// place while the closing animation runs instead of flashing empty.
type FormState = { open: boolean; supplier: Supplier | null; formKey: number }
type ConfirmState = { open: boolean; supplier: Supplier | null }

export default function SuppliersPage() {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [form, setForm] = useState<FormState>({ open: false, supplier: null, formKey: 0 })
  const [confirm, setConfirm] = useState<ConfirmState>({ open: false, supplier: null })
  const debouncedSearch = useDebouncedValue(search)

  const suppliers = useSuppliers({ search: debouncedSearch, page, limit: PAGE_SIZE })
  const saveSupplier = useSaveSupplier()
  const deleteSupplier = useDeleteSupplier()
  const toast = useToast()

  const rows = suppliers.data?.items ?? []
  const isSearching = debouncedSearch.trim() !== ''
  const isEditing = form.supplier !== null

  function openForm(supplier: Supplier | null) {
    setForm((current) => ({ open: true, supplier, formKey: current.formKey + 1 }))
  }

  function closeForm() {
    setForm((current) => ({ ...current, open: false }))
  }

  function closeConfirm() {
    setConfirm((current) => ({ ...current, open: false }))
  }

  async function save(request: SupplierRequest) {
    const id = form.supplier?.id
    const saved = await saveSupplier.mutateAsync({ id, request })
    closeForm()
    toast.success(id ? 'Proveedor actualizado' : 'Proveedor creado', `«${saved.businessName}» ya está guardado.`)
  }

  async function remove(supplier: Supplier) {
    try {
      await deleteSupplier.mutateAsync(supplier.id)
      if (rows.length === 1 && page > 1) {
        setPage(page - 1)
      }
      toast.success('Proveedor eliminado', `«${supplier.businessName}» ya no existe.`)
    } catch (error) {
      toast.error('No se pudo eliminar el proveedor', messageFromError(error))
    } finally {
      closeConfirm()
    }
  }

  function renderContent() {
    if (suppliers.isError) {
      return (
        <EmptyState
          icon={CircleAlert}
          title="No se pudieron cargar los proveedores"
          description={messageFromError(suppliers.error)}
          action={
            <Button type="button" variant="outline" size="lg" onClick={() => void suppliers.refetch()}>
              Reintentar
            </Button>
          }
        />
      )
    }

    if (!suppliers.isPending && rows.length === 0) {
      return isSearching ? (
        <EmptyState
          icon={SearchX}
          title="No se encontraron proveedores"
          description="Probar con otra parte de la razón social, el CUIT, el contacto, el email o el teléfono."
        />
      ) : (
        <EmptyState
          icon={Truck}
          title="Todavía no hay proveedores"
          description="Cargar a quienes se les compra mercadería, con sus datos de contacto."
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
        <SupplierList
          suppliers={rows}
          isLoading={suppliers.isPending}
          onEdit={openForm}
          onDelete={(supplier) => setConfirm({ open: true, supplier })}
        />
        {suppliers.data && (
          <Pagination
            page={suppliers.data.page}
            limit={suppliers.data.limit}
            total={suppliers.data.total}
            onPageChange={setPage}
          />
        )}
      </>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Proveedores"
        description="A quiénes se les compra mercadería."
        actions={
          <Button type="button" size="lg" className="h-11 md:h-9" onClick={() => openForm(null)}>
            <Plus aria-hidden />
            Nuevo proveedor
          </Button>
        }
      />

      <SearchField
        label="Buscar proveedores"
        placeholder="Buscar por razón social, CUIT o contacto"
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
        title={isEditing ? 'Editar proveedor' : 'Nuevo proveedor'}
        description={form.supplier?.businessName ?? 'Completar los datos para crearlo.'}
      >
        <SupplierForm
          key={form.formKey}
          initialValues={toFormValues(form.supplier)}
          submitLabel={isEditing ? 'Guardar cambios' : 'Crear proveedor'}
          onSubmit={save}
          onCancel={closeForm}
        />
      </FormDialog>

      <ConfirmDialog
        open={confirm.open}
        onOpenChange={(open) => {
          if (!open) closeConfirm()
        }}
        title={`¿Eliminar «${confirm.supplier?.businessName ?? ''}»?`}
        description="Se borra definitivamente y no se puede deshacer."
        confirmLabel="Eliminar"
        isPending={deleteSupplier.isPending}
        onConfirm={() => {
          if (confirm.supplier) void remove(confirm.supplier)
        }}
      />
    </div>
  )
}
