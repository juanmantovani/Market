import { z } from 'zod'
import { pageSchema } from '../../shared/api/pagination'
import { parsePercent, percentToInput } from '../../shared/lib/percent'

const MAX_PERCENT = 10_000

export const sellerSchema = z.object({
  id: z.string(),
  name: z.string(),
  code: z.string(),
  commissionRate: z.number().int().nullable(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export type Seller = z.infer<typeof sellerSchema>

export const sellerPageSchema = pageSchema(sellerSchema)

// Mirrors domain.NewSeller in the backend, so most mistakes are caught
// before the request; the backend still has the final word.
export const sellerFormSchema = z
  .object({
    name: z.string().trim().min(1, 'El nombre es obligatorio').max(80, 'El nombre admite hasta 80 caracteres'),
    code: z
      .string()
      .trim()
      .min(1, 'El código es obligatorio')
      .regex(/^[A-Za-z0-9]{1,6}$/, 'El código lleva de 1 a 6 letras o números, sin espacios'),
    commissionRate: z.string().refine((value) => {
      if (value.trim() === '') return true
      const rate = parsePercent(value)
      return rate !== null && rate <= MAX_PERCENT
    }, 'Ingresar un porcentaje de 0 a 100, por ejemplo 2,5'),
  })
  .transform((values) => ({
    name: values.name,
    code: values.code.toUpperCase(),
    commissionRate: values.commissionRate.trim() === '' ? null : parsePercent(values.commissionRate),
  }))

export type SellerFormValues = z.input<typeof sellerFormSchema>
export type SellerRequest = z.output<typeof sellerFormSchema>
export type SellerField = keyof SellerFormValues

// Backend error codes that belong to a specific form field.
export const SELLER_FIELD_ERRORS: Record<string, SellerField> = {
  seller_name_required: 'name',
  seller_name_too_long: 'name',
  seller_code_required: 'code',
  seller_code_invalid: 'code',
  seller_code_in_use: 'code',
  seller_commission_invalid: 'commissionRate',
}

export function toFormValues(seller: Seller | null): SellerFormValues {
  return {
    name: seller?.name ?? '',
    code: seller?.code ?? '',
    commissionRate: percentToInput(seller?.commissionRate ?? null),
  }
}
