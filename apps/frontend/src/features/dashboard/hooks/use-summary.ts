import { useQuery } from '@tanstack/react-query'
import { queryKeys } from '../../../shared/api/query-keys'
import { getSummary } from '../api/dashboard'
import { todayRange } from '../lib/today'

export function useTodaySummary() {
  const { from, to } = todayRange()

  return useQuery({
    queryKey: queryKeys.dashboard.summary(from, to),
    queryFn: ({ signal }) => getSummary(from, to, signal),
  })
}
