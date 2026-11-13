import { apiFetch } from '../../../shared/api/client'
import { summarySchema } from '../types'

// `from` is included and `to` excluded: both are RFC 3339 instants.
export function getSummary(from: string, to: string, signal?: AbortSignal) {
  const query = new URLSearchParams({ from, to })
  return apiFetch(`/api/v1/dashboard/summary?${query}`, summarySchema, { signal })
}
