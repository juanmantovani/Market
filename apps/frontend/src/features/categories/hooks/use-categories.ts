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

const ALL_CATEGORIES: ListParams = { search: '', page: 1, limit: 100 }

// Every category, for pickers and filters in other screens. A catalog with
// more than 100 categories would need a searchable picker instead.
export function useCategoryOptions() {
  const query = useCategories(ALL_CATEGORIES)

  return {
    categories: (query.data?.items ?? []).map(({ id, name }) => ({ id, name })),
    isLoading: query.isPending,
  }
}
