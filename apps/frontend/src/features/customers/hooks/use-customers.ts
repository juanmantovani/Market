import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import type { ListParams } from '../../../shared/api/pagination'
import { queryKeys } from '../../../shared/api/query-keys'
import { createCustomer, deleteCustomer, listCustomers, updateCustomer } from '../api/customers'
import type { CustomerRequest } from '../types'

export function useCustomers(params: ListParams) {
  return useQuery({
    queryKey: queryKeys.customers.list(params),
    queryFn: ({ signal }) => listCustomers(params, signal),
    placeholderData: keepPreviousData,
  })
}

export function useSaveCustomer() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, request }: { id?: string; request: CustomerRequest }) =>
      id ? updateCustomer(id, request) : createCustomer(request),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.customers.all }),
  })
}

export function useDeleteCustomer() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteCustomer(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.customers.all }),
  })
}
