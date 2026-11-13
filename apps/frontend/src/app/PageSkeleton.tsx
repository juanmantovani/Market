import { Skeleton } from '../shared/ui/skeleton'

export default function PageSkeleton() {
  return (
    <div role="status" aria-label="Cargando la página" className="flex flex-col gap-6">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
        <Skeleton className="h-8 w-40" />
        <Skeleton className="h-11 w-full rounded-lg sm:w-44 md:h-9" />
      </div>
      <Skeleton className="h-11 w-full rounded-lg sm:max-w-sm md:h-9" />
      <div className="flex flex-col gap-3">
        {Array.from({ length: 5 }, (_, index) => (
          <Skeleton key={index} className="h-16 w-full rounded-xl md:h-12" />
        ))}
      </div>
    </div>
  )
}
