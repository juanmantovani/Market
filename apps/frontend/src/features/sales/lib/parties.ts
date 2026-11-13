import { formatCuit, formatDni } from '../../../shared/lib/document'
import type { CustomerOption, SellerOption } from '../types'

const DOCUMENT_LABELS = { dni: 'DNI', cuit: 'CUIT', cuil: 'CUIL' } as const

// "DNI 28.456.789", or null for a customer without document.
export function documentText({ documentType, documentNumber }: Pick<CustomerOption, 'documentType' | 'documentNumber'>): string | null {
  if (documentType === 'none') return null
  const number = documentType === 'dni' ? formatDni(documentNumber) : formatCuit(documentNumber)
  return `${DOCUMENT_LABELS[documentType]} ${number}`
}

// The seller whose code was typed, so "jp" + Enter picks Juan Pérez even
// when other names contain those letters.
export function sellerByCode(term: string, sellers: readonly SellerOption[]): SellerOption | null {
  const code = term.trim().toUpperCase()
  return sellers.find((seller) => seller.code === code) ?? null
}
