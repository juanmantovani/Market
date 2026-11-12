import { z } from 'zod'
import { pageSchema } from '../../shared/api/pagination'
import { documentDigits, formatCuit } from '../../shared/lib/document'

export const supplierSchema = z.object({
  id: z.string(),
  businessName: z.string(),
  cuit: z.string().nullable(),
  contactName: z.string().nullable(),
  phone: z.string().nullable(),
  email: z.string().nullable(),
  address: z.string().nullable(),
  notes: z.string().nullable(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export type Supplier = z.infer<typeof supplierSchema>

export const supplierPageSchema = pageSchema(supplierSchema)

const EMAIL_SHAPE = /^[^\s@]+@[^\s@]+\.[^\s@]+$/

// Mirrors domain.NewSupplier in the backend, so most mistakes are caught
// before the request; the backend still has the final word.
export const supplierFormSchema = z
  .object({
    businessName: z
      .string()
      .trim()
      .min(1, 'La razón social es obligatoria')
      .max(120, 'La razón social admite hasta 120 caracteres'),
    cuit: z.string().refine((value) => {
      const digits = documentDigits(value)
      return digits === '' || /^\d{11}$/.test(digits)
    }, 'El CUIT lleva 11 números, por ejemplo 30-71234567-8'),
    contactName: z.string().trim().max(120, 'El nombre de contacto admite hasta 120 caracteres'),
    phone: z.string().trim().max(30, 'El teléfono admite hasta 30 caracteres'),
    email: z
      .string()
      .trim()
      .max(254, 'El email admite hasta 254 caracteres')
      .refine((value) => value === '' || EMAIL_SHAPE.test(value), 'Ingresar un email válido, por ejemplo nombre@dominio.com'),
    address: z.string().trim().max(200, 'La dirección admite hasta 200 caracteres'),
    notes: z.string().trim().max(500, 'Las notas admiten hasta 500 caracteres'),
  })
  .transform((values) => ({
    businessName: values.businessName,
    cuit: documentDigits(values.cuit) || null,
    contactName: values.contactName || null,
    phone: values.phone || null,
    email: values.email || null,
    address: values.address || null,
    notes: values.notes || null,
  }))

export type SupplierFormValues = z.input<typeof supplierFormSchema>
export type SupplierRequest = z.output<typeof supplierFormSchema>
export type SupplierField = keyof SupplierFormValues

// Backend error codes that belong to a specific form field.
export const SUPPLIER_FIELD_ERRORS: Record<string, SupplierField> = {
  supplier_business_name_required: 'businessName',
  supplier_business_name_too_long: 'businessName',
  supplier_cuit_invalid: 'cuit',
  supplier_cuit_in_use: 'cuit',
  supplier_contact_name_too_long: 'contactName',
  supplier_email_invalid: 'email',
  supplier_phone_too_long: 'phone',
  supplier_address_too_long: 'address',
  supplier_notes_too_long: 'notes',
}

export function toFormValues(supplier: Supplier | null): SupplierFormValues {
  return {
    businessName: supplier?.businessName ?? '',
    cuit: supplier?.cuit ? formatCuit(supplier.cuit) : '',
    contactName: supplier?.contactName ?? '',
    phone: supplier?.phone ?? '',
    email: supplier?.email ?? '',
    address: supplier?.address ?? '',
    notes: supplier?.notes ?? '',
  }
}
