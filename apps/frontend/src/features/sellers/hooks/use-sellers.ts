import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { ListParams } from '../../../shared/api/pagination'
import { queryKeys } from '../../../shared/api/query-keys'
import { createSeller, deleteSeller, listSellers, updateSeller } from '../api/sellers'
import type { SellerRequest } from '../types'

export function useSellers(params: ListParams) {
  return useQuery({
    queryKey: queryKeys.sellers.list(params),
    queryFn: ({ signal }) => listSellers(params, signal),
    placeholderData: keepPreviousData,
  })
}

export function useSaveSeller() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, request }: { id?: string; request: SellerRequest }) =>
      id ? updateSeller(id, request) : createSeller(request),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.sellers.all }),
  })
}

export function useDeleteSeller() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteSeller(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.sellers.all }),
  })
}
