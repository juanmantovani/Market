import { describe, expect, it } from 'vitest'
import { centsToInput, formatMoney, parseMoney } from './money'

describe('formatMoney', () => {
  it('formats cents as Argentine pesos', () => {
    expect(formatMoney(123_450)).toMatch(/^\$\s?1\.234,50$/)
    expect(formatMoney(0)).toMatch(/^\$\s?0,00$/)
  })
})

describe('centsToInput', () => {
  it('shows an editable amount without thousands separators', () => {
    expect(centsToInput(123_450)).toBe('1234,50')
    expect(centsToInput(123_405)).toBe('1234,05')
    expect(centsToInput(100_000)).toBe('1000')
    expect(centsToInput(null)).toBe('')
  })
})

describe('parseMoney', () => {
  it.each([
    ['1.234,50', 123_450],
    ['1234,5', 123_450],
    ['1234', 123_400],
    ['$ 1.234', 123_400],
    ['1.234.567', 123_456_700],
    ['1234.50', 123_450],
    ['0,99', 99],
    ['  12  ', 1_200],
  ])('reads %s as %i cents', (input, cents) => {
    expect(parseMoney(input)).toBe(cents)
  })

  it.each(['', '   ', 'abc', '1,234,5', '12,345', '-5', '1.2.3,4,5'])('rejects %j', (input) => {
    expect(parseMoney(input)).toBeNull()
  })
})
