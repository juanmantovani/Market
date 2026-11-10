import { Search, X } from 'lucide-react'

import { cn } from '@/shared/lib/utils'
import { Input } from '@/shared/ui/input'

type SearchFieldProps = {
  label: string
  value: string
  onChange: (value: string) => void
  placeholder?: string
  className?: string
}

export function SearchField({ label, value, onChange, placeholder, className }: SearchFieldProps) {
  return (
    <div className={cn('relative w-full', className)}>
      <Search
        className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-muted-foreground"
        aria-hidden
      />
      <Input
        type="search"
        aria-label={label}
        placeholder={placeholder ?? label}
        value={value}
        onChange={(event) => onChange(event.target.value)}
        className="pr-11 pl-9 [&::-webkit-search-cancel-button]:hidden"
      />
      {value && (
        <button
          type="button"
          aria-label="Limpiar búsqueda"
          onClick={() => onChange('')}
          className="absolute top-1/2 right-1 flex size-9 -translate-y-1/2 items-center justify-center rounded-md text-muted-foreground hover:text-foreground focus-visible:ring-3 focus-visible:ring-ring/50 focus-visible:outline-none md:size-7"
        >
          <X className="size-4" aria-hidden />
        </button>
      )}
    </div>
  )
}
