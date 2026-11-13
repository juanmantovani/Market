import { parseMoney } from '../../../shared/lib/money'
import type { PaymentMethod, PaymentRequest } from '../types'

export const PAYMENT_METHODS: readonly PaymentMethod[] = ['cash', 'debit_card', 'credit_card', 'transfer', 'qr']

export const PAYMENT_METHOD_LABELS: Record<PaymentMethod, string> = {
  cash: 'Efectivo',
  debit_card: 'Tarjeta de débito',
  credit_card: 'Tarjeta de crédito',
  transfer: 'Transferencia',
  qr: 'QR',
}

// The backend accepts up to ten payments in a sale (domain.MaxSalePayments).
export const MAX_PAYMENTS = 10
export const MAX_REFERENCE_LENGTH = 60

// A payment as it is being typed: the amount is still text.
export type PaymentDraft = { id: string; method: PaymentMethod; amount: string; reference: string }

export type PaymentSummary = {
  payments: PaymentRequest[]
  // Rows whose amount is filled in but isn't a valid amount.
  invalidIds: string[]
  paid: number
  remaining: number
  change: number
  // Only cash is given back: the other methods can't add up to more than the
  // total.
  nonCashExceedsTotal: boolean
  canConfirm: boolean
}

// Reads the drafts the way the backend will: empty rows are ignored, the sum
// has to cover the total, and the change is what cash adds beyond it.
export function summarizePayments(total: number, drafts: readonly PaymentDraft[]): PaymentSummary {
  const payments: PaymentRequest[] = []
  const invalidIds: string[] = []
  let paid = 0
  let nonCash = 0

  for (const draft of drafts) {
    if (!draft.amount.trim()) continue
    const amount = parseMoney(draft.amount)
    if (amount === null || amount <= 0) {
      invalidIds.push(draft.id)
      continue
    }
    const reference = draft.method === 'cash' ? '' : draft.reference.trim()
    payments.push({ method: draft.method, amount, reference: reference || null })
    paid += amount
    if (draft.method !== 'cash') nonCash += amount
  }

  const nonCashExceedsTotal = nonCash > total
  const change = nonCashExceedsTotal ? 0 : Math.max(0, paid - total)

  return {
    payments,
    invalidIds,
    paid,
    remaining: Math.max(0, total - paid),
    change,
    nonCashExceedsTotal,
    canConfirm: total > 0 && payments.length > 0 && invalidIds.length === 0 && paid >= total && !nonCashExceedsTotal,
  }
}

// What is left to pay once every row but this one is counted.
export function amountDueExcluding(total: number, drafts: readonly PaymentDraft[], excludedId: string): number {
  const others = summarizePayments(
    total,
    drafts.filter((draft) => draft.id !== excludedId),
  )
  return others.remaining
}
