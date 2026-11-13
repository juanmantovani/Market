import { z } from 'zod'
import { pageSchema } from '../../shared/api/pagination'
import { parseMoney } from '../../shared/lib/money'
import { formatUnitQuantity, parseQuantity, type Unit } from '../../shared/lib/quantity'

const unitSchema = z.enum(['unit', 'kg', 'l', 'm'])

export const stockLevelSchema = z.object({
  itemId: z.string(),
  code: z.string(),
  name: z.string(),
  categoryId: z.string(),
  unit: unitSchema,
  minStock: z.number().int(),
  quantity: z.number().int(),
  lowStock: z.boolean(),
  updatedAt: z.string().nullable(),
})

export type StockLevel = z.infer<typeof stockLevelSchema>

export const stockLevelPageSchema = pageSchema(stockLevelSchema)

export const MOVEMENT_TYPE_LABELS = {
  initial: 'Stock inicial',
  entry: 'Ingreso',
  adjustment: 'Ajuste',
  sale: 'Venta',
  sale_void: 'Anulación de venta',
} as const

export const stockMovementSchema = z.object({
  id: z.string(),
  itemId: z.string(),
  type: z.enum(['initial', 'entry', 'adjustment', 'sale', 'sale_void']),
  quantity: z.number().int(),
  balanceAfter: z.number().int(),
  reason: z.string().nullable(),
  reference: z.string().nullable(),
  supplierId: z.string().nullable(),
  unitCost: z.number().int().nullable(),
  actorId: z.string(),
  createdAt: z.string(),
})

export type StockMovement = z.infer<typeof stockMovementSchema>

export const stockMovementPageSchema = pageSchema(stockMovementSchema)

export const stockEntryResponseSchema = z.object({ movements: z.array(stockMovementSchema) })

export type CategoryOption = { id: string; name: string }
export type SupplierOption = { id: string; name: string }

// A positive quantity in thousandths, whole for items sold by the unit, or
// null when the text isn't one.
function positiveQuantity(text: string, unit: Unit): number | null {
  const quantity = parseQuantity(text, unit !== 'unit')
  return quantity !== null && quantity > 0 ? quantity : null
}

function quantityMessage(unit: Unit) {
  return unit === 'unit' ? 'Ingresar una cantidad entera mayor a cero' : 'Ingresar una cantidad mayor a cero, por ejemplo 2,5'
}

export const ADJUSTMENT_REASONS = ['Conteo', 'Rotura', 'Vencimiento', 'Pérdida o robo', 'Consumo interno'] as const

// The form keeps what the person typed; the schema checks it against the
// current stock, as domain.StockLevel.Adjust does, and turns it into a signed
// quantity.
export function adjustmentFormSchema(level: Pick<StockLevel, 'unit' | 'quantity'>) {
  return z
    .object({
      direction: z.enum(['in', 'out']),
      quantity: z.string(),
      reason: z.string().trim().min(1, 'Ingresar el motivo del ajuste').max(200, 'El motivo admite hasta 200 caracteres'),
    })
    .superRefine((values, ctx) => {
      const quantity = positiveQuantity(values.quantity, level.unit)
      if (quantity === null) {
        ctx.addIssue({ code: 'custom', path: ['quantity'], message: quantityMessage(level.unit) })
      } else if (values.direction === 'out' && level.quantity - quantity < 0) {
        ctx.addIssue({
          code: 'custom',
          path: ['quantity'],
          message: `No hay stock suficiente: hay ${formatUnitQuantity(Math.max(level.quantity, 0), level.unit)}`,
        })
      }
    })
    .transform((values) => {
      const quantity = positiveQuantity(values.quantity, level.unit) as number
      return { quantity: values.direction === 'out' ? -quantity : quantity, reason: values.reason }
    })
}

export type AdjustmentFormValues = z.input<ReturnType<typeof adjustmentFormSchema>>
export type AdjustmentFields = z.output<ReturnType<typeof adjustmentFormSchema>>
export type AdjustmentField = keyof AdjustmentFormValues
export type StockAdjustmentRequest = AdjustmentFields & { itemId: string }

