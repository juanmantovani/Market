import { apiFetch } from '../../../shared/api/client'
import { saleSchema, type CreateSaleRequest } from '../types'

export function createSale(request: CreateSaleRequest, idempotencyKey: string) {
  return apiFetch('/api/v1/sales', saleSchema, {
    method: 'POST',
    body: request,
    headers: { 'Idempotency-Key': idempotencyKey },
  })
}
