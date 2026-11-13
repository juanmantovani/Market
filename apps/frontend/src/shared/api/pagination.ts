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

// The API serves at most this many items per page.
export const MAX_PAGE_SIZE = 100

// Every item of a list, asking page after page until the total is reached:
// for pickers and filters that need the whole list, not a page of it.
export async function listAllPages<T>(fetchPage: (page: number, limit: number) => Promise<Page<T>>): Promise<T[]> {
  const items: T[] = []
  for (let page = 1; ; page++) {
    const result = await fetchPage(page, MAX_PAGE_SIZE)
    items.push(...result.items)
    if (result.items.length === 0 || items.length >= result.total) return items
  }
}
