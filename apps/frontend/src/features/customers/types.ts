import { z } from 'zod'
import { pageSchema } from '../../shared/api/pagination'
import { documentDigits, formatCuit, formatDni } from '../../shared/lib/document'

export const DOCUMENT_TYPES = [
  { value: 'dni', label: 'DNI' },
  { value: 'cuit', label: 'CUIT' },
  { value: 'cuil', label: 'CUIL' },
  { value: 'none', label: 'Sin documento' },
] as const

export const TAX_CONDITIONS = [
  { value: 'final_consumer', label: 'Consumidor final' },
  { value: 'monotax', label: 'Monotributista' },
  { value: 'registered', label: 'Responsable inscripto' },
  { value: 'exempt', label: 'Exento' },
] as const

const documentTypeSchema = z.enum(['dni', 'cuit', 'cuil', 'none'])
const taxConditionSchema = z.enum(['final_consumer', 'monotax', 'registered', 'exempt'])

export type DocumentType = z.infer<typeof documentTypeSchema>
export type TaxCondition = z.infer<typeof taxConditionSchema>

export const customerSchema = z.object({
  id: z.string(),
  name: z.string(),
  documentType: documentTypeSchema,
  documentNumber: z.string(),
  taxCondition: taxConditionSchema,
  email: z.string().nullable(),
  phone: z.string().nullable(),
  address: z.string().nullable(),
  notes: z.string().nullable(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export type Customer = z.infer<typeof customerSchema>

export const customerPageSchema = pageSchema(customerSchema)

const EMAIL_SHAPE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

export function documentTypeLabel(type: DocumentType): string {
  return DOCUMENT_TYPES.find((option) => option.value === type)?.label ?? type
}

export function taxConditionLabel(condition: TaxCondition): string {
  return TAX_CONDITIONS.find((option) => option.value === condition)?.label ?? condition
}

function documentProblem(type: DocumentType, number: string): string | null {
  if (type === 'none') return null

  const label = documentTypeLabel(type)
  const digits = documentDigits(number)
  if (digits === '') return `Ingresá el número de ${label}`
  if (!/^\d+$/.test(digits)) return `El ${label} lleva solo números`
  if (type === 'dni' && (digits.length < 7 || digits.length > 8)) return 'El DNI lleva 7 u 8 números'
  if (type !== 'dni' && digits.length !== 11) return `El ${label} lleva 11 números`
  return null
}

const documentFieldsSchema = z.object({
  documentType: documentTypeSchema,
  documentNumber: z.string(),
  taxCondition: taxConditionSchema,
})

// Mirrors domain.NewCustomer in the backend, so most mistakes are caught
// before the request; the backend still has the final word.
export const customerFormSchema = z
  .object({
    name: z
      .string()
      .trim()
      .min(1, 'El nombre es obligatorio')
      .max(120, 'El nombre admite hasta 120 caracteres'),
    documentType: documentTypeSchema,
    documentNumber: z.string(),
    taxCondition: taxConditionSchema,
    email: z
      .string()
      .trim()
      .max(254, 'El email admite hasta 254 caracteres')
      .refine((value) => value === '' || EMAIL_SHAPE.test(value), 'Ingresá un email válido, por ejemplo nombre@dominio.com'),
    phone: z.string().trim().max(30, 'El teléfono admite hasta 30 caracteres'),
    address: z.string().trim().max(200, 'La dirección admite hasta 200 caracteres'),
    notes: z.string().trim().max(500, 'Las notas admiten hasta 500 caracteres'),
  })
  .superRefine(
    (values, ctx) => {
      const problem = documentProblem(values.documentType, values.documentNumber)
      if (problem) {
        ctx.addIssue({ code: 'custom', path: ['documentNumber'], message: problem })
      }
      if (values.taxCondition !== 'final_consumer' && values.documentType !== 'cuit') {
        ctx.addIssue({
          code: 'custom',
          path: ['taxCondition'],
          message: 'Para monotributo, responsable inscripto o exento hace falta un CUIT',
        })
      }
    },
    // The document rules only read these fields, so they report alongside
    // errors in the others instead of waiting for them to be fixed.
    { when: (payload) => documentFieldsSchema.safeParse(payload.value).success },
  )
  .transform((values) => ({
    name: values.name,
    documentType: values.documentType,
    documentNumber: values.documentType === 'none' ? '' : documentDigits(values.documentNumber),
    taxCondition: values.taxCondition,
    email: values.email || null,
    phone: values.phone || null,
    address: values.address || null,
    notes: values.notes || null,
  }))

export type CustomerFormValues = z.input<typeof customerFormSchema>
export type CustomerRequest = z.output<typeof customerFormSchema>
export type CustomerField = keyof CustomerFormValues

// Backend error codes that belong to a specific form field.
export const CUSTOMER_FIELD_ERRORS: Record<string, CustomerField> = {
  customer_name_required: 'name',
  customer_name_too_long: 'name',
  customer_document_type_invalid: 'documentType',
  customer_document_invalid: 'documentNumber',
  customer_document_in_use: 'documentNumber',
  customer_tax_condition_invalid: 'taxCondition',
  customer_tax_condition_requires_cuit: 'taxCondition',
  customer_email_invalid: 'email',
  customer_phone_too_long: 'phone',
  customer_address_too_long: 'address',
  customer_notes_too_long: 'notes',
}

export function formatDocumentNumber(type: DocumentType, digits: string): string {
  if (type === 'dni') return formatDni(digits)
  if (type === 'cuit' || type === 'cuil') return formatCuit(digits)
  return digits
}

export function formatDocument(customer: Pick<Customer, 'documentType' | 'documentNumber'>): string | null {
  if (customer.documentType === 'none') return null
  return `${documentTypeLabel(customer.documentType)} ${formatDocumentNumber(customer.documentType, customer.documentNumber)}`
}

export function toFormValues(customer: Customer | null): CustomerFormValues {
  return {
    name: customer?.name ?? '',
    documentType: customer?.documentType ?? 'dni',
    documentNumber: customer ? formatDocumentNumber(customer.documentType, customer.documentNumber) : '',
    taxCondition: customer?.taxCondition ?? 'final_consumer',
    email: customer?.email ?? '',
    phone: customer?.phone ?? '',
    address: customer?.address ?? '',
    notes: customer?.notes ?? '',
  }
}
