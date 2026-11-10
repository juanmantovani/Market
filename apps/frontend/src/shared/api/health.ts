import { z } from 'zod'
import { apiFetch } from './client'

const healthSchema = z.object({ status: z.literal('ok') })

export type Health = z.infer<typeof healthSchema>

export function fetchHealth(signal?: AbortSignal): Promise<Health> {
  return apiFetch('/health/live', healthSchema, { signal })
}
