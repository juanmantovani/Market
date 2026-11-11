// Money travels as integer cents, like the backend's domain.Money. These
// helpers are the only place that turns cents into text and back.

const currency = new Intl.NumberFormat('es-AR', { style: 'currency', currency: 'ARS' })

export function formatMoney(cents: number): string {
  return currency.format(cents / 100)
}

// The value to show in an editable field: no thousands separator, comma
// for decimals ("1234,50").
export function centsToInput(cents: number | null): string {
  if (cents === null) return ''
  const whole = Math.trunc(cents / 100)
  const fraction = Math.abs(cents % 100)
  return fraction === 0 ? String(whole) : `${whole},${String(fraction).padStart(2, '0')}`
}

// Reads what a person types in Argentina: "1.234,50", "1234,5", "1234" or
// "$ 1.234". A dot followed by one or two digits at the end ("1234.50") is
// read as a decimal point too. Returns null for anything else.
export function parseMoney(input: string): number | null {
  const text = input.replace(/[\s$]/g, '')
  if (!text) return null

  let normalized: string
  if (text.includes(',')) {
    normalized = text.replace(/\./g, '').replace(',', '.')
  } else if (/^\d+\.\d{1,2}$/.test(text)) {
    normalized = text
  } else {
    normalized = text.replace(/\./g, '')
  }

  if (!/^\d+(\.\d{1,2})?$/.test(normalized)) return null

  const [whole, fraction = ''] = normalized.split('.')
  return Number(whole) * 100 + Number(fraction.padEnd(2, '0'))
}
