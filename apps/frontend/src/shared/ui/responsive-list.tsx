import type { ReactNode } from 'react'

import { cn } from '@/shared/lib/utils'
import { Skeleton } from '@/shared/ui/skeleton'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/shared/ui/table'

export type Column<T> = {
  header: string
  cell: (item: T) => ReactNode
  className?: string
}

type ResponsiveListProps<T> = {
  label: string
  items: readonly T[]
  columns: readonly Column<T>[]
  getKey: (item: T) => string
  renderCard: (item: T) => ReactNode
  isLoading?: boolean
  skeletonRows?: number
}

// A table from md: up and a stack of cards below it, so listings stay
// readable on a phone without horizontal scrolling.
export function ResponsiveList<T>({
  label,
  items,
  columns,
  getKey,
  renderCard,
  isLoading = false,
  skeletonRows = 5,
}: ResponsiveListProps<T>) {
  if (isLoading) {
    return (
      <div role="status" aria-label={`Cargando ${label.toLowerCase()}`} className="flex flex-col gap-3">
        {Array.from({ length: skeletonRows }, (_, index) => (
          <Skeleton key={index} className="h-16 w-full rounded-xl md:h-12" />
        ))}
      </div>
    )
  }

  return (
    <>
      <div className="hidden overflow-hidden rounded-xl border bg-card md:block">
        <Table aria-label={label}>
          <TableHeader>
            <TableRow className="hover:bg-transparent">
              {columns.map((column) => (
                <TableHead key={column.header} className={cn('px-4', column.className)}>
                  {column.header}
                </TableHead>
              ))}
            </TableRow>
          </TableHeader>
          <TableBody>
            {items.map((item) => (
              <TableRow key={getKey(item)}>
                {columns.map((column) => (
                  <TableCell key={column.header} className={cn('px-4 py-3', column.className)}>
                    {column.cell(item)}
                  </TableCell>
                ))}
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </div>

      <ul aria-label={label} className="flex flex-col gap-3 md:hidden">
        {items.map((item) => (
          <li key={getKey(item)} className="rounded-xl border bg-card p-4">
            {renderCard(item)}
          </li>
        ))}
      </ul>
    </>
  )
}
