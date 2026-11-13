import { keepPreviousData, useQuery, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../../../shared/api/query-keys'
import { pickByCode } from '../lib/cart'
import type { CatalogItem, CatalogParams, SearchCatalog } from '../types'

export const CATALOG_LIMIT = 24

// The catalog shares the items cache: editing an item refreshes the counter
// too.
export function useCatalog(searchCatalog: SearchCatalog, params: CatalogParams) {
  return useQuery({
    queryKey: queryKeys.items.list(params),
    queryFn: ({ signal }) => searchCatalog(params, signal),
    placeholderData: keepPreviousData,
  })
}

export type CodeLookup = { item: CatalogItem } | { item: null; results: number }

// Finds the item a scanner (or the cashier) typed, across every category.
// It searches right away instead of waiting for the debounced listing: a
// scanner types the code and Enter in a few milliseconds.
export function useCodeLookup(searchCatalog: SearchCatalog) {
  const queryClient = useQueryClient()

  return async (code: string): Promise<CodeLookup> => {
    const params: CatalogParams = { search: code.trim(), categoryId: '', page: 1, limit: CATALOG_LIMIT }
    const page = await queryClient.fetchQuery({
      queryKey: queryKeys.items.list(params),
      queryFn: ({ signal }) => searchCatalog(params, signal),
    })
    const item = pickByCode(code, page.items)
    return item ? { item } : { item: null, results: page.items.length }
  }
}
