import { useState } from 'react'
import { CircleAlert, Receipt, SearchX } from 'lucide-react'
import { messageFromError } from '../../../shared/api/client'
import { Button } from '../../../shared/ui/button'
import { EmptyState } from '../../../shared/ui/empty-state'
import { Field, FieldError, FieldLabel } from '../../../shared/ui/field'
import { FormDialog } from '../../../shared/ui/form-dialog'
import { Input } from '../../../shared/ui/input'
import { NativeSelect } from '../../../shared/ui/native-select'
import { PageHeader } from '../../../shared/ui/page-header'
import { Pagination } from '../../../shared/ui/pagination'
import { useToast } from '../../../shared/ui/use-toast'
import { SaleDetail } from '../components/SaleDetail'
import { SaleList } from '../components/SaleList'
import { useSales, useVoidSale } from '../hooks/use-sales'
import { isDateRangeInverted } from '../lib/sale-filters'
import { SALE_STATUS_LABELS, type SaleRecord, type SaleStatusFilter } from '../types'

const PAGE_SIZE = 25

// `open` is kept apart from the sale so the content stays in place while the
// closing animation runs.
type DetailState = { open: boolean; sale: SaleRecord | null }

export default function SalesPage() {
  const [status, setStatus] = useState<SaleStatusFilter>('')
  const [from, setFrom] = useState('')
  const [to, setTo] = useState('')
  const [page, setPage] = useState(1)
  const [detail, setDetail] = useState<DetailState>({ open: false, sale: null })

  const rangeInverted = isDateRangeInverted(from, to)
  const sales = useSales({ status, from, to: rangeInverted ? '' : to, page, limit: PAGE_SIZE })
  const voidSale = useVoidSale()
  const toast = useToast()

  const rows = sales.data?.items ?? []
  const isFiltering = status !== '' || from !== '' || to !== ''

  function changeFilter(apply: () => void) {
    apply()
    setPage(1)
  }

  async function voidSelected(reason: string) {
    if (!detail.sale) return
    const number = detail.sale.number
    try {
      const voided = await voidSale.mutateAsync({ id: detail.sale.id, reason })
      setDetail({ open: true, sale: voided })
      toast.success('Venta anulada', `La venta N.º ${number} quedó anulada y el stock se repuso.`)
    } catch (error) {
      toast.error('No se pudo anular la venta', messageFromError(error))
    }
  }

  function renderContent() {
    if (sales.isError) {
      return (
        <EmptyState
          icon={CircleAlert}
          title="No se pudieron cargar las ventas"
          description={messageFromError(sales.error)}
          action={
            <Button type="button" variant="outline" size="lg" onClick={() => void sales.refetch()}>
              Reintentar
            </Button>
          }
        />
      )
    }

    if (!sales.isPending && rows.length === 0) {
      return isFiltering ? (
        <EmptyState icon={SearchX} title="No hay ventas con estos filtros" description="Probar con otro estado o con otras fechas." />
      ) : (
        <EmptyState icon={Receipt} title="Todavía no hay ventas" description="Las ventas que se cobren en Vender van a aparecer acá." />
      )
    }

    return (
      <>
        <SaleList sales={rows} isLoading={sales.isPending} onOpen={(sale) => setDetail({ open: true, sale })} />
        {sales.data && (
          <Pagination page={sales.data.page} limit={sales.data.limit} total={sales.data.total} onPageChange={setPage} />
        )}
      </>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title="Ventas" />

      <form
        aria-label="Filtros de ventas"
        className="grid grid-cols-1 gap-3 sm:grid-cols-3 lg:max-w-3xl"
        onSubmit={(event) => event.preventDefault()}
      >
        <Field>
          <FieldLabel htmlFor="sales-status">Estado</FieldLabel>
          <NativeSelect
            id="sales-status"
            value={status}
            onChange={(event) => changeFilter(() => setStatus(event.target.value as SaleStatusFilter))}
          >
            <option value="">Todas</option>
            <option value="completed">{SALE_STATUS_LABELS.completed}</option>
            <option value="voided">{SALE_STATUS_LABELS.voided}</option>
          </NativeSelect>
        </Field>
        <Field>
          <FieldLabel htmlFor="sales-from">Desde</FieldLabel>
          <Input id="sales-from" type="date" value={from} onChange={(event) => changeFilter(() => setFrom(event.target.value))} />
        </Field>
        <Field>
          <FieldLabel htmlFor="sales-to">Hasta</FieldLabel>
          <Input
            id="sales-to"
            type="date"
            value={to}
            aria-invalid={rangeInverted}
            onChange={(event) => changeFilter(() => setTo(event.target.value))}
          />
        </Field>
        <FieldError className="sm:col-span-3">
          {rangeInverted ? 'La fecha «Desde» no puede ser posterior a «Hasta».' : ''}
        </FieldError>
      </form>

      {renderContent()}

      <FormDialog
        open={detail.open}
        onOpenChange={(open) => {
          if (!open) setDetail((current) => ({ ...current, open: false }))
        }}
        size="wide"
        title={detail.sale ? `Venta N.º ${detail.sale.number}` : 'Venta'}
      >
        {detail.sale && (
          <SaleDetail key={detail.sale.id} sale={detail.sale} isVoiding={voidSale.isPending} onVoid={voidSelected} />
        )}
      </FormDialog>
    </div>
  )
}
