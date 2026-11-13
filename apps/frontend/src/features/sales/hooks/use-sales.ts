import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../../../shared/api/query-keys'
import { listSales, voidSale } from '../api/sales'
import type { SaleListParams } from '../types'

export function useSales(params: SaleListParams) {
  return useQuery({
    queryKey: queryKeys.sales.list(params),
    queryFn: ({ signal }) => listSales(params, signal),
    placeholderData: keepPreviousData,
  })
}

// Voiding puts the stock back, so the items and their levels are read again.
export function useVoidSale() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, reason }: { id: string; reason: string }) => voidSale(id, reason),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.sales.all })
      void queryClient.invalidateQueries({ queryKey: queryKeys.items.all })
      void queryClient.invalidateQueries({ queryKey: queryKeys.stock.all })
    },
  })
}
