import { z } from 'zod'
import { apiFetch } from '../../../shared/api/client'
import { listQuery, type ListParams } from '../../../shared/api/pagination'
import { categoryPageSchema, categorySchema, type CategoryFormValues } from '../types'

const CATEGORIES_PATH = '/api/v1/categories'

function categoryPath(id: string) {
  return `${CATEGORIES_PATH}/${encodeURIComponent(id)}`
}

export function listCategories(params: ListParams, signal?: AbortSignal) {
  return apiFetch(`${CATEGORIES_PATH}?${listQuery(params)}`, categoryPageSchema, { signal })
}

export function createCategory(values: CategoryFormValues) {
  return apiFetch(CATEGORIES_PATH, categorySchema, {
    method: 'POST',
    body: { name: values.name.trim(), description: values.description.trim() || null },
  })
}

// An empty description is sent as is: the backend reads it as "clear it".
export function updateCategory(id: string, values: CategoryFormValues) {
  return apiFetch(categoryPath(id), categorySchema, {
    method: 'PATCH',
    body: { name: values.name.trim(), description: values.description.trim() },
  })
}

// The backend answers 204 with an empty body.
export function deleteCategory(id: string) {
  return apiFetch(categoryPath(id), z.undefined(), { method: 'DELETE' })
}
