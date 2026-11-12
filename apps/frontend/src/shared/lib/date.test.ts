import { describe, expect, it } from 'vitest'
import { formatDateTime } from './date'

describe('formatDateTime', () => {
  it('shows the day and the time in the local time zone', () => {
    const iso = '2026-11-11T15:30:00Z'
    const minutes = String(new Date(iso).getMinutes()).padStart(2, '0')

    const text = formatDateTime(iso)

    expect(text).toMatch(/^\d{1,2}\/\d{1,2}\/\d{2,4}/)
    expect(text).toContain(`:${minutes}`)
  })
})
