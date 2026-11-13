import { useState, type KeyboardEvent, type ReactNode } from 'react'
import { keepPreviousData, useQuery, useQueryClient, type QueryKey } from '@tanstack/react-query'
import { Check } from 'lucide-react'
import { messageFromError } from '../../../shared/api/client'
import type { Page } from '../../../shared/api/pagination'
import { useDebouncedValue } from '../../../shared/hooks/use-debounced-value'
import { Button } from '../../../shared/ui/button'
import { FormDialog } from '../../../shared/ui/form-dialog'
import { SearchField } from '../../../shared/ui/search-field'
import type { PartyParams } from '../types'

const LIMIT = 20

type ExtraOption = {
  label: string
  description: string
  isCurrent: boolean
  onPick: () => void
}

type PartyPickerProps<T> = {
  open: boolean
  onClose: () => void
  title: string
  searchLabel: string
  hint: string
  emptyText: string
  queryKey: (params: PartyParams) => QueryKey
  search: (params: PartyParams, signal?: AbortSignal) => Promise<Page<T>>
  getKey: (option: T) => string
  isCurrent: (option: T) => boolean
  renderOption: (option: T) => ReactNode
  // What Enter picks among the results of what was typed, such as an exact
  // code; the first result when it returns null.
  pickOnEnter?: (term: string, results: readonly T[]) => T | null
  onPick: (option: T) => void
  // A choice outside the results, such as "no seller".
  extra?: ExtraOption
}

// A modal to choose who takes part in the sale: type to search, Enter picks,
// the arrows walk the results and Escape closes it.
export function PartyPicker<T>(props: PartyPickerProps<T>) {
  return (
    <FormDialog open={props.open} onOpenChange={(open) => !open && props.onClose()} title={props.title}>
      {props.open && <PickerContent {...props} />}
    </FormDialog>
  )
}

const optionClass =
  'flex min-h-11 w-full items-center justify-between gap-3 rounded-lg px-3 py-2 text-left text-sm transition-colors hover:bg-muted focus-visible:ring-3 focus-visible:ring-ring/50 focus-visible:outline-none'

function PickerContent<T>({
  onClose,
  searchLabel,
  hint,
  emptyText,
  queryKey,
  search,
  getKey,
  isCurrent,
  renderOption,
  pickOnEnter,
  onPick,
  extra,
}: PartyPickerProps<T>) {
  const queryClient = useQueryClient()
  const [term, setTerm] = useState('')
  const [notice, setNotice] = useState<string | null>(null)
  const debounced = useDebouncedValue(term.trim(), 200)
  const params: PartyParams = { search: debounced, page: 1, limit: LIMIT }
  const results = useQuery({
    queryKey: queryKey(params),
    queryFn: ({ signal }) => search(params, signal),
    placeholderData: keepPreviousData,
  })

  function pick(option: T) {
    onPick(option)
    onClose()
  }

  // Enter searches right away instead of trusting a listing that may still
  // show what was typed a moment ago.
  async function submit() {
    const fresh: PartyParams = { search: term.trim(), page: 1, limit: LIMIT }
    try {
      const page = await queryClient.fetchQuery({ queryKey: queryKey(fresh), queryFn: ({ signal }) => search(fresh, signal) })
      const chosen = pickOnEnter?.(term, page.items) ?? page.items[0]
      if (chosen) {
        pick(chosen)
      } else {
        setNotice(`No hay resultados para “${term.trim()}”.`)
      }
    } catch (error) {
      setNotice(messageFromError(error))
    }
  }

  function handleSearchKey(event: KeyboardEvent<HTMLInputElement>) {
    if (event.key === 'Enter') {
      event.preventDefault()
      void submit()
    } else if (event.key === 'ArrowDown') {
      const first = event.currentTarget.closest('[data-party-picker]')?.querySelector<HTMLButtonElement>('[data-party-option]')
      if (first) {
        event.preventDefault()
        first.focus()
      }
    }
  }

  function moveFocus(event: KeyboardEvent<HTMLUListElement>) {
    const step = event.key === 'ArrowDown' ? 1 : event.key === 'ArrowUp' ? -1 : 0
    if (step === 0) return
    const options = [...event.currentTarget.querySelectorAll<HTMLButtonElement>('[data-party-option]')]
    const next = options[options.indexOf(document.activeElement as HTMLButtonElement) + step]
    if (next) {
      event.preventDefault()
      next.focus()
    }
  }

  const items = results.data?.items ?? []

  return (
    <div data-party-picker className="flex flex-col gap-3">
      <SearchField
        label={searchLabel}
        hint={hint}
        value={term}
        onChange={(value) => {
          setTerm(value)
          setNotice(null)
        }}
        onKeyDown={handleSearchKey}
        autoFocus
      />
      {notice && (
        <p role="status" className="rounded-lg bg-muted px-3 py-2 text-sm">
          {notice}
        </p>
      )}
      <ul aria-label="Resultados" className="flex max-h-80 flex-col gap-0.5 overflow-y-auto" onKeyDown={moveFocus}>
        {extra && (
          <li>
            <button
              type="button"
              data-party-option
              aria-current={extra.isCurrent || undefined}
              className={optionClass}
              onClick={() => {
                extra.onPick()
                onClose()
              }}
            >
              <span className="flex flex-col">
                <span className="font-medium">{extra.label}</span>
                <span className="text-xs text-muted-foreground">{extra.description}</span>
              </span>
              {extra.isCurrent && <Check className="size-4 text-primary" aria-hidden />}
            </button>
          </li>
        )}
        {items.map((option) => (
          <li key={getKey(option)}>
            <button
              type="button"
              data-party-option
              aria-current={isCurrent(option) || undefined}
              className={optionClass}
              onClick={() => pick(option)}
            >
              {renderOption(option)}
              {isCurrent(option) && <Check className="size-4 shrink-0 text-primary" aria-hidden />}
            </button>
          </li>
        ))}
      </ul>
      {results.isPending && <p className="px-3 text-sm text-muted-foreground">Buscando…</p>}
      {results.isError && !results.data && (
        <div className="flex items-center justify-between gap-3 px-3">
          <p className="text-sm text-muted-foreground">No se pudo buscar. Revisar la conexión.</p>
          <Button type="button" variant="outline" size="lg" className="h-11 md:h-9" onClick={() => void results.refetch()}>
            Reintentar
          </Button>
        </div>
      )}
      {results.isSuccess && items.length === 0 && <p className="px-3 text-sm text-muted-foreground">{emptyText}</p>}
    </div>
  )
}
