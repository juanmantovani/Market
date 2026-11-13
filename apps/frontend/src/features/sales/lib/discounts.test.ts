import { describe, expect, it } from 'vitest'
import type { CartLine, Discount } from '../types'
import { discountOff, effectiveDiscount, formatDiscount, PERCENT_SCALE, priceCart, prorate } from './discounts'

const percent = (value: number): Discount => ({ type: 'percent', value })
const amount = (value: number): Discount => ({ type: 'amount', value })

// The same sale as the backend's domain tests: two colas and 0,750 kg of
// bread with 10 % off, 640.000 + 189.000 cents once the lines are priced.
const cola: CartLine = { itemId: 'coca', code: 'B1', name: 'Gaseosa cola', unit: 'unit', unitPrice: 320_000, quantity: 2_000 }
const bread: CartLine = { itemId: 'pan', code: 'P1', name: 'Pan', unit: 'kg', unitPrice: 280_000, quantity: 750, discount: percent(1_000) }

describe('discountOff', () => {
  it.each([
    { name: 'a percent', discount: percent(1_000), base: 640_000, want: 64_000 },
    { name: 'half a cent rounds away from zero', discount: percent(1_000), base: 1_005, want: 101 },
    { name: 'below half a cent rounds down', discount: percent(1_000), base: 1_004, want: 100 },
    { name: 'a fraction of a point', discount: percent(1_250), base: 10_001, want: 1_250 },
    { name: '100 %', discount: percent(PERCENT_SCALE), base: 777, want: 777 },
    { name: 'a fixed amount', discount: amount(5_000), base: 640_000, want: 5_000 },
    { name: 'a fixed amount over the base is capped', discount: amount(9_000), base: 640, want: 640 },
    { name: 'no discount', discount: null, base: 640_000, want: 0 },
    { name: 'an undefined discount', discount: undefined, base: 640_000, want: 0 },
    { name: 'a percent over 100 is capped', discount: percent(20_000), base: 500, want: 500 },
    { name: 'a product beyond 2^53 stays exact', discount: percent(3_333), base: Number.MAX_SAFE_INTEGER, want: Number((BigInt(Number.MAX_SAFE_INTEGER) * 3_333n * 2n + 10_000n) / 20_000n) },
  ])('$name', ({ discount, base, want }) => {
    expect(discountOff(discount, base)).toBe(want)
  })
})

describe('effectiveDiscount', () => {
  it('keeps what applies and drops what takes nothing off', () => {
    expect(effectiveDiscount(percent(1_000), 500)).toEqual(percent(1_000))
    expect(effectiveDiscount(amount(100), 500)).toEqual(amount(100))
    expect(effectiveDiscount(amount(900), 500)).toEqual(amount(500))
    expect(effectiveDiscount(percent(20_000), 500)).toEqual(percent(PERCENT_SCALE))
    expect(effectiveDiscount(amount(100), 0)).toBeNull()
    expect(effectiveDiscount(null, 500)).toBeNull()
    expect(effectiveDiscount(undefined, 500)).toBeNull()
  })
})

// The backend's TestProrate, plus ties and remainders.
describe('prorate', () => {
  it.each([
    { name: 'nothing to spread', total: 0, weights: [5, 7], want: [0, 0] },
    { name: 'no weight', total: 0, weights: [0, 0], want: [0, 0] },
    { name: 'weights that add up to zero', total: 3, weights: [0, 0], want: [0, 0] },
    { name: 'no lines', total: 0, weights: [], want: [] },
    { name: 'exact shares', total: 30, weights: [200, 100], want: [20, 10] },
    { name: 'cents go to the largest fractions', total: 6, weights: [3, 3, 1], want: [3, 2, 1] },
    { name: 'the first line wins a tie', total: 1, weights: [1, 1], want: [1, 0] },
    { name: 'a remainder spread over equal weights', total: 100, weights: [1, 1, 1], want: [34, 33, 33] },
    { name: 'the largest remainder wins over the first line', total: 5, weights: [1, 2], want: [2, 3] },
    { name: 'a line without amount gets nothing', total: 5, weights: [5, 0], want: [5, 0] },
    {
      name: 'beyond 2^53 before dividing',
      total: 10_000_000_000_000,
      weights: [5_000_000_000_000, 5_000_000_000_000],
      want: [5_000_000_000_000, 5_000_000_000_000],
    },
  ])('$name', ({ total, weights, want }) => {
    expect(prorate(total, weights)).toEqual(want)
  })
})

