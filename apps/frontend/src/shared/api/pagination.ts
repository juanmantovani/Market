import { z, type ZodType } from 'zod'

export type ListParams = {
  search: string
  page: number
  limit: number
}

export type Page<T> = {
  items: T[]
  total: number
  page: number
  limit: number
}

export function pageSchema<T>(item: ZodType<T>): ZodType<Page<T>> {
  return z.object({
    items: z.array(item),
    total: z.number().int().nonnegative(),
    page: z.number().int().positive(),
    limit: z.number().int().positive(),
  })
}

export function listQuery({ search, page, limit }: ListParams): string {
  const query = new URLSearchParams()
  if (search.trim()) query.set('q', search.trim())
  query.set('page', String(page))
  query.set('limit', String(limit))
  return query.toString()
}
