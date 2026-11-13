import { describe, expect, it } from 'vitest'
import { amountDueExcluding, summarizePayments, type PaymentDraft } from './payments'

const draft = (overrides: Partial<PaymentDraft> & { id: string }): PaymentDraft => ({
  method: 'cash',
  amount: '',
  reference: '',
  ...overrides,
})

describe('summarizePayments', () => {
  it('accepts an exact payment', () => {
    const summary = summarizePayments(150_000, [draft({ id: 'a', amount: '1500' })])
    expect(summary).toMatchObject({ paid: 150_000, remaining: 0, change: 0, canConfirm: true })
    expect(summary.payments).toEqual([{ method: 'cash', amount: 150_000, reference: null }])
  })

  it('gives back what cash adds beyond the total', () => {
    const summary = summarizePayments(150_000, [draft({ id: 'a', amount: '2.000' })])
    expect(summary).toMatchObject({ change: 50_000, canConfirm: true })
  })

  it('reports what is missing and blocks the sale', () => {
    const summary = summarizePayments(150_000, [draft({ id: 'a', amount: '1000,50' })])
    expect(summary).toMatchObject({ remaining: 49_950, change: 0, canConfirm: false })
  })

  it('splits a sale between methods and only keeps the reference of non-cash payments', () => {
    const summary = summarizePayments(150_000, [
      draft({ id: 'a', amount: '1000', reference: 'ignored' }),
      draft({ id: 'b', method: 'transfer', amount: '500', reference: '  OP-77 ' }),
    ])
    expect(summary.canConfirm).toBe(true)
    expect(summary.payments).toEqual([
      { method: 'cash', amount: 100_000, reference: null },
      { method: 'transfer', amount: 50_000, reference: 'OP-77' },
    ])
  })

  it('refuses non-cash payments that add up to more than the total', () => {
    const summary = summarizePayments(150_000, [draft({ id: 'a', method: 'qr', amount: '1600' })])
    expect(summary).toMatchObject({ nonCashExceedsTotal: true, change: 0, canConfirm: false })
  })

  it('ignores empty rows and flags amounts that are not valid', () => {
    const summary = summarizePayments(150_000, [
      draft({ id: 'a', amount: '1500' }),
      draft({ id: 'b', method: 'qr', amount: '   ' }),
      draft({ id: 'c', method: 'qr', amount: 'abc' }),
      draft({ id: 'd', method: 'qr', amount: '0' }),
    ])
    expect(summary.invalidIds).toEqual(['c', 'd'])
    expect(summary.canConfirm).toBe(false)
  })

  it('cannot confirm with nothing typed', () => {
    expect(summarizePayments(150_000, [draft({ id: 'a' })]).canConfirm).toBe(false)
  })
})

describe('amountDueExcluding', () => {
  it('is what the other rows leave to pay', () => {
    const drafts = [draft({ id: 'a', amount: '1000' }), draft({ id: 'b', method: 'qr', amount: '' })]
    expect(amountDueExcluding(150_000, drafts, 'b')).toBe(50_000)
    expect(amountDueExcluding(150_000, drafts, 'a')).toBe(150_000)
  })
})
