import { useState } from 'react'
import { CircleAlert, Plus, SearchX, Tags } from 'lucide-react'
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
import { CategoryForm } from '../components/CategoryForm'
import { CategoryList } from '../components/CategoryList'
import { useCategories } from '../hooks/use-categories'
import { useDeleteCategory, useSaveCategory } from '../hooks/use-category-mutations'
import { toFormValues, type Category, type CategoryFormValues } from '../types'

const PAGE_SIZE = 25

// `open` is kept apart from what the overlay shows, so its content stays in
// place while the closing animation runs instead of flashing empty.
type FormState = { open: boolean; category: Category | null; formKey: number }
type ConfirmState = { open: boolean; category: Category | null }

export default function CategoriesPage() {
  const [search, setSearch] = useState('')
  const [page, setPage] = useState(1)
  const [form, setForm] = useState<FormState>({ open: false, category: null, formKey: 0 })
  const [confirm, setConfirm] = useState<ConfirmState>({ open: false, category: null })
  const debouncedSearch = useDebouncedValue(search)

  const categories = useCategories({ search: debouncedSearch, page, limit: PAGE_SIZE })
  const saveCategory = useSaveCategory()
  const deleteCategory = useDeleteCategory()
  const toast = useToast()

  const items = categories.data?.items ?? []
  const isSearching = debouncedSearch.trim() !== ''
  const isEditing = form.category !== null

  function changeSearch(value: string) {
    setSearch(value)
    setPage(1)
  }

  function openForm(category: Category | null) {
    setForm((current) => ({ open: true, category, formKey: current.formKey + 1 }))
  }

  function closeForm() {
    setForm((current) => ({ ...current, open: false }))
  }

  function closeConfirm() {
    setConfirm((current) => ({ ...current, open: false }))
  }

  async function save(values: CategoryFormValues) {
    const id = form.category?.id
    const saved = await saveCategory.mutateAsync({ id, values })
    closeForm()
    toast.success(id ? 'Categoría actualizada' : 'Categoría creada', `«${saved.name}» ya está guardada.`)
  }

  async function remove(category: Category) {
    try {
      await deleteCategory.mutateAsync(category.id)
      if (items.length === 1 && page > 1) {
        setPage(page - 1)
      }
      toast.success('Categoría eliminada', `«${category.name}» ya no existe.`)
    } catch (error) {
      toast.error('No se pudo eliminar la categoría', messageFromError(error))
    } finally {
      closeConfirm()
    }
  }

  function renderContent() {
    if (categories.isError) {
      return (
        <EmptyState
          icon={CircleAlert}
          title="No pudimos cargar las categorías"
          description={messageFromError(categories.error)}
          action={
            <Button type="button" variant="outline" size="lg" onClick={() => void categories.refetch()}>
              Reintentar
            </Button>
          }
        />
      )
    }

    if (!categories.isPending && items.length === 0) {
      return isSearching ? (
        <EmptyState
          icon={SearchX}
          title="No encontramos categorías"
          description="Probá con otra búsqueda."
        />
      ) : (
        <EmptyState
          icon={Tags}
          title="Todavía no hay categorías"
          description="Las categorías agrupan los artículos para encontrarlos rápido en la venta y en el stock."
          action={
            <Button type="button" size="lg" onClick={() => openForm(null)}>
              <Plus aria-hidden />
              Crear la primera
            </Button>
          }
        />
      )
    }

    return (
      <>
        <CategoryList
          categories={items}
          isLoading={categories.isPending}
          onEdit={openForm}
          onDelete={(category) => setConfirm({ open: true, category })}
        />
        {categories.data && (
          <Pagination
            page={categories.data.page}
            limit={categories.data.limit}
            total={categories.data.total}
            onPageChange={setPage}
          />
        )}
      </>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Categorías"
        description="Agrupan los artículos para buscarlos y filtrarlos."
        actions={
          <Button type="button" size="lg" className="h-11 md:h-9" onClick={() => openForm(null)}>
            <Plus aria-hidden />
            Nueva categoría
          </Button>
        }
      />

      <SearchField
        label="Buscar categorías"
        placeholder="Buscar por nombre o descripción"
        value={search}
        onChange={changeSearch}
        className="sm:max-w-sm"
      />

      {renderContent()}

      <FormDialog
        open={form.open}
        onOpenChange={(open) => {
          if (!open) closeForm()
        }}
        title={isEditing ? 'Editar categoría' : 'Nueva categoría'}
        description={form.category?.name ?? 'Completá los datos para crearla.'}
      >
        <CategoryForm
          key={form.formKey}
          initialValues={toFormValues(form.category)}
          submitLabel={isEditing ? 'Guardar cambios' : 'Crear categoría'}
          onSubmit={save}
          onCancel={closeForm}
        />
      </FormDialog>

      <ConfirmDialog
        open={confirm.open}
        onOpenChange={(open) => {
          if (!open) closeConfirm()
        }}
        title={`¿Eliminar «${confirm.category?.name ?? ''}»?`}
        description="Se borra definitivamente y no se puede deshacer."
        confirmLabel="Eliminar"
        isPending={deleteCategory.isPending}
        onConfirm={() => {
          if (confirm.category) void remove(confirm.category)
        }}
      />
    </div>
  )
}
