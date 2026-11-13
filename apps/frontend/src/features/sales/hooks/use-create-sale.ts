import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../../../shared/api/query-keys'
import { createSale } from '../api/sales'
import type { CreateSaleRequest } from '../types'

// A sale moves stock, so the catalog and the stock levels are read again.
export function useCreateSale() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ request, idempotencyKey }: { request: CreateSaleRequest; idempotencyKey: string }) =>
      createSale(request, idempotencyKey),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: queryKeys.sales.all })
      void queryClient.invalidateQueries({ queryKey: queryKeys.items.all })
      void queryClient.invalidateQueries({ queryKey: queryKeys.stock.all })
    },
  })
}
