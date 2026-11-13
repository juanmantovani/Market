import { cn } from '../../../shared/lib/utils'
import type { CategoryOption } from '../types'

type CategoryFilterProps = {
  categories: CategoryOption[]
  value: string
  onChange: (categoryId: string) => void
}

// Chips that narrow the catalog to one category. On a phone they scroll
// sideways instead of wrapping over the items.
export function CategoryFilter({ categories, value, onChange }: CategoryFilterProps) {
  const options = [{ id: '', name: 'Todas' }, ...categories]

  return (
    <div role="group" aria-label="Categorías" className="-mx-4 flex gap-2 overflow-x-auto px-4 pb-1 md:mx-0 md:flex-wrap md:px-0">
      {options.map((category) => {
        const selected = category.id === value
        return (
          <button
            key={category.id || 'all'}
            type="button"
            aria-pressed={selected}
            onClick={() => onChange(category.id)}
            className={cn(
              'h-11 shrink-0 rounded-full border px-4 text-sm font-medium whitespace-nowrap transition-colors md:h-9',
              'focus-visible:ring-3 focus-visible:ring-ring/50 focus-visible:outline-none',
              selected ? 'border-primary bg-primary text-primary-foreground' : 'bg-background hover:bg-muted',
            )}
          >
            {category.name}
          </button>
        )
      })}
    </div>
  )
}