describe('priceCart', () => {
  it('prices a cart without discounts as the sum of the lines', () => {
    const priced = priceCart([cola, { ...bread, discount: null }], null)

    expect(priced).toMatchObject({ subtotal: 850_000, cartDiscount: null, cartDiscountAmount: 0, discountTotal: 0, total: 850_000 })
    expect(priced.lines.map((line) => line.lineTotal)).toEqual([640_000, 210_000])
    expect(priceCart([], null)).toEqual({ lines: [], subtotal: 0, cartDiscount: null, cartDiscountAmount: 0, discountTotal: 0, total: 0 })
  })

  it('takes a line discount off its own line', () => {
    const priced = priceCart([cola, bread], null)

    expect(priced.lines[1]).toMatchObject({ amount: 210_000, discount: percent(1_000), discountAmount: 21_000, lineTotal: 189_000 })
    expect(priced).toMatchObject({ subtotal: 850_000, discountTotal: 21_000, total: 829_000 })
  })

  it.each([
    { name: 'a percent on what is left after the line discounts', discount: percent(1_000), shares: [64_000, 18_900], off: 82_900 },
    { name: 'a fixed amount spread by weight', discount: amount(100_000), shares: [77_201, 22_799], off: 100_000 },
    { name: 'everything', discount: amount(829_000), shares: [640_000, 189_000], off: 829_000 },
    { name: 'nothing', discount: amount(0), shares: [0, 0], off: 0 },
  ])('spreads a general discount: $name', ({ discount, shares, off }) => {
    const priced = priceCart([cola, bread], discount)

    expect(priced.lines.map((line) => line.cartDiscountAmount)).toEqual(shares)
    expect(priced.cartDiscountAmount).toBe(off)
    expect(priced.discountTotal).toBe(21_000 + off)
    expect(priced.total).toBe(829_000 - off)
    expect(priced.lines.reduce((sum, line) => sum + line.lineTotal, 0)).toBe(priced.total)
  })

  it('caps discounts at what they are taken from, so the request is never refused for them', () => {
    const priced = priceCart([{ ...cola, discount: amount(9_999_999) }, bread], amount(9_999_999))

    expect(priced.lines[0]).toMatchObject({ discount: amount(640_000), discountAmount: 640_000, lineTotal: 0 })
    expect(priced.cartDiscount).toEqual(amount(189_000))
    expect(priced.total).toBe(0)
    expect(priced.lines.every((line) => line.lineTotal >= 0)).toBe(true)
  })

  it('takes nothing off with a general percent when nothing is left', () => {
    const priced = priceCart([{ ...cola, discount: percent(PERCENT_SCALE) }], percent(1_000))

    expect(priced.cartDiscountAmount).toBe(0)
    expect(priced.total).toBe(0)
  })

  it('never leaves a line negative and keeps the total the sum of the lines', () => {
    const lines: CartLine[] = [
      { ...cola, unitPrice: 333, quantity: 3_000 },
      { ...bread, unitPrice: 101, quantity: 1_500, discount: percent(3_333) },
      { ...cola, itemId: 'x', unitPrice: 7, quantity: 1_000 },
    ]
    const priced = priceCart(lines, percent(1_250))

    expect(priced.lines.every((line) => line.lineTotal >= 0)).toBe(true)
    expect(priced.lines.reduce((sum, line) => sum + line.lineTotal, 0)).toBe(priced.total)
    expect(priced.lines.reduce((sum, line) => sum + line.cartDiscountAmount, 0)).toBe(priced.cartDiscountAmount)
  })

  it('carries the quantity and item of each line to the request', () => {
    expect(priceCart([cola], null).lines[0]).toMatchObject({ itemId: 'coca', quantity: 2_000 })
  })
})

describe('formatDiscount', () => {
  it('reads a percent and an amount', () => {
    expect(formatDiscount(percent(1_250))).toBe('12,5 %')
    expect(formatDiscount(amount(50_000)).replace(/\s/g, ' ')).toBe('$ 500,00')
  })
})
