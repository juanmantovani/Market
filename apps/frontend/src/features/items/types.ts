import { z } from 'zod'
import { pageSchema } from '../../shared/api/pagination'
import { centsToInput, parseMoney } from '../../shared/lib/money'
import { formatQuantity, parseQuantity, type Unit } from '../../shared/lib/quantity'

export const UNITS: readonly { value: Unit; label: string }[] = [
  { value: 'unit', label: 'Unidad' },
  { value: 'kg', label: 'Kilogramo' },
  { value: 'l', label: 'Litro' },
  { value: 'm', label: 'Metro' },
]

// Hundredths of a percent, like the backend's domain.Percent.
export const TAX_RATES = [
  { value: 2100, label: '21 %' },
  { value: 1050, label: '10,5 %' },
  { value: 2700, label: '27 %' },
  { value: 0, label: 'Exento (0 %)' },
] as const

const unitSchema = z.enum(['unit', 'kg', 'l', 'm'])

export const itemSchema = z.object({
  id: z.string(),
  code: z.string(),
  barcode: z.string().nullable(),
  name: z.string(),
  description: z.string().nullable(),
  categoryId: z.string(),
  supplierIds: z.array(z.string()),
  unit: unitSchema,
  salePrice: z.number().int(),
  cost: z.number().int().nullable(),
  taxRate: z.number().int(),
  minStock: z.number().int(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export type Item = z.infer<typeof itemSchema>

export const itemPageSchema = pageSchema(itemSchema)

export type CategoryOption = { id: string; name: string }
export type SupplierOption = { id: string; name: string }

// The current stock of the items listed, composed in app/ItemsRoute: items
// don't import the stock feature.
export type StockQuantities = {
  // In thousandths of each item's unit, by item ID.
  quantities: ReadonlyMap<string, number>
  isLoading: boolean
}

export type UseStockQuantities = (itemIds: readonly string[]) => StockQuantities

// The form keeps what the person typed (amounts as text) and the schema
// turns it into the request the backend expects, mirroring domain.NewItem.
export const itemFormSchema = z
  .object({
    code: z.string().trim().min(1, 'El código es obligatorio').max(30, 'El código admite hasta 30 caracteres'),
    barcode: z
      .string()
      .trim()
      .max(32, 'El código de barras admite hasta 32 caracteres')
      .refine((value) => !/\s/.test(value), 'El código de barras no puede tener espacios'),
    name: z.string().trim().min(1, 'El nombre es obligatorio').max(100, 'El nombre admite hasta 100 caracteres'),
    description: z.string().trim().max(500, 'La descripción admite hasta 500 caracteres'),
    categoryId: z.string().min(1, 'Elegir una categoría'),
    supplierIds: z.array(z.string()),
    unit: unitSchema,
    salePrice: z.string().refine((value) => {
      const cents = parseMoney(value)
      return cents !== null && cents > 0
    }, 'Ingresar un precio mayor a cero, por ejemplo 1.250,50'),
    cost: z
      .string()
      .refine((value) => value.trim() === '' || parseMoney(value) !== null, 'Ingresar un costo válido, por ejemplo 800,00'),
    taxRate: z.string(),
    minStock: z
      .string()
      .refine((value) => value.trim() === '' || parseQuantity(value, true) !== null, 'Ingresar una cantidad, por ejemplo 2,5'),
  })
  // Runs once every field is valid on its own: the only rule left is the one
  // that depends on two fields.
  .transform((values, ctx) => {
    const minStock = values.minStock.trim() === '' ? 0 : parseQuantity(values.minStock, values.unit !== 'unit')
    if (minStock === null) {
      ctx.addIssue({ code: 'custom', path: ['minStock'], message: 'Ingresar una cantidad entera' })
      return z.NEVER
    }
    const cost = values.cost.trim() === '' ? null : parseMoney(values.cost)

    return {
      code: values.code,
      barcode: values.barcode || null,
      name: values.name,
      description: values.description || null,
      categoryId: values.categoryId,
      supplierIds: values.supplierIds,
      unit: values.unit,
      salePrice: parseMoney(values.salePrice) as number,
      cost,
      taxRate: Number(values.taxRate),
      minStock,
    }
  })

export type ItemFormValues = z.input<typeof itemFormSchema>
export type ItemRequest = z.output<typeof itemFormSchema>
export type ItemField = keyof ItemFormValues

// Backend error codes that belong to a specific form field.
export const ITEM_FIELD_ERRORS: Record<string, ItemField> = {
  item_code_required: 'code',
  item_code_too_long: 'code',
  item_code_in_use: 'code',
  item_barcode_invalid: 'barcode',
  item_barcode_in_use: 'barcode',
  item_name_required: 'name',
  item_name_too_long: 'name',
  item_description_too_long: 'description',
  item_category_required: 'categoryId',
  item_category_not_found: 'categoryId',
  item_supplier_not_found: 'supplierIds',
  item_unit_invalid: 'unit',
  item_price_invalid: 'salePrice',
  item_cost_invalid: 'cost',
  item_tax_rate_invalid: 'taxRate',
  item_min_stock_invalid: 'minStock',
}

export function toFormValues(item: Item | null): ItemFormValues {
  return {
    code: item?.code ?? '',
    barcode: item?.barcode ?? '',
    name: item?.name ?? '',
    description: item?.description ?? '',
    categoryId: item?.categoryId ?? '',
    supplierIds: item?.supplierIds ?? [],
    unit: item?.unit ?? 'unit',
    salePrice: item ? centsToInput(item.salePrice) : '',
    cost: item ? centsToInput(item.cost) : '',
    taxRate: String(item?.taxRate ?? 2100),
    minStock: item ? formatQuantity(item.minStock) : '0',
  }
}
