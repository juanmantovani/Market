import { apiFetch } from '../../../shared/api/client'
import { salesQuery } from '../lib/sale-filters'
import {
  saleRecordPageSchema,
  saleRecordSchema,
  saleSchema,
  type CreateSaleRequest,
  type SaleListParams,
} from '../types'

const SALES_PATH = '/api/v1/sales'

export function createSale(request: CreateSaleRequest, idempotencyKey: string) {
  return apiFetch(SALES_PATH, saleSchema, {
    method: 'POST',
    body: request,
    headers: { 'Idempotency-Key': idempotencyKey },
  })
}

export function listSales(params: SaleListParams, signal?: AbortSignal) {
  return apiFetch(`${SALES_PATH}?${salesQuery(params)}`, saleRecordPageSchema, { signal })
}

export function voidSale(id: string, reason: string) {
  return apiFetch(`${SALES_PATH}/${encodeURIComponent(id)}/void`, saleRecordSchema, {
    method: 'POST',
    body: { reason },
  })
}
