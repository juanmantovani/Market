import { describe, expect, it } from 'vitest'
import { formatPercent, parsePercent, percentToInput } from './percent'

describe('formatPercent', () => {
  it('writes hundredths of a percent the Argentine way', () => {
    expect(formatPercent(250)).toBe('2,5 %')
    expect(formatPercent(1225)).toBe('12,25 %')
    expect(formatPercent(300)).toBe('3 %')
    expect(formatPercent(0)).toBe('0 %')
  })
})

describe('percentToInput', () => {
  it('leaves out needless decimals', () => {
    expect(percentToInput(250)).toBe('2,5')
    expect(percentToInput(1205)).toBe('12,05')
    expect(percentToInput(10000)).toBe('100')
    expect(percentToInput(null)).toBe('')
  })
})

describe('parsePercent', () => {
  it('reads a comma or a dot as the decimal separator', () => {
    expect(parsePercent('2,5')).toBe(250)
    expect(parsePercent('2.5')).toBe(250)
    expect(parsePercent(' 12,25 % ')).toBe(1225)
    expect(parsePercent('100')).toBe(10000)
    expect(parsePercent('0')).toBe(0)
  })

  it('rejects anything that is not a percentage', () => {
    expect(parsePercent('')).toBeNull()
    expect(parsePercent('dos')).toBeNull()
    expect(parsePercent('2,555')).toBeNull()
    expect(parsePercent('-1')).toBeNull()
    expect(parsePercent('1000')).toBeNull()
  })
})
