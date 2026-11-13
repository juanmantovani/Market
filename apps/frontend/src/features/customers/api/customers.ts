import { z } from 'zod'
import { apiFetch } from '../../../shared/api/client'
import { listQuery, type ListParams } from '../../../shared/api/pagination'
import { customerPageSchema, customerSchema, type CustomerRequest } from '../types'

const CUSTOMERS_PATH = '/api/v1/customers'

function customerPath(id: string) {
  return `${CUSTOMERS_PATH}/${encodeURIComponent(id)}`
}

export function listCustomers(params: ListParams, signal?: AbortSignal) {
  return apiFetch(`${CUSTOMERS_PATH}?${listQuery(params)}`, customerPageSchema, { signal })
}

export function createCustomer(request: CustomerRequest) {
  return apiFetch(CUSTOMERS_PATH, customerSchema, { method: 'POST', body: request })
}

// PUT replaces every editable field, so optional fields sent as null are
// cleared.
export function updateCustomer(id: string, request: CustomerRequest) {
  return apiFetch(customerPath(id), customerSchema, { method: 'PUT', body: request })
}

// The backend answers 204 with an empty body.
export function deleteCustomer(id: string) {
  return apiFetch(customerPath(id), z.undefined(), { method: 'DELETE' })
}
