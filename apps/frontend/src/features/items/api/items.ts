import { z } from 'zod'
import { apiFetch } from '../../../shared/api/client'
import { listQuery, type ListParams } from '../../../shared/api/pagination'
import { itemPageSchema, itemSchema, type ItemRequest } from '../types'

const ITEMS_PATH = '/api/v1/items'

export type ItemListParams = ListParams & { categoryId: string; lowStock?: boolean }

function itemPath(id: string) {
  return `${ITEMS_PATH}/${encodeURIComponent(id)}`
}

export function listItems({ categoryId, lowStock, ...params }: ItemListParams, signal?: AbortSignal) {
  const query = new URLSearchParams(listQuery(params))
  if (categoryId) query.set('categoryId', categoryId)
  if (lowStock) query.set('lowStock', 'true')
  return apiFetch(`${ITEMS_PATH}?${query}`, itemPageSchema, { signal })
}

export function createItem(request: ItemRequest) {
  return apiFetch(ITEMS_PATH, itemSchema, { method: 'POST', body: request })
}

// PUT replaces every editable field, so optional fields sent as null are
// cleared.
export function updateItem(id: string, request: ItemRequest) {
  return apiFetch(itemPath(id), itemSchema, { method: 'PUT', body: request })
}

// The backend answers 204 with an empty body.
export function deleteItem(id: string) {
  return apiFetch(itemPath(id), z.undefined(), { method: 'DELETE' })
}
