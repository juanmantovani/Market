import { describe, expect, it } from 'vitest'
import { isDateRangeInverted, salesQuery } from './sale-filters'

describe('salesQuery', () => {
  it('sends only the page without filters', () => {
    expect(salesQuery({ status: '', from: '', to: '', page: 2, limit: 25 })).toBe('page=2&limit=25')
  })

  it('turns the picked days into local instants, the last day included', () => {
    const query = new URLSearchParams(salesQuery({ status: 'voided', from: '2026-11-10', to: '2026-11-12', page: 1, limit: 25 }))
    expect(query.get('from')).toBe(new Date(2026, 10, 10).toISOString())
    expect(query.get('to')).toBe(new Date(2026, 10, 13).toISOString())
    expect(query.get('status')).toBe('voided')
  })

  it('ignores a half-typed date', () => {
    expect(salesQuery({ status: '', from: '2026-11', to: '', page: 1, limit: 25 })).toBe('page=1&limit=25')
  })
})

describe('isDateRangeInverted', () => {
  it('flags a start after the end only when both are set', () => {
    expect(isDateRangeInverted('2026-11-12', '2026-11-10')).toBe(true)
    expect(isDateRangeInverted('2026-11-10', '2026-11-10')).toBe(false)
    expect(isDateRangeInverted('', '2026-11-10')).toBe(false)
  })
})
