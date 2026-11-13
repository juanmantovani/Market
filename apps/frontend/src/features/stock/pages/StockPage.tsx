import { useState } from 'react'
import { Link } from 'react-router'
import { CircleAlert, PackageCheck, PackagePlus, PackageSearch, SearchX } from 'lucide-react'
import { messageFromError } from '../../../shared/api/client'
import { useDebouncedValue } from '../../../shared/hooks/use-debounced-value'
import { formatUnitQuantity } from '../../../shared/lib/quantity'
import { Button } from '../../../shared/ui/button'
import { buttonVariants } from '../../../shared/ui/button-variants'
import { EmptyState } from '../../../shared/ui/empty-state'
import { FormDialog } from '../../../shared/ui/form-dialog'
import { NativeSelect } from '../../../shared/ui/native-select'
import { PageHeader } from '../../../shared/ui/page-header'
import { Pagination } from '../../../shared/ui/pagination'
import { SearchField } from '../../../shared/ui/search-field'
import { useToast } from '../../../shared/ui/use-toast'
import { AdjustmentForm } from '../components/AdjustmentForm'
import { EntryForm } from '../components/EntryForm'
import { MovementHistory } from '../components/MovementHistory'
import { StockList } from '../components/StockList'
import { useRegisterStockAdjustment, useRegisterStockEntry, useStockLevels } from '../hooks/use-stock'
import type { AdjustmentFields, CategoryOption, StockEntryRequest, StockLevel, SupplierOption } from '../types'

const PAGE_SIZE = 25

type StockPageProps = {
  // Composed in app/StockRoute: stock doesn't import the categories or the
  // suppliers feature.
  categories: readonly CategoryOption[]
  suppliers: readonly SupplierOption[]
}

// `open` is kept apart from what each overlay shows, so its content stays in
// place while the closing animation runs instead of flashing empty.
type LevelDialog = { open: boolean; level: StockLevel | null; formKey: number }
type EntryDialog = { open: boolean; formKey: number }

