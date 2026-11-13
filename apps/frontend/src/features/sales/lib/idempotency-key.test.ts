import { afterEach, describe, expect, it, vi } from 'vitest'
import { newIdempotencyKey } from './idempotency-key'

afterEach(() => vi.unstubAllGlobals())

describe('newIdempotencyKey', () => {
  it('is a different value each time', () => {
    expect(newIdempotencyKey()).not.toBe(newIdempotencyKey())
  })

  it('still works where crypto.randomUUID is not available', () => {
    vi.stubGlobal('crypto', { getRandomValues: (bytes: Uint8Array) => bytes.fill(171) })
    expect(newIdempotencyKey()).toBe('ab'.repeat(16))
  })
})
