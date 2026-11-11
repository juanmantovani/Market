import { useMutation, useQueryClient } from '@tanstack/react-query'
import { queryKeys } from '../../../shared/api/query-keys'
import { createCategory, deleteCategory, updateCategory } from '../api/categories'
import type { CategoryFormValues } from '../types'

type SaveCategoryInput = {
  id?: string
  values: CategoryFormValues
}

export function useSaveCategory() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: ({ id, values }: SaveCategoryInput) =>
      id ? updateCategory(id, values) : createCategory(values),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.categories.all }),
  })
}

export function useDeleteCategory() {
  const queryClient = useQueryClient()

  return useMutation({
    mutationFn: (id: string) => deleteCategory(id),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: queryKeys.categories.all }),
  })
}
