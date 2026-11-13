import type { SaleListParams } from '../types'

// "2026-11-13" as the local midnight that starts that day, in RFC 3339.
function startOfDay(day: string, offsetDays = 0): string | null {
  const match = /^(\d{4})-(\d{2})-(\d{2})$/.exec(day)
  if (!match) return null
  const [, year, month, date] = match
  return new Date(Number(year), Number(month) - 1, Number(date) + offsetDays).toISOString()
}

// The API takes instants with `from` included and `to` excluded; people pick
// whole days with both ends included, so `to` becomes the next midnight.
export function salesQuery({ status, from, to, page, limit }: SaleListParams): string {
  const query = new URLSearchParams()
  const fromInstant = startOfDay(from)
  const toInstant = startOfDay(to, 1)
  if (fromInstant) query.set('from', fromInstant)
  if (toInstant) query.set('to', toInstant)
  if (status) query.set('status', status)
  query.set('page', String(page))
  query.set('limit', String(limit))
  return query.toString()
}

export function isDateRangeInverted(from: string, to: string): boolean {
  return from !== '' && to !== '' && from > to
}
