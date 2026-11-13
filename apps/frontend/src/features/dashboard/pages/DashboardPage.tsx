import { CircleAlert, PackageCheck, TriangleAlert } from 'lucide-react'
import type { ReactNode } from 'react'
import { messageFromError } from '../../../shared/api/client'
import { formatMoney } from '../../../shared/lib/money'
import { formatUnitQuantity } from '../../../shared/lib/quantity'
import { Button } from '../../../shared/ui/button'
import { EmptyState } from '../../../shared/ui/empty-state'
import { PageHeader } from '../../../shared/ui/page-header'
import { Skeleton } from '../../../shared/ui/skeleton'
import { useTodaySummary } from '../hooks/use-summary'
import { METHOD_LABELS, type Summary } from '../types'

function Card({ title, children }: { title: string; children: ReactNode }) {
  return (
    <section className="flex flex-col gap-3 rounded-2xl border bg-card p-5">
      <h2 className="text-sm font-medium text-muted-foreground">{title}</h2>
      {children}
    </section>
  )
}

function SalesCards({ sales }: { sales: Summary['sales'] }) {
  return (
    <>
      <Card title="Ventas de hoy">
        <p className="text-3xl font-semibold tabular-nums">{sales.count}</p>
      </Card>
      <Card title="Total vendido">
        <p className="text-3xl font-semibold tabular-nums">{formatMoney(sales.total)}</p>
      </Card>
    </>
  )
}

function MethodsCard({ sales }: { sales: Summary['sales'] }) {
  return (
    <Card title="Total por medio de pago">
      {sales.byMethod.length === 0 ? (
        <p className="text-sm text-muted-foreground">Todavía no hay ventas hoy.</p>
      ) : (
        <ul className="flex flex-col divide-y">
          {sales.byMethod.map(({ method, amount }) => (
            <li key={method} className="flex items-center justify-between gap-3 py-2 first:pt-0 last:pb-0">
              <span>{METHOD_LABELS[method]}</span>
              <span className="font-medium tabular-nums">{formatMoney(amount)}</span>
            </li>
          ))}
        </ul>
      )}
    </Card>
  )
}

function LowStockCard({ lowStock }: { lowStock: Summary['lowStock'] }) {
  const hidden = lowStock.total - lowStock.items.length

  return (
    <Card title="Artículos con stock bajo">
      {lowStock.total === 0 ? (
        <p className="flex items-center gap-2 text-sm text-muted-foreground">
          <PackageCheck className="size-4" aria-hidden />
          Ningún artículo llegó a su stock mínimo.
        </p>
      ) : (
        <>
          <p className="text-3xl font-semibold tabular-nums">{lowStock.total}</p>
          <ul className="flex flex-col divide-y">
            {lowStock.items.map((item) => (
              <li key={item.itemId} className="flex flex-wrap items-center justify-between gap-x-3 gap-y-1 py-2">
                <span className="min-w-0">
                  <span className="block truncate font-medium">{item.name}</span>
                  <span className="block text-xs text-muted-foreground">
                    Mínimo {formatUnitQuantity(item.minStock, item.unit)}
                  </span>
                </span>
                <span className="inline-flex items-center gap-1.5 text-sm tabular-nums">
                  <TriangleAlert className="size-4 text-warning" aria-hidden />
                  <span className="sr-only">Stock bajo: quedan </span>
                  {formatUnitQuantity(item.quantity, item.unit)}
                </span>
              </li>
            ))}
          </ul>
          {hidden > 0 && (
            <p className="text-sm text-muted-foreground">
              y {hidden} más en Artículos, filtrando por stock bajo.
            </p>
          )}
        </>
      )}
    </Card>
  )
}

function LoadingCards() {
  return (
    <div role="status" aria-label="Cargando el resumen" className="grid gap-4 md:grid-cols-2">
      <Skeleton className="h-28 rounded-2xl" />
      <Skeleton className="h-28 rounded-2xl" />
      <Skeleton className="h-48 rounded-2xl" />
      <Skeleton className="h-48 rounded-2xl" />
    </div>
  )
}

export default function DashboardPage() {
  const summary = useTodaySummary()

  function renderContent() {
    if (summary.isError) {
      return (
        <EmptyState
          icon={CircleAlert}
          title="No se pudo cargar el resumen"
          description={messageFromError(summary.error)}
          action={
            <Button type="button" variant="outline" size="lg" onClick={() => void summary.refetch()}>
              Reintentar
            </Button>
          }
        />
      )
    }
    if (!summary.data) return <LoadingCards />

    return (
      <div className="grid gap-4 md:grid-cols-2">
        <SalesCards sales={summary.data.sales} />
        <MethodsCard sales={summary.data.sales} />
        <LowStockCard lowStock={summary.data.lowStock} />
      </div>
    )
  }

  return (
    <div className="flex flex-col gap-6">
      <PageHeader title="Resumen" />
      {renderContent()}
    </div>
  )
}
