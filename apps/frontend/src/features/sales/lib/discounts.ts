import { formatMoney } from '../../../shared/lib/money'
import { formatPercent } from '../../../shared/lib/percent'
import type { CartLine, Discount } from '../types'
import { lineAmount, mulDivRound } from './cart'

// 100 % in hundredths of a point, like the backend's domain.PercentScale.
export const PERCENT_SCALE = 10_000

// What a discount takes off an amount, like the backend's amountOff: a
// percent rounded to the cent, or the fixed amount. An amount can't be more
// than what it is taken from, so it is capped.
export function discountOff(discount: Discount | null | undefined, amount: number): number {
  if (!discount) return 0
  if (discount.type === 'percent') return mulDivRound(amount, Math.min(discount.value, PERCENT_SCALE), BigInt(PERCENT_SCALE))
  return Math.min(discount.value, amount)
}

// The discount that will really be applied (and sent) to an amount: an
// amount discount is capped at it, and one that takes nothing off is none.
export function effectiveDiscount(discount: Discount | null | undefined, amount: number): Discount | null {
  if (!discount) return null
  if (discount.type === 'percent') {
    const value = Math.min(discount.value, PERCENT_SCALE)
    return value > 0 ? { type: 'percent', value } : null
  }
  const value = Math.min(discount.value, amount)
  return value > 0 ? { type: 'amount', value } : null
}

// Splits total among weights in proportion to each one by largest
// remainder, like the backend's prorate: every share is its exact share
// rounded down, and the cents left go one each to the largest fractions,
// the first line winning a tie. The shares add up to total.
export function prorate(total: number, weights: readonly number[]): number[] {
  const shares = weights.map(() => 0)
  const sum = weights.reduce((acc, weight) => acc + BigInt(weight), 0n)
  if (total === 0 || sum === 0n) return shares

  const remainders: bigint[] = []
  let left = total
  weights.forEach((weight, index) => {
    const product = BigInt(total) * BigInt(weight)
    shares[index] = Number(product / sum)
    remainders[index] = product % sum
    left -= shares[index]
  })

  const order = weights.map((_, index) => index).sort((a, b) => (remainders[b] > remainders[a] ? 1 : remainders[b] < remainders[a] ? -1 : 0))
  for (const index of order.slice(0, left)) shares[index] += 1
  return shares
}

export type PricedLine = {
  itemId: string
  quantity: number
  // Price × quantity, before any discount.
  amount: number
  discount: Discount | null
  discountAmount: number
  // This line's share of the general discount.
  cartDiscountAmount: number
  lineTotal: number
}

export type PricedCart = {
  lines: PricedLine[]
  subtotal: number
  cartDiscount: Discount | null
  cartDiscountAmount: number
  discountTotal: number
  total: number
}

// Prices the cart the way the backend prices a sale (domain.NewSale): each
// line's amount is rounded to the cent and its own discount taken off; the
// general discount applies to what the lines add up to after theirs and is
// spread over them in proportion. The total is the sum of the lines. The
// cart only previews it: the backend prices the sale again and refuses it
// when the total differs.
export function priceCart(lines: readonly CartLine[], cartDiscount: Discount | null | undefined): PricedCart {
  const priced = lines.map((line): PricedLine => {
    const amount = lineAmount(line)
    const discount = effectiveDiscount(line.discount, amount)
    const discountAmount = discountOff(discount, amount)
    return { itemId: line.itemId, quantity: line.quantity, amount, discount, discountAmount, cartDiscountAmount: 0, lineTotal: amount - discountAmount }
  })

  const afterLineDiscounts = priced.reduce((sum, line) => sum + line.lineTotal, 0)
  const effectiveCartDiscount = effectiveDiscount(cartDiscount, afterLineDiscounts)
  const cartDiscountAmount = discountOff(effectiveCartDiscount, afterLineDiscounts)
  const shares = prorate(
    cartDiscountAmount,
    priced.map((line) => line.lineTotal),
  )
  priced.forEach((line, index) => {
    line.cartDiscountAmount = shares[index]
    line.lineTotal -= shares[index]
  })

  const subtotal = priced.reduce((sum, line) => sum + line.amount, 0)
  const lineDiscounts = priced.reduce((sum, line) => sum + line.discountAmount, 0)
  const discountTotal = lineDiscounts + cartDiscountAmount
  return {
    lines: priced,
    subtotal,
    cartDiscount: effectiveCartDiscount,
    cartDiscountAmount,
    discountTotal,
    total: subtotal - discountTotal,
  }
}

// "10 %" or "$ 500,00".
export function formatDiscount(discount: Discount): string {
  return discount.type === 'percent' ? formatPercent(discount.value) : formatMoney(discount.value)
}
