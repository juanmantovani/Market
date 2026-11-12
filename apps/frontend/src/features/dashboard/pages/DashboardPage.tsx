import { ShoppingBag } from 'lucide-react'
import { EmptyState } from '../../../shared/ui/empty-state'
import { PageHeader } from '../../../shared/ui/page-header'

export default function DashboardPage() {
  return (
    <div className="flex flex-col gap-6">
      <PageHeader title="Inicio" description="Resumen del día" />
      <EmptyState
        icon={ShoppingBag}
        title="Todavía no hay ventas hoy"
        description="Cuando se registren ventas, aquí se verán el total del día, los medios de pago y los artículos con stock bajo."
      />
    </div>
  )
}
