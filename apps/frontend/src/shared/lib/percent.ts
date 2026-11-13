// Percentages travel as integer hundredths of a percent, like the backend's
// domain.Percent (2,5 % is 250). These helpers are the only place that turns
// them into text and back.

const PERCENT_SCALE = 100

const display = new Intl.NumberFormat('es-AR', { maximumFractionDigits: 2 })

// 250 → "2,5 %"
export function formatPercent(hundredths: number): string {
  return `${display.format(hundredths / PERCENT_SCALE)} %`
}

// The value to show in an editable field: 250 → "2,5".
export function percentToInput(hundredths: number | null): string {
  if (hundredths === null) return ''
  const whole = Math.trunc(hundredths / PERCENT_SCALE)
  const fraction = hundredths % PERCENT_SCALE
  if (fraction === 0) return String(whole)
  return `${whole},${String(fraction).padStart(2, '0').replace(/0$/, '')}`
}

// Reads "2,5", "2.5", "12" or "3 %" with up to two decimals. Returns null for
// anything else; checking the range is up to the caller.
export function parsePercent(input: string): number | null {
  const text = input.replace(/[\s%]/g, '').replace(',', '.')
  if (!/^\d{1,3}(\.\d{1,2})?$/.test(text)) return null

  const [whole, fraction = ''] = text.split('.')
  return Number(whole) * PERCENT_SCALE + Number(fraction.padEnd(2, '0'))
}
