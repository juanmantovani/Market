// Quantities travel as integer thousandths of the unit, like the backend's
// domain.Quantity: 0,250 kg is 250 and three pieces are 3000.

export const QUANTITY_SCALE = 1000

export type Unit = 'unit' | 'kg' | 'l' | 'm'

const UNIT_SHORT: Record<Unit, string> = { unit: 'u.', kg: 'kg', l: 'l', m: 'm' }

export function formatQuantity(thousandths: number): string {
  const sign = thousandths < 0 ? '-' : ''
  const absolute = Math.abs(thousandths)
  const whole = Math.trunc(absolute / QUANTITY_SCALE)
  const fraction = absolute % QUANTITY_SCALE
  if (fraction === 0) return `${sign}${whole}`
  return `${sign}${whole},${String(fraction).padStart(3, '0').replace(/0+$/, '')}`
}

// "4 u.", "1,8 kg": a stock or a movement as people read it.
export function formatUnitQuantity(thousandths: number, unit: Unit): string {
  return `${formatQuantity(thousandths)} ${UNIT_SHORT[unit]}`
}

export function unitShort(unit: Unit): string {
  return UNIT_SHORT[unit]
}

// Mirrors the backend's domain.IsLowStock: an item without a minimum never
// counts as low.
export function isLowStock(quantity: number, minStock: number): boolean {
  return minStock > 0 && quantity <= minStock
}

// Reads "2", "2,5" or "2.5". Fractions (up to three decimals) are only
// accepted when allowFraction is set, as for items sold by weight.
export function parseQuantity(input: string, allowFraction: boolean): number | null {
  const text = input.trim().replace(',', '.')
  if (!/^\d+(\.\d{1,3})?$/.test(text)) return null

  const [whole, fraction = ''] = text.split('.')
  if (fraction && Number(fraction) !== 0 && !allowFraction) return null

  return Number(whole) * QUANTITY_SCALE + Number(fraction.padEnd(3, '0'))
}
