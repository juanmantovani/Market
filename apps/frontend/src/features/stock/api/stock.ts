import { apiFetch } from '../../../shared/api/client'
import { listQuery, type ListParams } from '../../../shared/api/pagination'
import {
  stockEntryResponseSchema,
  stockLevelPageSchema,
  stockMovementPageSchema,
  stockMovementSchema,
  type StockAdjustmentRequest,
  type StockEntryRequest,
} from '../types'

const STOCK_PATH = '/api/v1/stock'

export type StockListParams = ListParams & { categoryId: string; lowStock: boolean }

export type MovementListParams = { itemId: string; page: number; limit: number }

export function listStockLevels({ categoryId, lowStock, ...params }: StockListParams, signal?: AbortSignal) {
  const query = new URLSearchParams(listQuery(params))
  if (categoryId) query.set('categoryId', categoryId)
  if (lowStock) query.set('lowStock', 'true')
  return apiFetch(`${STOCK_PATH}?${query}`, stockLevelPageSchema, { signal })
}

export function listStockMovements({ itemId, page, limit }: MovementListParams, signal?: AbortSignal) {
  const query = new URLSearchParams({ itemId, page: String(page), limit: String(limit) })
  return apiFetch(`${STOCK_PATH}/movements?${query}`, stockMovementPageSchema, { signal })
}

export function createStockEntry(request: StockEntryRequest) {
  return apiFetch(`${STOCK_PATH}/entries`, stockEntryResponseSchema, { method: 'POST', body: request })
}

export function createStockAdjustment(request: StockAdjustmentRequest) {
  return apiFetch(`${STOCK_PATH}/adjustments`, stockMovementSchema, { method: 'POST', body: request })
}
