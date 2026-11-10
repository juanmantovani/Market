import type { StatusFilter as Status } from '@/shared/api/pagination'
import { ToggleGroup, ToggleGroupItem } from '@/shared/ui/toggle-group'

const STATUSES: readonly Status[] = ['active', 'inactive', 'all']

type StatusFilterProps = {
  value: Status
  onChange: (status: Status) => void
  // Spanish agrees in gender with the entity: "Activas" for categorías,
  // "Activos" for artículos.
  labels: Record<Status, string>
}

export function StatusFilter({ value, onChange, labels }: StatusFilterProps) {
  return (
    <ToggleGroup
      aria-label="Estado"
      variant="outline"
      value={[value]}
      onValueChange={(selected: string[]) => {
        const next = STATUSES.find((status) => status === selected[0])
        if (next) onChange(next)
      }}
      className="w-full flex-nowrap sm:w-fit"
    >
      {STATUSES.map((status) => (
        <ToggleGroupItem key={status} value={status} className="h-11 flex-1 px-4 sm:flex-none md:h-9">
          {labels[status]}
        </ToggleGroupItem>
      ))}
    </ToggleGroup>
  )
}
