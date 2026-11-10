import { ChevronLeft, ChevronRight } from 'lucide-react'

import { Button } from '@/shared/ui/button'

type PaginationProps = {
  page: number
  limit: number
  total: number
  onPageChange: (page: number) => void
}

export function Pagination({ page, limit, total, onPageChange }: PaginationProps) {
  if (total === 0) {
    return null
  }

  const first = (page - 1) * limit + 1
  const last = Math.min(page * limit, total)
  const lastPage = Math.max(1, Math.ceil(total / limit))

  return (
    <nav
      aria-label="Paginación"
      className="flex flex-col items-center gap-3 sm:flex-row sm:justify-between"
    >
      <p className="text-sm text-muted-foreground">
        {first}–{last} de {total}
      </p>
      <div className="flex gap-2">
        <Button
          type="button"
          variant="outline"
          size="lg"
          className="h-11 md:h-9"
          disabled={page <= 1}
          onClick={() => onPageChange(page - 1)}
        >
          <ChevronLeft aria-hidden />
          Anterior
        </Button>
        <Button
          type="button"
          variant="outline"
          size="lg"
          className="h-11 md:h-9"
          disabled={page >= lastPage}
          onClick={() => onPageChange(page + 1)}
        >
          Siguiente
          <ChevronRight aria-hidden />
        </Button>
      </div>
    </nav>
  )
}
