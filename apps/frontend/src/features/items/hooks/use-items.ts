import { keepPreviousData, useMutation, useQuery, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../../../shared/api/query-keys'
import { createItem, deleteItem, listItems, updateItem, type ItemListParams } from '../api/items'
import type { ItemRequest } from '../types'

export function useItems(params: ItemListParams) {
  return useQuery({
    queryKey: queryKeys.items.list(params),
    queryFn: ({ signal }) => listItems(params, signal),
    placeholderData: keepPreviousData,
  })
}

export function useSaveItem() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, request }: { id?: string; request: ItemRequest }) =>
      id ? updateItem(id, request) : createItem(request),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.items.all }),
  })
}

export function useDeleteItem() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteItem(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.items.all }),
  })
}
