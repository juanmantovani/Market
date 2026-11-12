import { z } from 'zod'
import { apiFetch } from '../../../shared/api/client'
import { listQuery, type ListParams } from '../../../shared/api/pagination'
import { sellerPageSchema, sellerSchema, type SellerRequest } from '../types'

const SELLERS_PATH = '/api/v1/sellers'

function sellerPath(id: string) {
  return `${SELLERS_PATH}/${encodeURIComponent(id)}`
}

export function listSellers(params: ListParams, signal?: AbortSignal) {
  return apiFetch(`${SELLERS_PATH}?${listQuery(params)}`, sellerPageSchema, { signal })
}

export function createSeller(request: SellerRequest) {
  return apiFetch(SELLERS_PATH, sellerSchema, { method: 'POST', body: request })
}

// PUT replaces every editable field, so optional fields sent as null are
// cleared.
export function updateSeller(id: string, request: SellerRequest) {
  return apiFetch(sellerPath(id), sellerSchema, { method: 'PUT', body: request })
}

// The backend answers 204 with an empty body.
export function deleteSeller(id: string) {
  return apiFetch(sellerPath(id), z.undefined(), { method: 'DELETE' })
}
