import { keepPreviousData, useQuery } from '@tanstack/react-query'
import type { ListParams } from '../../../shared/api/pagination'
import { queryKeys } from '../../../shared/api/query-keys'
import { listCategories } from '../api/categories'

// Keeps showing the previous page while the next one loads, so typing a
// search or paging doesn't flash a skeleton.
export function useCategories(params: ListParams) {
  return useQuery({
    queryKey: queryKeys.categories.list(params),
    queryFn: ({ signal }) => listCategories(params, signal),
    placeholderData: keepPreviousData,
  })
}
