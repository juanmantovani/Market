import { useQuery } from '@tanstack/react-query'
import { fetchHealth } from '../shared/api/health'
import { queryKeys } from '../shared/api/query-keys'
import { cn } from '../shared/lib/utils'

const STATUS = {
  checking: { label: 'Conectando…', dot: 'bg-muted-foreground animate-pulse' },
  online: { label: 'En línea', dot: 'bg-success' },
  offline: { label: 'Sin conexión', dot: 'bg-destructive' },
}

export default function ServerStatus() {
  const { isPending, isError } = useQuery({
    queryKey: queryKeys.health,
    queryFn: ({ signal }) => fetchHealth(signal),
    refetchInterval: 30_000,
    retry: false,
  })

  const status = STATUS[isPending ? 'checking' : isError ? 'offline' : 'online']

  return (
    <span
      role="status"
      title={status.label}
      className="inline-flex h-8 items-center gap-2 rounded-full border bg-card px-3 text-xs font-medium text-muted-foreground"
    >
      <span className={cn('size-2 rounded-full', status.dot)} aria-hidden />
      <span className="sr-only sm:not-sr-only">{status.label}</span>
    </span>
  )
}
