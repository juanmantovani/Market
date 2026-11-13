// crypto.randomUUID only exists in secure contexts, and a counter on the
// local network is served over plain http.
export function newIdempotencyKey(): string {
  const cryptoApi = globalThis.crypto
  if (typeof cryptoApi.randomUUID === 'function') return cryptoApi.randomUUID()
  const bytes = cryptoApi.getRandomValues(new Uint8Array(16))
  return Array.from(bytes, (byte) => byte.toString(16).padStart(2, '0')).join('')
}
