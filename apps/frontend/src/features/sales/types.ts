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

export type PartyParams = { search: string; page: number; limit: number }

const documentTypeSchema = z.enum(['dni', 'cuit', 'cuil', 'none'])

// Who the sale is made to. `system` marks Consumidor Final, the customer of
// every sale without one. Composed from the customers API in app/PosRoute.
export const customerOptionSchema = z.object({
  id: z.string().min(1),
  name: z.string(),
  documentType: documentTypeSchema,
  documentNumber: z.string(),
  system: z.boolean(),
})

export type CustomerOption = z.infer<typeof customerOptionSchema>

export type SearchCustomers = (params: PartyParams, signal?: AbortSignal) => Promise<Page<CustomerOption>>

// Who attends the sale. Composed from the sellers API in app/PosRoute.
export const sellerOptionSchema = z.object({
  id: z.string().min(1),
  name: z.string(),
  code: z.string(),
})

export type SellerOption = z.infer<typeof sellerOptionSchema>

export type SearchSellers = (params: PartyParams, signal?: AbortSignal) => Promise<Page<SellerOption>>

const unitSchema = z.enum(['unit', 'kg', 'l', 'm'])

// A discount as the backend takes it: a percent in hundredths of a point
// (10 % is 1000) or an amount in cents.
export const discountSchema = z.object({
  type: z.enum(['percent', 'amount']),
  value: z.number().int().positive(),
})

export type Discount = z.infer<typeof discountSchema>

// A cart line keeps a copy of the item as it was when it was added: the
// backend prices the sale again and refuses it if the total changed.
export const cartLineSchema = z.object({
  itemId: z.string().min(1),
  code: z.string(),
  name: z.string(),
  unit: unitSchema,
  unitPrice: z.number().int().positive(),
  quantity: z.number().int().positive(),
  // Absent or null: no discount on this line.
  discount: discountSchema.nullable().optional(),
})

export type CartLine = z.infer<typeof cartLineSchema>

export const storedCartSchema = z.object({
  // Version 1 had no discounts: its carts read as version 2 without any.
  version: z.union([z.literal(1), z.literal(2)]),
  lines: z.array(cartLineSchema),
  // Absent or null: no general discount.
  cartDiscount: discountSchema.nullable().optional(),
  // Absent or null: Consumidor Final.
  customer: customerOptionSchema.nullable().optional(),
})

export const paymentMethodSchema = z.enum(['cash', 'debit_card', 'credit_card', 'transfer', 'qr'])

export type PaymentMethod = z.infer<typeof paymentMethodSchema>

// The sale the backend just recorded, as the checkout and the ticket show it.
export const saleSchema = z.object({
  id: z.string().min(1),
  number: z.number().int(),
  customer: z.object({ name: z.string() }),
  seller: z.object({ name: z.string() }).nullable(),
  lines: z.array(
    z.object({
      itemId: z.string().min(1),
      name: z.string(),
      unit: unitSchema,
      quantity: z.number().int(),
      unitPrice: z.number().int(),
      discount: discountSchema.nullable(),
      discountAmount: z.number().int(),
      cartDiscountAmount: z.number().int(),
      lineTotal: z.number().int(),
    }),
  ),
  subtotal: z.number().int(),
  cartDiscount: discountSchema.nullable(),
  cartDiscountAmount: z.number().int(),
  discountTotal: z.number().int(),
  total: z.number().int(),
  change: z.number().int(),
  payments: z.array(
    z.object({
      method: paymentMethodSchema,
      amount: z.number().int(),
      reference: z.string().nullable(),
    }),
  ),
  createdAt: z.string(),
})

export type Sale = z.infer<typeof saleSchema>

export type PaymentRequest = { method: PaymentMethod; amount: number; reference: string | null }

export type CreateSaleRequest = {
  customerId: string | null
  sellerId: string | null
  lines: Array<{ itemId: string; quantity: number; discount: Discount | null }>
  cartDiscount: Discount | null
  payments: PaymentRequest[]
  expectedTotal: number
}
