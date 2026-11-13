import { z } from 'zod'
import { apiFetch } from '../../../shared/api/client'
import { listQuery, type ListParams } from '../../../shared/api/pagination'
import { supplierPageSchema, supplierSchema, type SupplierRequest } from '../types'

const SUPPLIERS_PATH = '/api/v1/suppliers'

function supplierPath(id: string) {
  return `${SUPPLIERS_PATH}/${encodeURIComponent(id)}`
}

export function listSuppliers(params: ListParams, signal?: AbortSignal) {
  return apiFetch(`${SUPPLIERS_PATH}?${listQuery(params)}`, supplierPageSchema, { signal })
}

export function createSupplier(request: SupplierRequest) {
  return apiFetch(SUPPLIERS_PATH, supplierSchema, { method: 'POST', body: request })
}

// PUT replaces every editable field, so optional fields sent as null are
// cleared.
export function updateSupplier(id: string, request: SupplierRequest) {
  return apiFetch(supplierPath(id), supplierSchema, { method: 'PUT', body: request })
}

// The backend answers 204 with an empty body.
export function deleteSupplier(id: string) {
  return apiFetch(supplierPath(id), z.undefined(), { method: 'DELETE' })
}
