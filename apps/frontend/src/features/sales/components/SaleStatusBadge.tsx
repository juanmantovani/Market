import { Ban, CircleCheck } from 'lucide-react'
import { cn } from '../../../shared/lib/utils'
import { SALE_STATUS_LABELS, type SaleStatus } from '../types'

// The state is always written out: color and icon only back it up.
export function SaleStatusBadge({ status }: { status: SaleStatus }) {
  const Icon = status === 'voided' ? Ban : CircleCheck
  return (
    <span
      className={cn(
        'inline-flex items-center gap-1 rounded-full px-2.5 py-1 text-xs font-medium',
        status === 'voided' ? 'bg-destructive/10 text-destructive' : 'bg-success/15 text-foreground',
      )}
    >
      <Icon className={cn('size-3.5', status === 'completed' && 'text-success')} aria-hidden />
      {SALE_STATUS_LABELS[status]}
    </span>
  )
}
