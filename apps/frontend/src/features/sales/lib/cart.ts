import { QUANTITY_SCALE } from '../../../shared/lib/quantity'
import type { CartLine, CatalogItem } from '../types'

// The backend caps a line at a million units (domain.MaxStockQuantity).
export const MAX_LINE_QUANTITY = 1_000_000 * QUANTITY_SCALE

// a × b ÷ divisor rounded half away from zero, like the backend's
// mulDivRound. BigInt keeps the product exact beyond 2^53.
function mulDivRound(a: number, b: number, divisor: bigint): number {
  const product = BigInt(a) * BigInt(b)
  const negative = product < 0n
  const absolute = negative ? -product : product
  let quotient = absolute / divisor
  if ((absolute % divisor) * 2n >= divisor) quotient += 1n
  return Number(negative ? -quotient : quotient)
}

// Price × quantity rounded to the cent, as the backend prices a line
// (domain.Money.Times). The cart only previews it: the backend prices the
// sale again.
export function lineAmount(line: Pick<CartLine, 'unitPrice' | 'quantity'>): number {
  return mulDivRound(line.unitPrice, line.quantity, BigInt(QUANTITY_SCALE))
}

// The sum of the lines, so the total always matches what they show.
export function cartTotal(lines: readonly CartLine[]): number {
  return lines.reduce((total, line) => total + lineAmount(line), 0)
}

// What + and − change: a unit, or 100 g (or ml, or cm) for items sold by
// weight or length.
export function quantityStep(unit: CartLine['unit']): number {
  return unit === 'unit' ? QUANTITY_SCALE : QUANTITY_SCALE / 10
}

// Adds quantity of item: to its line if it's already in the cart (one line
// per item, as the backend requires), at the end otherwise.
export function addItem(lines: readonly CartLine[], item: CatalogItem, quantity: number): CartLine[] {
  const existing = lines.find((line) => line.itemId === item.id)
  if (existing) {
    return lines.map((line) =>
      line.itemId === item.id ? { ...line, quantity: Math.min(line.quantity + quantity, MAX_LINE_QUANTITY) } : line,
    )
  }
  return [
    ...lines,
    {
      itemId: item.id,
      code: item.code,
      name: item.name,
      unit: item.unit,
      unitPrice: item.salePrice,
      quantity: Math.min(quantity, MAX_LINE_QUANTITY),
    },
  ]
}

export function isValidQuantity(quantity: number): boolean {
  return Number.isInteger(quantity) && quantity > 0 && quantity <= MAX_LINE_QUANTITY
}

export function setQuantity(lines: readonly CartLine[], itemId: string, quantity: number): CartLine[] {
  if (!isValidQuantity(quantity)) return [...lines]
  return lines.map((line) => (line.itemId === itemId ? { ...line, quantity } : line))
}

export type RemovedLine = { line: CartLine; index: number }

export function removeLine(lines: readonly CartLine[], itemId: string): { lines: CartLine[]; removed: RemovedLine | null } {
  const index = lines.findIndex((line) => line.itemId === itemId)
  if (index === -1) return { lines: [...lines], removed: null }
  return { lines: lines.filter((_, i) => i !== index), removed: { line: lines[index], index } }
}

// Puts a removed line back where it was, unless the item was added again
// meanwhile: then its quantities add up.
export function restoreLine(lines: readonly CartLine[], { line, index }: RemovedLine): CartLine[] {
  const existing = lines.find((current) => current.itemId === line.itemId)
  if (existing) {
    return lines.map((current) =>
      current.itemId === line.itemId
        ? { ...current, quantity: Math.min(current.quantity + line.quantity, MAX_LINE_QUANTITY) }
        : current,
    )
  }
  const restored = [...lines]
  restored.splice(Math.min(index, restored.length), 0, line)
  return restored
}

// The item a code read by the scanner (or typed) names among the search
// results: an exact code or barcode first, the only result otherwise.
export function pickByCode<T extends Pick<CatalogItem, 'code' | 'barcode'>>(term: string, results: readonly T[]): T | null {
  const wanted = term.trim().toLowerCase()
  if (!wanted) return null
  const exact = results.find((item) => item.code.toLowerCase() === wanted || item.barcode?.toLowerCase() === wanted)
  if (exact) return exact
  return results.length === 1 ? results[0] : null
}
