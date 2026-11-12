import { useState } from 'react'
import { CircleAlert, Plus, SearchX, UserCheck } from 'lucide-react'
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
import { SellerForm } from '../components/SellerForm'
import { SellerList } from '../components/SellerList'
import { useSellers, useDeleteSeller, useSaveSeller } from '../hooks/use-sellers'
import { toFormValues, type Seller, type SellerRequest } from '../types'

const PAGE_SIZE = 25

// `open` is kept apart from what the overlay shows, so its content stays in
// place while the closing animation runs instead of flashing empty.
type FormState = { open: boolean; seller: Seller | null; formKey: number }
type ConfirmState = { open: boolean; seller: Seller | null }

export default function SellersPage() {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [form, setForm] = useState<FormState>({ open: false, seller: null, formKey: 0 })
  const [confirm, setConfirm] = useState<ConfirmState>({ open: false, seller: null })
  const debouncedSearch = useDebouncedValue(search)

  const sellers = useSellers({ search: debouncedSearch, page, limit: PAGE_SIZE })
  const saveSeller = useSaveSeller()
  const deleteSeller = useDeleteSeller()
  const toast = useToast()

  const rows = sellers.data?.items ?? []
  const isSearching = debouncedSearch.trim() !== ''
  const isEditing = form.seller !== null

  function openForm(seller: Seller | null) {
    setForm((current) => ({ open: true, seller, formKey: current.formKey + 1 }))
  }

  function closeForm() {
    setForm((current) => ({ ...current, open: false }))
  }

  function closeConfirm() {
    setConfirm((current) => ({ ...current, open: false }))
  }

  async function save(request: SellerRequest) {
    const id = form.seller?.id
    const saved = await saveSeller.mutateAsync({ id, request })
    closeForm()
    toast.success(id ? 'Vendedor actualizado' : 'Vendedor creado', `«${saved.name}» ya está guardado.`)
  }

  async function remove(seller: Seller) {
    try {
      await deleteSeller.mutateAsync(seller.id)
      if (rows.length === 1 && page > 1) {
        setPage(page - 1)
      }
      toast.success('Vendedor eliminado', `«${seller.name}» ya no existe.`)
    } catch (error) {
      toast.error('No se pudo eliminar el vendedor', messageFromError(error))
    } finally {
      closeConfirm()
    }
  }

  function renderContent() {
    if (sellers.isError) {
      return (
        <EmptyState
          icon={CircleAlert}
          title="No se pudieron cargar los vendedores"
          description={messageFromError(sellers.error)}
          action={
            <Button type="button" variant="outline" size="lg" onClick={() => void sellers.refetch()}>
              Reintentar
            </Button>
          }
        />
      )
    }

    if (!sellers.isPending && rows.length === 0) {
      return isSearching ? (
        <EmptyState
          icon={SearchX}
          title="No se encontraron vendedores"
          description="Probar con otra parte del nombre o con el código."
        />
      ) : (
        <EmptyState
          icon={UserCheck}
          title="Todavía no hay vendedores"
          description="Cargar a quienes atienden en el mostrador, con un código corto para elegirlos al vender."
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
        <SellerList
          sellers={rows}
          isLoading={sellers.isPending}
          onEdit={openForm}
          onDelete={(seller) => setConfirm({ open: true, seller })}
        />
        {sellers.data && (
          <Pagination
            page={sellers.data.page}
            limit={sellers.data.limit}
            total={sellers.data.total}
            onPageChange={setPage}
          />
        )}
      </>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Vendedores"
        actions={
          <Button type="button" size="lg" className="h-11 md:h-9" onClick={() => openForm(null)}>
            <Plus aria-hidden />
            Nuevo vendedor
          </Button>
        }
      />

      <SearchField
        label="Buscar vendedores"
        hint="Por nombre o código"
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
        title={isEditing ? 'Editar vendedor' : 'Nuevo vendedor'}
        description={form.seller?.name}
      >
        <SellerForm
          key={form.formKey}
          initialValues={toFormValues(form.seller)}
          submitLabel={isEditing ? 'Guardar cambios' : 'Crear vendedor'}
          onSubmit={save}
          onCancel={closeForm}
        />
      </FormDialog>

      <ConfirmDialog
        open={confirm.open}
        onOpenChange={(open) => {
          if (!open) closeConfirm()
        }}
        title={`¿Eliminar «${confirm.seller?.name ?? ''}»?`}
        description="Se borra definitivamente y no se puede deshacer."
        confirmLabel="Eliminar"
        isPending={deleteSeller.isPending}
        onConfirm={() => {
          if (confirm.seller) void remove(confirm.seller)
        }}
      />
    </div>
  )
}