export const ADJUSTMENT_FIELD_ERRORS: Record<string, AdjustmentField> = {
  stock_adjustment_quantity_invalid: 'quantity',
  stock_quantity_not_whole: 'quantity',
  stock_balance_out_of_range: 'quantity',
  insufficient_stock: 'quantity',
  stock_reason_required: 'reason',
  stock_reason_too_long: 'reason',
}

const itemChoiceSchema = z.object({
  itemId: z.string(),
  code: z.string(),
  name: z.string(),
  unit: unitSchema,
  quantity: z.number().int(),
})

// The item a line of a stock entry receives, as the search returned it.
export type ItemChoice = z.infer<typeof itemChoiceSchema>

const entryLineSchema = z
  .object({
    key: z.number(),
    item: itemChoiceSchema.nullable(),
    quantity: z.string(),
    unitCost: z
      .string()
      .refine((value) => value.trim() === '' || parseMoney(value) !== null, 'Ingresar un costo válido, por ejemplo 800,00'),
  })
  .superRefine((line, ctx) => {
    if (line.item === null) {
      ctx.addIssue({ code: 'custom', path: ['item'], message: 'Elegir el artículo' })
    } else if (positiveQuantity(line.quantity, line.item.unit) === null) {
      ctx.addIssue({ code: 'custom', path: ['quantity'], message: quantityMessage(line.item.unit) })
    }
  })

// Mirrors domain.NewStockEntry: at least one line, each item once.
export const entryFormSchema = z
  .object({
    supplierId: z.string(),
    reference: z.string().trim().max(60, 'El comprobante admite hasta 60 caracteres'),
    lines: z.array(entryLineSchema).min(1, 'Agregar al menos un artículo').max(100, 'El ingreso admite hasta 100 artículos'),
  })
  .superRefine((values, ctx) => {
    const seen = new Set<string>()
    values.lines.forEach((line, index) => {
      if (line.item === null) return
      if (seen.has(line.item.itemId)) {
        ctx.addIssue({ code: 'custom', path: ['lines', index, 'item'], message: 'Este artículo ya está en otra línea' })
      }
      seen.add(line.item.itemId)
    })
  })
  .transform((values) => ({
    supplierId: values.supplierId || null,
    reference: values.reference || null,
    lines: values.lines.map((line) => {
      const item = line.item as ItemChoice
      return {
        itemId: item.itemId,
        quantity: positiveQuantity(line.quantity, item.unit) as number,
        unitCost: line.unitCost.trim() === '' ? null : parseMoney(line.unitCost),
      }
    }),
  }))

export type EntryFormValues = z.input<typeof entryFormSchema>
export type EntryLineValues = EntryFormValues['lines'][number]
export type StockEntryRequest = z.output<typeof entryFormSchema>

// Backend error codes that belong to a field of the entry form. The backend
// doesn't say which line failed, so line errors show for the whole form.
export const ENTRY_FIELD_ERRORS: Record<string, 'supplierId' | 'reference' | 'lines'> = {
  stock_supplier_not_found: 'supplierId',
  stock_reference_too_long: 'reference',
  stock_entry_lines_required: 'lines',
  stock_entry_too_many_lines: 'lines',
}

export function emptyEntryLine(key: number, item: ItemChoice | null = null): EntryLineValues {
  return { key, item, quantity: '', unitCost: '' }
}

// The item a stock action is about: what Artículos knows of it, handed over
// by app/ItemsRoute because stock doesn't import the items feature.
export type StockItem = {
  id: string
  code: string
  name: string
  categoryId: string
  unit: Unit
  minStock: number
}

export function initialStockFormSchema(unit: Unit) {
  return z
    .object({ quantity: z.string() })
    .superRefine((values, ctx) => {
      if (positiveQuantity(values.quantity, unit) === null) {
        ctx.addIssue({ code: 'custom', path: ['quantity'], message: quantityMessage(unit) })
      }
    })
    .transform((values) => ({ quantity: positiveQuantity(values.quantity, unit) as number }))
}

export type InitialStockFormValues = z.input<ReturnType<typeof initialStockFormSchema>>
export type InitialStockFields = z.output<ReturnType<typeof initialStockFormSchema>>
export type StockInitialRequest = InitialStockFields & { itemId: string }

export const INITIAL_STOCK_FIELD_ERRORS: Record<string, 'quantity'> = {
  stock_quantity_not_whole: 'quantity',
  stock_balance_out_of_range: 'quantity',
}
