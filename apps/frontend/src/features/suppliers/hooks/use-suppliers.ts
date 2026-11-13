import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { listAllPages, type ListParams } from '../../../shared/api/pagination'
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

// Every supplier, for the item form and stock entries, however many pages
// it takes.
export function useSupplierOptions() {
  const query = useQuery({
    queryKey: queryKeys.suppliers.options,
    queryFn: ({ signal }) => listAllPages((page, limit) => listSuppliers({ search: '', page, limit }, signal)),
  })

  return {
    suppliers: (query.data ?? []).map(({ id, businessName }) => ({ id, name: businessName })),
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
