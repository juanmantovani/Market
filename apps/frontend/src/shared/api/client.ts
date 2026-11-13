import type { ZodType } from 'zod'
import { apiBaseUrl } from '../config/env'

const NETWORK_ERROR_MESSAGE =
  'No se pudo conectar con el servidor. Revisar la conexión e intentar de nuevo.'
const UNEXPECTED_ERROR_MESSAGE = 'Ocurrió un error inesperado. Intentar de nuevo en unos minutos.'
const REQUEST_FAILED_MESSAGE = 'La solicitud no pudo completarse.'

// ApiError carries the backend's own `code` and `message` (see
// response.ErrorResponse in the Go API), so callers can show the message as
// is and branch on the code. `status` is 0 when the request never reached
// the server.
export class ApiError extends Error {
  readonly code: string
  readonly status: number

  constructor(message: string, code: string, status: number) {
    super(message)
    this.name = 'ApiError'
    this.code = code
    this.status = status
  }
}

export type RequestOptions = {
  method?: string
  body?: unknown
  signal?: AbortSignal
  headers?: Record<string, string>
}

// apiFetch requests path and validates the JSON response against schema, so
// a contract drift between frontend and backend fails loudly here instead of
// rendering undefined values.
export async function apiFetch<T>(
  path: string,
  schema: ZodType<T>,
  { method = 'GET', body, signal, headers: extraHeaders }: RequestOptions = {},
): Promise<T> {
  const headers: Record<string, string> = { Accept: 'application/json', ...extraHeaders }
  let payload: string | undefined
  if (body !== undefined) {
    payload = JSON.stringify(body)
    headers['Content-Type'] = 'application/json'
  }

  let response: Response
  try {
    response = await fetch(`${apiBaseUrl}${path}`, { method, headers, body: payload, signal })
  } catch (error) {
    if (error instanceof DOMException && error.name === 'AbortError') {
      throw error
    }
    throw new ApiError(NETWORK_ERROR_MESSAGE, 'network_error', 0)
  }

  if (!response.ok) {
    throw await readError(response)
  }

  const text = await response.text()
  let data: unknown
  try {
    data = text ? JSON.parse(text) : undefined
  } catch {
    throw new ApiError(UNEXPECTED_ERROR_MESSAGE, 'invalid_response', response.status)
  }

  const parsed = schema.safeParse(data)
  if (!parsed.success) {
    throw new ApiError(UNEXPECTED_ERROR_MESSAGE, 'invalid_response', response.status)
  }

  return parsed.data
}

async function readError(response: Response): Promise<ApiError> {
  let code = 'http_error'
  let message = response.status >= 500 ? UNEXPECTED_ERROR_MESSAGE : REQUEST_FAILED_MESSAGE

  try {
    const data: unknown = await response.json()
    if (data && typeof data === 'object') {
      const shape = data as { code?: unknown; message?: unknown }
      if (typeof shape.code === 'string') code = shape.code
      if (typeof shape.message === 'string' && shape.message !== '') message = shape.message
    }
  } catch {
    // Not JSON (e.g. a proxy error page): the status-based message stands.
  }

  return new ApiError(message, code, response.status)
}

export function messageFromError(error: unknown): string {
  return error instanceof ApiError ? error.message : UNEXPECTED_ERROR_MESSAGE
}
