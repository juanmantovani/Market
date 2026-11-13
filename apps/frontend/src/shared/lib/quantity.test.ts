import { describe, expect, it } from 'vitest'
import { formatQuantity, formatUnitQuantity, isLowStock, parseQuantity, unitShort } from './quantity'

describe('formatQuantity', () => {
  it('drops trailing zeros of the fraction', () => {
    expect(formatQuantity(3000)).toBe('3')
    expect(formatQuantity(2500)).toBe('2,5')
    expect(formatQuantity(250)).toBe('0,25')
    expect(formatQuantity(1005)).toBe('1,005')
  })

  it('keeps the sign of negative quantities, also below one unit', () => {
    expect(formatQuantity(-2000)).toBe('-2')
    expect(formatQuantity(-1500)).toBe('-1,5')
    expect(formatQuantity(-500)).toBe('-0,5')
    expect(formatQuantity(0)).toBe('0')
  })
})

describe('formatUnitQuantity', () => {
  it('adds the short name of the unit', () => {
    expect(formatUnitQuantity(4000, 'unit')).toBe('4 u.')
    expect(formatUnitQuantity(1800, 'kg')).toBe('1,8 kg')
    expect(formatUnitQuantity(-250, 'l')).toBe('-0,25 l')
    expect(unitShort('m')).toBe('m')
  })
})

describe('isLowStock', () => {
  it('is low when the stock is at or below a minimum', () => {
    expect(isLowStock(3000, 5000)).toBe(true)
    expect(isLowStock(5000, 5000)).toBe(true)
    expect(isLowStock(5001, 5000)).toBe(false)
  })

  it('is never low without a minimum', () => {
    expect(isLowStock(0, 0)).toBe(false)
    expect(isLowStock(2000, 0)).toBe(false)
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
