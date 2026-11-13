// Argentine identity numbers (DNI, CUIT, CUIL). The backend stores them as
// digits only; these helpers read what a person types and write them back the
// way people are used to seeing them.

// Like the backend, separators don't count: "20-12345678-9" and
// "20123456789" are the same CUIT.
export function documentDigits(value: string): string {
  return value.trim().replace(/[\s./-]/g, '')
}

// 28456789 → 28.456.789. Anything that isn't a 7 or 8 digit DNI is left as is.
export function formatDni(digits: string): string {
  return /^\d{7,8}$/.test(digits) ? digits.replace(/\B(?=(\d{3})+$)/g, '.') : digits
}

// 30712345678 → 30-71234567-8, for CUIT and CUIL. Anything that isn't 11
// digits is left as is.
export function formatCuit(digits: string): string {
  return /^\d{11}$/.test(digits) ? `${digits.slice(0, 2)}-${digits.slice(2, 10)}-${digits.slice(10)}` : digits
}
