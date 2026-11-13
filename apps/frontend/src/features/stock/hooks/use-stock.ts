import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../../../shared/api/query-keys'
import {
  createStockAdjustment,
  createStockEntry,
  createStockInitial,
  listStockLevels,
  listStockLevelsOf,
  listStockMovements,
  type MovementListParams,
  type StockListParams,
} from '../api/stock'
import type { ItemChoice, StockAdjustmentRequest, StockEntryRequest, StockInitialRequest } from '../types'

// The stock levels of the given items, which are not listed again while the
// list is empty.
export function useStockLevelsOf(itemIds: readonly string[]) {
  return useQuery({
    queryKey: queryKeys.stock.levelsOf(itemIds),
    queryFn: ({ signal }) => listStockLevelsOf(itemIds, signal),
    enabled: itemIds.length > 0,
  })
}

export function useStockMovements(params: MovementListParams) {
  return useQuery({
    queryKey: queryKeys.stock.movements(params),
    queryFn: ({ signal }) => listStockMovements(params, signal),
    placeholderData: keepPreviousData,
  })
}

const SEARCH_LIMIT = 8

function searchParams(search: string): StockListParams {
  return { search: search.trim(), page: 1, limit: SEARCH_LIMIT }
}

// Items to pick for a stock entry, found by name, code or barcode. `find`
// searches right away instead of waiting for the debounce, so a barcode
// scanner (which types and presses Enter at once) gets its item.
export function useItemSearch(search: string) {
  const queryClient = useQueryClient()
  const params = searchParams(search)
  const query = useQuery({
    queryKey: queryKeys.stock.levels(params),
    queryFn: ({ signal }) => listStockLevels(params, signal),
    enabled: params.search !== '',
  })

  async function find(text: string): Promise<ItemChoice[]> {
    const exact = searchParams(text)
    const page = await queryClient.fetchQuery({
      queryKey: queryKeys.stock.levels(exact),
      queryFn: ({ signal }) => listStockLevels(exact, signal),
    })
    return page.items
  }

  return { results: query.data?.items ?? [], isFetching: query.isFetching, find }
}

export function useRegisterStockEntry() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: StockEntryRequest) => createStockEntry(request),
    // An entry with a unit cost also changes the cost of its items.
    onSuccess: () =>
      Promise.all([
        queryClient.invalidateQueries({ queryKey: queryKeys.stock.all }),
        queryClient.invalidateQueries({ queryKey: queryKeys.items.all }),
      ]),
  })
}

export function useRegisterStockAdjustment() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: StockAdjustmentRequest) => createStockAdjustment(request),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.stock.all }),
  })
}

export function useRegisterStockInitial() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (request: StockInitialRequest) => createStockInitial(request),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.stock.all }),
  })
}
