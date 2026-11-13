import { z } from 'zod'
import type { Page } from '../../shared/api/pagination'
import type { Unit } from '../../shared/lib/quantity'

// What the counter needs to know about an item to sell it. The catalog
// comes from the items API, composed in app/PosRoute.
export type CatalogItem = {
  id: string
  code: string
  barcode: string | null
  name: string
  unit: Unit
  salePrice: number
  categoryId: string
}

export type CatalogParams = {
  search: string
  categoryId: string
  page: number
  limit: number
}

export type SearchCatalog = (params: CatalogParams, signal?: AbortSignal) => Promise<Page<CatalogItem>>

export type CategoryOption = { id: string; name: string }

const unitSchema = z.enum(['unit', 'kg', 'l', 'm'])

// A cart line keeps a copy of the item as it was when it was added: the
// backend prices the sale again and refuses it if the total changed.
export const cartLineSchema = z.object({
  itemId: z.string().min(1),
  code: z.string(),
  name: z.string(),
  unit: unitSchema,
  unitPrice: z.number().int().positive(),
  quantity: z.number().int().positive(),
})

export type CartLine = z.infer<typeof cartLineSchema>

export const storedCartSchema = z.object({
  version: z.literal(1),
  lines: z.array(cartLineSchema),
})
