import { z } from 'zod'
import type { Unit } from '../../shared/lib/quantity'

const methodSchema = z.enum(['cash', 'debit_card', 'credit_card', 'transfer', 'qr'])

export const METHOD_LABELS: Record<z.infer<typeof methodSchema>, string> = {
  cash: 'Efectivo',
  debit_card: 'Tarjeta de débito',
  credit_card: 'Tarjeta de crédito',
  transfer: 'Transferencia',
  qr: 'QR',
}

const lowStockItemSchema = z.object({
  itemId: z.string(),
  code: z.string(),
  name: z.string(),
  unit: z.enum(['unit', 'kg', 'l', 'm']).transform((unit): Unit => unit),
  minStock: z.number().int(),
  quantity: z.number().int(),
})

export const summarySchema = z.object({
  from: z.string(),
  to: z.string(),
  sales: z.object({
    count: z.number().int(),
    total: z.number().int(),
    byMethod: z.array(z.object({ method: methodSchema, amount: z.number().int() })),
  }),
  lowStock: z.object({
    total: z.number().int(),
    items: z.array(lowStockItemSchema),
  }),
})

export type Summary = z.infer<typeof summarySchema>
