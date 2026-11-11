import { describe, expect, it } from 'vitest'
import { formatQuantity, parseQuantity } from './quantity'

describe('formatQuantity', () => {
  it('drops trailing zeros of the fraction', () => {
    expect(formatQuantity(3000)).toBe('3')
    expect(formatQuantity(2500)).toBe('2,5')
    expect(formatQuantity(250)).toBe('0,25')
    expect(formatQuantity(1005)).toBe('1,005')
  })
})

describe('parseQuantity', () => {
  it('reads whole and fractional amounts', () => {
    expect(parseQuantity('3', false)).toBe(3000)
    expect(parseQuantity('2,5', true)).toBe(2500)
    expect(parseQuantity('0.250', true)).toBe(250)
    expect(parseQuantity(' 4,000 ', false)).toBe(4000)
  })

  it('rejects fractions for items sold by the piece', () => {
    expect(parseQuantity('2,5', false)).toBeNull()
  })

  it.each(['', 'dos', '-1', '1,2345', '1,2,3'])('rejects %j', (input) => {
    expect(parseQuantity(input, true)).toBeNull()
  })
})
