// Quantities travel as integer thousandths of the unit, like the backend's
// domain.Quantity: 0,250 kg is 250 and three pieces are 3000.

export const QUANTITY_SCALE = 1000

export function formatQuantity(thousandths: number): string {
  const whole = Math.trunc(thousandths / QUANTITY_SCALE)
  const fraction = Math.abs(thousandths % QUANTITY_SCALE)
  if (fraction === 0) return String(whole)
  return `${whole},${String(fraction).padStart(3, '0').replace(/0+$/, '')}`
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
