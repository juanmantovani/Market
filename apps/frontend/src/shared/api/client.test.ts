import { afterEach, describe, expect, it, vi } from 'vitest'
import { z } from 'zod'
import { ApiError, apiFetch, messageFromError } from './client'

const itemSchema = z.object({ id: z.string(), name: z.string() })

function stubFetch(response: Response | Error | DOMException) {
  const fetchMock = vi.fn(() =>
    response instanceof Response ? Promise.resolve(response) : Promise.reject(response),
  )
  vi.stubGlobal('fetch', fetchMock)
  return fetchMock
}

function jsonResponse(body: unknown, status = 200) {
  return new Response(JSON.stringify(body), {
    status,
    headers: { 'Content-Type': 'application/json' },
  })
}

async function rejection(promise: Promise<unknown>): Promise<ApiError> {
  try {
    await promise
  } catch (error) {
    if (error instanceof ApiError) return error
    throw error
  }
  throw new Error('expected the request to fail')
}

afterEach(() => {
  vi.unstubAllGlobals()
})

describe('apiFetch', () => {
  it('returns the response validated against the schema', async () => {
    const fetchMock = stubFetch(jsonResponse({ id: 'i1', name: 'Yerba' }))

    await expect(apiFetch('/api/v1/items/i1', itemSchema)).resolves.toEqual({
      id: 'i1',
      name: 'Yerba',
    })
    expect(fetchMock).toHaveBeenCalledWith(
      '/api/v1/items/i1',
      expect.objectContaining({ method: 'GET' }),
    )
  })

  it('sends a JSON body with its content type', async () => {
    const fetchMock = stubFetch(jsonResponse({ id: 'i2', name: 'Café' }, 201))

    await apiFetch('/api/v1/items', itemSchema, {
      method: 'POST',
      body: { name: 'Café' },
      headers: { 'Idempotency-Key': 'k1' },
    })

    const [, init] = fetchMock.mock.calls[0] as unknown as [string, RequestInit]
    expect(init.body).toBe('{"name":"Café"}')
    expect(init.headers).toMatchObject({
      'Content-Type': 'application/json',
      'Idempotency-Key': 'k1',
    })
  })

  it('accepts an empty body when the schema allows it', async () => {
    stubFetch(new Response(null, { status: 204 }))

    await expect(apiFetch('/api/v1/items/i1', z.undefined(), { method: 'DELETE' })).resolves.toBe(
      undefined,
    )
  })

  it('rejects a response that does not match the schema', async () => {
    stubFetch(jsonResponse({ id: 1 }))

    const error = await rejection(apiFetch('/api/v1/items/i1', itemSchema))

    expect(error.code).toBe('invalid_response')
  })

  it('rejects a body that is not JSON', async () => {
    stubFetch(new Response('<html>', { status: 200 }))

    const error = await rejection(apiFetch('/api/v1/items/i1', itemSchema))

    expect(error.code).toBe('invalid_response')
  })

  it('surfaces the backend code and message of an error', async () => {
    stubFetch(jsonResponse({ code: 'insufficient_stock', message: 'No hay stock suficiente' }, 409))

    const error = await rejection(apiFetch('/api/v1/sales', itemSchema, { method: 'POST' }))

    expect(error).toMatchObject({
      code: 'insufficient_stock',
      message: 'No hay stock suficiente',
      status: 409,
    })
  })

  it('falls back to a status-based message when the error is not JSON', async () => {
    stubFetch(new Response('Bad Gateway', { status: 502 }))

    const error = await rejection(apiFetch('/api/v1/items', itemSchema))

    expect(error.code).toBe('http_error')
    expect(error.status).toBe(502)
    expect(error.message).toMatch(/error inesperado/)
  })

  it('uses a generic message for a client error without details', async () => {
    stubFetch(jsonResponse({}, 400))

    const error = await rejection(apiFetch('/api/v1/items', itemSchema))

    expect(error.message).toBe('La solicitud no pudo completarse.')
  })

  it('reports a network failure without a status', async () => {
    stubFetch(new TypeError('Failed to fetch'))

    const error = await rejection(apiFetch('/api/v1/items', itemSchema))

    expect(error).toMatchObject({ code: 'network_error', status: 0 })
  })

  it('lets an aborted request propagate as is', async () => {
    stubFetch(new DOMException('Aborted', 'AbortError'))

    await expect(apiFetch('/api/v1/items', itemSchema)).rejects.toMatchObject({ name: 'AbortError' })
  })
})

describe('messageFromError', () => {
  it('shows the message of an ApiError', () => {
    expect(messageFromError(new ApiError('Sin stock', 'insufficient_stock', 409))).toBe('Sin stock')
  })

  it('hides anything else behind a generic message', () => {
    expect(messageFromError(new Error('stack trace'))).toMatch(/error inesperado/)
  })
})
