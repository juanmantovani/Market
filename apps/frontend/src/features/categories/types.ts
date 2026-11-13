import { z } from 'zod'
import { pageSchema } from '../../shared/api/pagination'

export const MAX_NAME_LENGTH = 60
export const MAX_DESCRIPTION_LENGTH = 200

export const categorySchema = z.object({
  id: z.string(),
  name: z.string(),
  description: z.string().nullable(),
  createdAt: z.string(),
  updatedAt: z.string(),
})

export type Category = z.infer<typeof categorySchema>

export const categoryPageSchema = pageSchema(categorySchema)

// Mirrors domain.NewCategory in the backend, so most mistakes are caught
// before the request; the backend still has the final word.
export const categoryFormSchema = z.object({
  name: z
    .string()
    .trim()
    .min(1, 'El nombre es obligatorio')
    .max(MAX_NAME_LENGTH, `El nombre admite hasta ${MAX_NAME_LENGTH} caracteres`),
  description: z
    .string()
    .trim()
    .max(MAX_DESCRIPTION_LENGTH, `La descripción admite hasta ${MAX_DESCRIPTION_LENGTH} caracteres`),
})

export type CategoryFormValues = z.input<typeof categoryFormSchema>

export type CategoryField = keyof CategoryFormValues

// Backend error codes that belong to a specific form field.
export const CATEGORY_FIELD_ERRORS: Record<string, CategoryField> = {
  category_name_required: 'name',
  category_name_too_long: 'name',
  category_name_in_use: 'name',
  category_description_too_long: 'description',
}

export function toFormValues(category: Category | null): CategoryFormValues {
  return { name: category?.name ?? '', description: category?.description ?? '' }
}
