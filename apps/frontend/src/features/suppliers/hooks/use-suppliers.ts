import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { ListParams } from '../../../shared/api/pagination'
import { queryKeys } from '../../../shared/api/query-keys'
import { createSupplier, deleteSupplier, listSuppliers, updateSupplier } from '../api/suppliers'
import type { SupplierRequest } from '../types'

export function useSuppliers(params: ListParams) {
  return useQuery({
    queryKey: queryKeys.suppliers.list(params),
    queryFn: ({ signal }) => listSuppliers(params, signal),
    placeholderData: keepPreviousData,
  })
}

// Enough for the item form to offer every supplier of a small business.
const ALL_SUPPLIERS: ListParams = { search: '', page: 1, limit: 100 }

export function useSupplierOptions() {
  const query = useSuppliers(ALL_SUPPLIERS)

  return {
    suppliers: (query.data?.items ?? []).map(({ id, businessName }) => ({ id, name: businessName })),
  }
}

export function useSaveSupplier() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, request }: { id?: string; request: SupplierRequest }) =>
      id ? updateSupplier(id, request) : createSupplier(request),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.suppliers.all }),
  })
}

export function useDeleteSupplier() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteSupplier(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.suppliers.all }),
  })
}
