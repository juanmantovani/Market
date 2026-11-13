import { keepPreviousData, useQuery } from '@tanstack/react-query'
import { listAllPages, type ListParams } from '../../../shared/api/pagination'
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

// Every category, for pickers and filters in other screens, however many
// pages it takes.
export function useCategoryOptions() {
  const query = useQuery({
    queryKey: queryKeys.categories.options,
    queryFn: ({ signal }) => listAllPages((page, limit) => listCategories({ search: '', page, limit }, signal)),
  })

  return {
    categories: (query.data ?? []).map(({ id, name }) => ({ id, name })),
    isLoading: query.isPending,
  }
}
