import { TriangleAlert } from 'lucide-react'

export function LowStockBadge() {
  return (
    <span className="inline-flex items-center gap-1 rounded-full bg-warning/15 px-2 py-0.5 text-xs font-medium text-foreground">
      <TriangleAlert className="size-3.5 text-warning" aria-hidden />
      Stock bajo
    </span>
  )
}
