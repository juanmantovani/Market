import { useState } from 'react'
import { Link } from 'react-router'
import { CircleAlert, Package, Plus, SearchX, Tags } from 'lucide-react'
import { messageFromError } from '../../../shared/api/client'
import { useDebouncedValue } from '../../../shared/hooks/use-debounced-value'
import { Button } from '../../../shared/ui/button'
import { buttonVariants } from '../../../shared/ui/button-variants'
import { ConfirmDialog } from '../../../shared/ui/confirm-dialog'
import { EmptyState } from '../../../shared/ui/empty-state'
import { FormDialog } from '../../../shared/ui/form-dialog'
import { NativeSelect } from '../../../shared/ui/native-select'
import { PageHeader } from '../../../shared/ui/page-header'
import { Pagination } from '../../../shared/ui/pagination'
import { SearchField } from '../../../shared/ui/search-field'
import { useToast } from '../../../shared/ui/use-toast'
import { ItemForm } from '../components/ItemForm'
import { ItemList } from '../components/ItemList'
import { useDeleteItem, useItems, useSaveItem } from '../hooks/use-items'
import {
  toFormValues,
  type CategoryOption,
  type Item,
  type ItemRequest,
  type SupplierOption,
  type UseStockQuantities,
} from '../types'

const PAGE_SIZE = 25

type ItemsPageProps = {
  // Composed in app/ItemsRoute: items don't import the categories, the
  // suppliers or the stock feature. The stock is asked for the items of the
  // page being shown, which only this page knows.
  categories: readonly CategoryOption[]
  categoriesLoading: boolean
  suppliers: readonly SupplierOption[]
  useStockQuantities: UseStockQuantities
}

// `open` is kept apart from what the overlay shows, so its content stays in
// place while the closing animation runs instead of flashing empty.
type FormState = { open: boolean; item: Item | null; formKey: number }
type ConfirmState = { open: boolean; item: Item | null }

export default function ItemsPage({ categories, categoriesLoading, suppliers, useStockQuantities }: ItemsPageProps) {
  const [search, setSearch] = useState('')
  const [categoryId, setCategoryId] = useState('')
  const [page, setPage] = useState(1)
  const [form, setForm] = useState<FormState>({ open: false, item: null, formKey: 0 })
  const [confirm, setConfirm] = useState<ConfirmState>({ open: false, item: null })
  const debouncedSearch = useDebouncedValue(search)

  const items = useItems({ search: debouncedSearch, categoryId, page, limit: PAGE_SIZE })
  const saveItem = useSaveItem()
  const deleteItem = useDeleteItem()
  const toast = useToast()

  const rows = items.data?.items ?? []
  const stock = useStockQuantities(rows.map((row) => row.id))
  const isFiltered = debouncedSearch.trim() !== '' || categoryId !== ''
  const isEditing = form.item !== null
  const categoryNames = new Map(categories.map((category) => [category.id, category.name]))

  function openForm(item: Item | null) {
    setForm((current) => ({ open: true, item, formKey: current.formKey + 1 }))
  }

  function closeForm() {
    setForm((current) => ({ ...current, open: false }))
  }

  function closeConfirm() {
    setConfirm((current) => ({ ...current, open: false }))
  }

  async function save(request: ItemRequest) {
    const id = form.item?.id
    const saved = await saveItem.mutateAsync({ id, request })
    closeForm()
    toast.success(id ? 'Artículo actualizado' : 'Artículo creado', `«${saved.name}» ya está guardado.`)
  }

  async function remove(item: Item) {
    try {
      await deleteItem.mutateAsync(item.id)
      if (rows.length === 1 && page > 1) {
        setPage(page - 1)
      }
      toast.success('Artículo eliminado', `«${item.name}» ya no existe.`)
    } catch (error) {
      toast.error('No se pudo eliminar el artículo', messageFromError(error))
    } finally {
      closeConfirm()
    }
  }

  function renderContent() {
    if (!categoriesLoading && categories.length === 0) {
      return (
        <EmptyState
          icon={Tags}
          title="Primero, crear una categoría"
          description="Cada artículo pertenece a una categoría. Crear al menos una para empezar a cargar artículos."
          action={
            <Link to="/categorias" className={buttonVariants({ size: 'lg' })}>
              Ir a categorías
            </Link>
          }
        />
      )
    }

    if (items.isError) {
      return (
        <EmptyState
          icon={CircleAlert}
          title="No se pudieron cargar los artículos"
          description={messageFromError(items.error)}
          action={
            <Button type="button" variant="outline" size="lg" onClick={() => void items.refetch()}>
              Reintentar
            </Button>
          }
        />
      )
    }

    if (!items.isPending && rows.length === 0) {
      return isFiltered ? (
        <EmptyState icon={SearchX} title="No se encontraron artículos" description="Probar con otra búsqueda o cambiar la categoría." />
      ) : (
        <EmptyState
          icon={Package}
          title="Todavía no hay artículos"
          description="Cargar los artículos que se venden, con su precio y su código de barras."
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
        <ItemList
          items={rows}
          stock={stock}
          categoryName={(id) => categoryNames.get(id) ?? '—'}
          isLoading={items.isPending}
          onEdit={openForm}
          onDelete={(item) => setConfirm({ open: true, item })}
        />
        {items.data && (
          <Pagination page={items.data.page} limit={items.data.limit} total={items.data.total} onPageChange={setPage} />
        )}
      </>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Artículos"
        actions={
          <Button
            type="button"
            size="lg"
            className="h-11 md:h-9"
            disabled={categories.length === 0}
            onClick={() => openForm(null)}
          >
            <Plus aria-hidden />
            Nuevo artículo
          </Button>
        }
      />

      <div className="flex flex-col gap-3 sm:flex-row sm:items-start">
        <SearchField
          label="Buscar artículos"
          hint="Por nombre, código o código de barras"
          value={search}
          onChange={(value) => {
            setSearch(value)
            setPage(1)
          }}
          className="sm:max-w-sm"
        />
        <NativeSelect
          aria-label="Filtrar por categoría"
          className="sm:max-w-56"
          value={categoryId}
          onChange={(event) => {
            setCategoryId(event.target.value)
            setPage(1)
          }}
        >
          <option value="">Todas las categorías</option>
          {categories.map((category) => (
            <option key={category.id} value={category.id}>
              {category.name}
            </option>
          ))}
        </NativeSelect>
      </div>

      {renderContent()}

      <FormDialog
        open={form.open}
        onOpenChange={(open) => {
          if (!open) closeForm()
        }}
        size="wide"
        title={isEditing ? 'Editar artículo' : 'Nuevo artículo'}
        description={form.item?.name}
      >
        <ItemForm
          key={form.formKey}
          initialValues={toFormValues(form.item)}
          categories={categories}
          suppliers={suppliers}
          submitLabel={isEditing ? 'Guardar cambios' : 'Crear artículo'}
          onSubmit={save}
        />
      </FormDialog>

      <ConfirmDialog
        open={confirm.open}
        onOpenChange={(open) => {
          if (!open) closeConfirm()
        }}
        title={`¿Eliminar «${confirm.item?.name ?? ''}»?`}
        description="Se borra definitivamente y no se puede deshacer."
        confirmLabel="Eliminar"
        isPending={deleteItem.isPending}
        onConfirm={() => {
          if (confirm.item) void remove(confirm.item)
        }}
      />
    </div>
  )
}
