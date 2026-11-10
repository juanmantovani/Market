import { useState, type FormEvent } from 'react'
import { ApiError, messageFromError } from '../../../shared/api/client'
import { Button } from '../../../shared/ui/button'
import { Field, FieldDescription, FieldError, FieldLabel } from '../../../shared/ui/field'
import { Input } from '../../../shared/ui/input'
import { Textarea } from '../../../shared/ui/textarea'
import {
  CATEGORY_FIELD_ERRORS,
  categoryFormSchema,
  MAX_DESCRIPTION_LENGTH,
  type CategoryField,
  type CategoryFormValues,
} from '../types'

type CategoryFormProps = {
  initialValues: CategoryFormValues
  submitLabel: string
  onSubmit: (values: CategoryFormValues) => Promise<unknown>
  onCancel: () => void
}

type FieldErrors = Partial<Record<CategoryField, string>>

export function CategoryForm({ initialValues, submitLabel, onSubmit, onCancel }: CategoryFormProps) {
  const [values, setValues] = useState(initialValues)
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  function change(field: CategoryField, value: string) {
    setValues((current) => ({ ...current, [field]: value }))
    setFieldErrors((current) => ({ ...current, [field]: undefined }))
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const parsed = categoryFormSchema.safeParse(values)
    if (!parsed.success) {
      const errors: FieldErrors = {}
      for (const issue of parsed.error.issues) {
        const field = issue.path[0] as CategoryField
        errors[field] ??= issue.message
      }
      setFieldErrors(errors)
      return
    }

    setIsSubmitting(true)
    try {
      await onSubmit(parsed.data)
    } catch (error) {
      const field = error instanceof ApiError ? CATEGORY_FIELD_ERRORS[error.code] : undefined
      if (field) {
        setFieldErrors({ [field]: messageFromError(error) })
      } else {
        setFormError(messageFromError(error))
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  return (
    <form noValidate onSubmit={handleSubmit} className="flex flex-col gap-5">
      <Field>
        <FieldLabel htmlFor="category-name">Nombre</FieldLabel>
        <Input
          id="category-name"
          autoFocus
          autoComplete="off"
          value={values.name}
          onChange={(event) => change('name', event.target.value)}
          aria-invalid={Boolean(fieldErrors.name)}
          aria-describedby={fieldErrors.name ? 'category-name-error' : undefined}
        />
        <FieldError id="category-name-error">{fieldErrors.name}</FieldError>
      </Field>

      <Field>
        <FieldLabel htmlFor="category-description">
          Descripción <span className="font-normal text-muted-foreground">(opcional)</span>
        </FieldLabel>
        <Textarea
          id="category-description"
          rows={3}
          value={values.description}
          onChange={(event) => change('description', event.target.value)}
          aria-invalid={Boolean(fieldErrors.description)}
          aria-describedby="category-description-hint category-description-error"
        />
        <FieldDescription id="category-description-hint">
          {values.description.trim().length}/{MAX_DESCRIPTION_LENGTH} caracteres
        </FieldDescription>
        <FieldError id="category-description-error">{fieldErrors.description}</FieldError>
      </Field>

      {formError && (
        <p role="alert" className="rounded-lg bg-destructive/10 p-3 text-sm text-destructive">
          {formError}
        </p>
      )}

      <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
        <Button type="button" variant="outline" size="lg" className="h-11 md:h-9" onClick={onCancel}>
          Cancelar
        </Button>
        <Button type="submit" size="lg" className="h-11 md:h-9" disabled={isSubmitting}>
          {isSubmitting ? 'Guardando…' : submitLabel}
        </Button>
      </div>
    </form>
  )
}
