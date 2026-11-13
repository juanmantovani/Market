import { useRef, useState, type KeyboardEvent } from 'react'
import { useDebouncedValue } from '../../../shared/hooks/use-debounced-value'
import { formatUnitQuantity } from '../../../shared/lib/quantity'
import { Button } from '../../../shared/ui/button'
import { Input } from '../../../shared/ui/input'
import { useItemSearch } from '../hooks/use-stock'
import type { ItemChoice } from '../types'

type ItemPickerProps = {
  id: string
  value: ItemChoice | null
  onChange: (item: ItemChoice | null) => void
  invalid: boolean
  describedBy?: string
  autoFocus?: boolean
}

// Finds the item of an entry line by name, code or barcode. Enter picks the
// first match, so a barcode scanner (which types the code and presses
// Enter) fills the line on its own.
export function ItemPicker({ id, value, onChange, invalid, describedBy, autoFocus = false }: ItemPickerProps) {
  const [search, setSearch] = useState('')
  const [notFound, setNotFound] = useState(false)
  const inputRef = useRef<HTMLInputElement>(null)
  const debouncedSearch = useDebouncedValue(search)
  const { results, isFetching, find } = useItemSearch(debouncedSearch)
  const showResults = search.trim() !== '' && debouncedSearch.trim() !== ''

  function pick(item: ItemChoice) {
    onChange(item)
    setSearch('')
    setNotFound(false)
  }

  async function handleKeyDown(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key !== 'Enter') return
    event.preventDefault()
    if (search.trim() === '') return
    const matches = await find(search)
    if (matches.length > 0) {
      pick(matches[0])
    } else {
      setNotFound(true)
    }
  }

  if (value) {
    return (
      <div className="flex min-h-11 items-center justify-between gap-3 rounded-lg border bg-muted/50 px-3 py-2">
        <div className="flex min-w-0 flex-col">
          <span className="truncate text-sm font-medium">{value.name}</span>
          <span className="text-xs text-muted-foreground">
            <span className="font-mono">{value.code}</span> · Stock: {formatUnitQuantity(value.quantity, value.unit)}
          </span>
        </div>
        <Button
          type="button"
          variant="ghost"
          className="h-11 shrink-0 md:h-9"
          aria-label={`Cambiar ${value.name}`}
          onClick={() => {
            onChange(null)
            requestAnimationFrame(() => inputRef.current?.focus())
          }}
        >
          Cambiar
        </Button>
      </div>
    )
  }

  const listId = `${id}-results`

  return (
    <div className="flex flex-col gap-2">
      <Input
        ref={inputRef}
        id={id}
        autoFocus={autoFocus}
        type="search"
        autoComplete="off"
        placeholder="Buscar por nombre, código o código de barras"
        aria-invalid={invalid}
        aria-describedby={describedBy}
        aria-controls={showResults ? listId : undefined}
        value={search}
        onChange={(event) => {
          setSearch(event.target.value)
          setNotFound(false)
        }}
        onKeyDown={(event) => void handleKeyDown(event)}
      />
      {showResults && (
        <div id={listId} className="flex flex-col gap-1">
          <p className="sr-only" aria-live="polite">
            {isFetching ? 'Buscando artículos' : `${results.length} artículos encontrados`}
          </p>
          {results.length > 0 ? (
            <ul aria-label="Artículos encontrados" className="flex max-h-52 flex-col gap-1 overflow-y-auto rounded-lg border p-1">
              {results.map((item) => (
                <li key={item.itemId}>
                  <button
                    type="button"
                    className="flex min-h-11 w-full items-center justify-between gap-3 rounded-md px-3 py-2 text-left text-sm outline-none hover:bg-muted focus-visible:ring-3 focus-visible:ring-ring/50"
                    onClick={() => pick(item)}
                  >
                    <span className="flex min-w-0 flex-col">
                      <span className="truncate font-medium">{item.name}</span>
                      <span className="font-mono text-xs text-muted-foreground">{item.code}</span>
                    </span>
                    <span className="shrink-0 text-xs text-muted-foreground tabular-nums">
                      {formatUnitQuantity(item.quantity, item.unit)}
                    </span>
                  </button>
                </li>
              ))}
            </ul>
          ) : (
            !isFetching && <p className="text-sm text-muted-foreground">No se encontraron artículos. Probar con otra búsqueda.</p>
          )}
        </div>
      )}
      {notFound && !showResults && (
        <p className="text-sm text-muted-foreground">No se encontraron artículos. Probar con otra búsqueda.</p>
      )}
    </div>
  )
}
