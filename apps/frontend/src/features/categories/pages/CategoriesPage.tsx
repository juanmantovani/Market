import { useState } from 'react'
import { CircleAlert, Plus, SearchX, Tags } from 'lucide-react'
import { messageFromError } from '../../../shared/api/client'
import type { StatusFilter as Status } from '../../../shared/api/pagination'
import { useDebouncedValue } from '../../../shared/hooks/use-debounced-value'
import { Button } from '../../../shared/ui/button'
import { ConfirmDialog } from '../../../shared/ui/confirm-dialog'
import { EmptyState } from '../../../shared/ui/empty-state'
import { FormSheet } from '../../../shared/ui/form-sheet'
import { PageHeader } from '../../../shared/ui/page-header'
import { Pagination } from '../../../shared/ui/pagination'
import { SearchField } from '../../../shared/ui/search-field'
import { StatusFilter } from '../../../shared/ui/status-filter'
import { useToast } from '../../../shared/ui/use-toast'
import { CategoryForm } from '../components/CategoryForm'
import { CategoryList } from '../components/CategoryList'
import { useCategories } from '../hooks/use-categories'
import { useSaveCategory, useSetCategoryActive } from '../hooks/use-category-mutations'
import { toFormValues, type Category } from '../types'

const PAGE_SIZE = 25
const STATUS_LABELS = { active: 'Activas', inactive: 'Inactivas', all: 'Todas' }

type Editing = { mode: 'create' } | { mode: 'edit'; category: Category } | null

export default function CategoriesPage() {
  const [search, setSearch] = useState('')
  const [status, setStatus] = useState<Status>('active')
  const [page, setPage] = useState(1)
  const [editing, setEditing] = useState<Editing>(null)
  const [toDeactivate, setToDeactivate] = useState<Category | null>(null)
  const debouncedSearch = useDebouncedValue(search)

  const categories = useCategories({ search: debouncedSearch, status, page, limit: PAGE_SIZE })
  const saveCategory = useSaveCategory()
  const setCategoryActive = useSetCategoryActive()
  const toast = useToast()

  const items = categories.data?.items ?? []
  const isFiltered = debouncedSearch.trim() !== '' || status !== 'active'

  function changeSearch(value: string) {
    setSearch(value)
    setPage(1)
  }

  function changeStatus(value: Status) {
    setStatus(value)
    setPage(1)
  }

  async function save(values: Parameters<typeof saveCategory.mutateAsync>[0]['values']) {
    const id = editing?.mode === 'edit' ? editing.category.id : undefined
    const saved = await saveCategory.mutateAsync({ id, values })
    setEditing(null)
    toast.success(id ? 'Categoría actualizada' : 'Categoría creada', `«${saved.name}» ya está guardada.`)
  }

  async function changeActive(category: Category, active: boolean) {
    try {
      await setCategoryActive.mutateAsync({ id: category.id, active })
      if (!active && items.length === 1 && page > 1 && status !== 'all') {
        setPage(page - 1)
      }
      toast.success(
        active ? 'Categoría reactivada' : 'Categoría desactivada',
        active
          ? `«${category.name}» vuelve a estar disponible.`
          : `«${category.name}» ya no aparece al cargar artículos. Podés reactivarla cuando quieras.`,
      )
    } catch (error) {
      toast.error('No se pudo completar la acción', messageFromError(error))
    } finally {
      setToDeactivate(null)
    }
  }

  const newCategoryButton = (
    <Button type="button" size="lg" className="h-11 md:h-9" onClick={() => setEditing({ mode: 'create' })}>
      <Plus aria-hidden />
      Nueva categoría
    </Button>
  )

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
      return isFiltered ? (
        <EmptyState
          icon={SearchX}
          title="No encontramos categorías"
          description="Probá con otra búsqueda o cambiá el filtro de estado."
        />
      ) : (
        <EmptyState
          icon={Tags}
          title="Todavía no hay categorías"
          description="Las categorías agrupan los artículos para encontrarlos rápido en la venta y en el stock."
          action={
            <Button type="button" size="lg" onClick={() => setEditing({ mode: 'create' })}>
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
          onEdit={(category) => setEditing({ mode: 'edit', category })}
          onDeactivate={setToDeactivate}
          onReactivate={(category) => void changeActive(category, true)}
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
        actions={newCategoryButton}
      />

      <div className="flex flex-col gap-3 sm:flex-row sm:items-center">
        <SearchField
          label="Buscar categorías"
          placeholder="Buscar por nombre o descripción"
          value={search}
          onChange={changeSearch}
          className="sm:max-w-sm"
        />
        <StatusFilter value={status} onChange={changeStatus} labels={STATUS_LABELS} />
      </div>

      {renderContent()}

      <FormSheet
        open={editing !== null}
        onOpenChange={(open) => {
          if (!open) setEditing(null)
        }}
        title={editing?.mode === 'edit' ? 'Editar categoría' : 'Nueva categoría'}
        description={editing?.mode === 'edit' ? editing.category.name : 'Completá los datos para crearla.'}
      >
        {editing && (
          <CategoryForm
            key={editing.mode === 'edit' ? editing.category.id : 'new'}
            initialValues={toFormValues(editing.mode === 'edit' ? editing.category : null)}
            submitLabel={editing.mode === 'edit' ? 'Guardar cambios' : 'Crear categoría'}
            onSubmit={save}
            onCancel={() => setEditing(null)}
          />
        )}
      </FormSheet>

      <ConfirmDialog
        open={toDeactivate !== null}
        onOpenChange={(open) => {
          if (!open) setToDeactivate(null)
        }}
        title={`¿Desactivar «${toDeactivate?.name ?? ''}»?`}
        description="No va a aparecer al cargar artículos. Los artículos que ya la usan no cambian, y podés reactivarla cuando quieras."
        confirmLabel="Desactivar"
        isPending={setCategoryActive.isPending}
        onConfirm={() => {
          if (toDeactivate) void changeActive(toDeactivate, false)
        }}
      />
    </div>
  )
}