export default function StockPage({ categories, suppliers }: StockPageProps) {
  const [search, setSearch] = useState('')
  const [categoryId, setCategoryId] = useState('')
  const [lowStock, setLowStock] = useState(false)
  const [page, setPage] = useState(1)
  const [adjustment, setAdjustment] = useState<LevelDialog>({ open: false, level: null, formKey: 0 })
  const [history, setHistory] = useState<LevelDialog>({ open: false, level: null, formKey: 0 })
  const [entry, setEntry] = useState<EntryDialog>({ open: false, formKey: 0 })
  const debouncedSearch = useDebouncedValue(search)

  const levels = useStockLevels({ search: debouncedSearch, categoryId, lowStock, page, limit: PAGE_SIZE })
  const registerAdjustment = useRegisterStockAdjustment()
  const registerEntry = useRegisterStockEntry()
  const toast = useToast()

  const rows = levels.data?.items ?? []
  const isFiltered = debouncedSearch.trim() !== '' || categoryId !== ''
  const categoryNames = new Map(categories.map((category) => [category.id, category.name]))
  const supplierNames = new Map(suppliers.map((supplier) => [supplier.id, supplier.name]))

  function openDialog(setter: typeof setAdjustment, level: StockLevel) {
    setter((current) => ({ open: true, level, formKey: current.formKey + 1 }))
  }

  function closeAdjustment() {
    setAdjustment((current) => ({ ...current, open: false }))
  }

  function closeEntry() {
    setEntry((current) => ({ ...current, open: false }))
  }

  async function adjust(fields: AdjustmentFields) {
    const level = adjustment.level as StockLevel
    const movement = await registerAdjustment.mutateAsync({ itemId: level.itemId, ...fields })
    closeAdjustment()
    toast.success('Stock ajustado', `Ahora hay ${formatUnitQuantity(movement.balanceAfter, level.unit)} de «${level.name}».`)
  }

  async function receive(request: StockEntryRequest) {
    const { movements } = await registerEntry.mutateAsync(request)
    closeEntry()
    const count = movements.length === 1 ? '1 artículo' : `${movements.length} artículos`
    toast.success('Ingreso registrado', `Se actualizó el stock de ${count}.`)
  }

  function changeFilter(update: () => void) {
    update()
    setPage(1)
  }

  function renderContent() {
    if (levels.isError) {
      return (
        <EmptyState
          icon={CircleAlert}
          title="No se pudo cargar el stock"
          description={messageFromError(levels.error)}
          action={
            <Button type="button" variant="outline" size="lg" onClick={() => void levels.refetch()}>
              Reintentar
            </Button>
          }
        />
      )
    }

    if (!levels.isPending && rows.length === 0) {
      if (isFiltered) {
        return <EmptyState icon={SearchX} title="No se encontraron artículos" description="Probar con otra búsqueda o cambiar los filtros." />
      }
      if (lowStock) {
        return (
          <EmptyState
            icon={PackageCheck}
            title="No hay artículos con stock bajo"
            description="Todos los artículos están por encima de su stock mínimo."
          />
        )
      }
      return (
        <EmptyState
          icon={PackageSearch}
          title="Todavía no hay artículos"
          description="El stock se lleva por artículo: cargar los artículos primero."
          action={
            <Link to="/articulos" className={buttonVariants({ size: 'lg' })}>
              Ir a artículos
            </Link>
          }
        />
      )
    }

    return (
      <>
        <StockList
          levels={rows}
          categoryName={(id) => categoryNames.get(id) ?? '—'}
          isLoading={levels.isPending}
          onAdjust={(level) => openDialog(setAdjustment, level)}
          onShowMovements={(level) => openDialog(setHistory, level)}
        />
        {levels.data && (
          <Pagination page={levels.data.page} limit={levels.data.limit} total={levels.data.total} onPageChange={setPage} />
        )}
      </>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader
        title="Stock"
        actions={
          <Button
            type="button"
            size="lg"
            className="h-11 md:h-9"
            onClick={() => setEntry((current) => ({ open: true, formKey: current.formKey + 1 }))}
          >
            <PackagePlus aria-hidden />
            Ingresar mercadería
          </Button>
        }
      />

      <div className="flex flex-col gap-3 sm:flex-row sm:flex-wrap sm:items-center">
        <SearchField
          label="Buscar artículos"
          hint="Por nombre, código o código de barras"
          value={search}
          onChange={(value) => changeFilter(() => setSearch(value))}
          className="sm:max-w-sm"
        />
        <NativeSelect
          aria-label="Filtrar por categoría"
          className="sm:max-w-56"
          value={categoryId}
          onChange={(event) => changeFilter(() => setCategoryId(event.target.value))}
        >
          <option value="">Todas las categorías</option>
          {categories.map((category) => (
            <option key={category.id} value={category.id}>
              {category.name}
            </option>
          ))}
        </NativeSelect>
        <label className="flex min-h-11 cursor-pointer items-center gap-3 rounded-lg border bg-card px-3 text-sm font-medium has-focus-visible:ring-3 has-focus-visible:ring-ring/50 md:min-h-9">
          <input
            type="checkbox"
            className="size-5 shrink-0 accent-primary"
            checked={lowStock}
            onChange={(event) => changeFilter(() => setLowStock(event.target.checked))}
          />
          Solo stock bajo
        </label>
      </div>

      {renderContent()}

      <FormDialog
        open={adjustment.open}
        onOpenChange={(open) => {
          if (!open) closeAdjustment()
        }}
        title="Ajustar stock"
        description={adjustment.level?.name}
      >
        {adjustment.level && (
          <AdjustmentForm key={adjustment.formKey} level={adjustment.level} onSubmit={adjust} />
        )}
      </FormDialog>

      <FormDialog
        open={entry.open}
        onOpenChange={(open) => {
          if (!open) closeEntry()
        }}
        size="wide"
        title="Ingresar mercadería"
      >
        <EntryForm key={entry.formKey} suppliers={suppliers} onSubmit={receive} />
      </FormDialog>

      <FormDialog
        open={history.open}
        onOpenChange={(open) => {
          if (!open) setHistory((current) => ({ ...current, open: false }))
        }}
        size="wide"
        title="Movimientos"
        description={history.level?.name}
      >
        {history.level && (
          <MovementHistory
            key={history.formKey}
            level={history.level}
            supplierName={(id) => supplierNames.get(id) ?? 'Proveedor eliminado'}
          />
        )}
      </FormDialog>
    </div>
  )
}
