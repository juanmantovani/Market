import { describe, expect, it } from 'vitest'
import { documentDigits, formatCuit, formatDni } from './document'

describe('documentDigits', () => {
  it('drops spaces, dots, dashes and slashes', () => {
    expect(documentDigits(' 30-71234567-8 ')).toBe('30712345678')
    expect(documentDigits('28.456.789')).toBe('28456789')
    expect(documentDigits('20 1234 5678/9')).toBe('20123456789')
  })

  it('keeps anything else, so validation can reject it', () => {
    expect(documentDigits('28A56')).toBe('28A56')
  })
})

describe('formatDni', () => {
  it('groups thousands with dots', () => {
    expect(formatDni('28456789')).toBe('28.456.789')
    expect(formatDni('1234567')).toBe('1.234.567')
  })

  it('leaves unexpected values as they are', () => {
    expect(formatDni('123')).toBe('123')
    expect(formatDni('123456789')).toBe('123456789')
  })
})

describe('formatCuit', () => {
  it('writes the prefix, the number and the check digit apart', () => {
    expect(formatCuit('30712345678')).toBe('30-71234567-8')
  })

  it('leaves unexpected values as they are', () => {
    expect(formatCuit('3071234567')).toBe('3071234567')
    expect(formatCuit('')).toBe('')
  })
})
