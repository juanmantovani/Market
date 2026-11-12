import { useState, type FormEvent } from 'react'
import { ApiError, messageFromError } from '../../../shared/api/client'
import { Button } from '../../../shared/ui/button'
import { Field, FieldDescription, FieldError, FieldLabel } from '../../../shared/ui/field'
import { Input } from '../../../shared/ui/input'
import {
  SELLER_FIELD_ERRORS,
  sellerFormSchema,
  type SellerField,
  type SellerFormValues,
  type SellerRequest,
} from '../types'

type SellerFormProps = {
  initialValues: SellerFormValues
  submitLabel: string
  onSubmit: (request: SellerRequest) => Promise<unknown>
  onCancel: () => void
}

type FieldErrors = Partial<Record<SellerField, string>>

function errorProps(field: SellerField, errors: FieldErrors, hint?: string) {
  const describedBy = [errors[field] && `seller-${field}-error`, hint].filter(Boolean).join(' ')
  return {
    id: `seller-${field}`,
    'aria-invalid': Boolean(errors[field]),
    'aria-describedby': describedBy || undefined,
  }
}

export function SellerForm({ initialValues, submitLabel, onSubmit, onCancel }: SellerFormProps) {
  const [values, setValues] = useState(initialValues)
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  function change(field: SellerField, value: string) {
    setValues((current) => ({ ...current, [field]: value }))
    setFieldErrors((current) => ({ ...current, [field]: undefined }))
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const parsed = sellerFormSchema.safeParse(values)
    if (!parsed.success) {
      const errors: FieldErrors = {}
      for (const issue of parsed.error.issues) {
        const field = issue.path[0] as SellerField
        errors[field] ??= issue.message
      }
      setFieldErrors(errors)
      return
    }

    setIsSubmitting(true)
    try {
      await onSubmit(parsed.data)
    } catch (error) {
      const field = error instanceof ApiError ? SELLER_FIELD_ERRORS[error.code] : undefined
      if (field) {
        setFieldErrors({ [field]: messageFromError(error) })
      } else {
        setFormError(messageFromError(error))
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  const error = (field: SellerField) => <FieldError id={`seller-${field}-error`}>{fieldErrors[field]}</FieldError>

  return (
    <form noValidate onSubmit={handleSubmit} className="flex flex-col gap-5">
      <Field>
        <FieldLabel htmlFor="seller-name">Nombre</FieldLabel>
        <Input
          {...errorProps('name', fieldErrors)}
          autoFocus
          autoComplete="off"
          value={values.name}
          onChange={(event) => change('name', event.target.value)}
        />
        {error('name')}
      </Field>
      <div className="grid gap-5 sm:grid-cols-2">
        <Field>
          <FieldLabel htmlFor="seller-code">Código</FieldLabel>
          <Input
            {...errorProps('code', fieldErrors, 'seller-code-hint')}
            autoComplete="off"
            autoCapitalize="characters"
            spellCheck={false}
            className="font-mono uppercase"
            value={values.code}
            onChange={(event) => change('code', event.target.value.toUpperCase())}
          />
          <FieldDescription id="seller-code-hint">Con este código se lo elige en la venta, por ejemplo JP.</FieldDescription>
          {error('code')}
        </Field>
        <Field>
          <FieldLabel htmlFor="seller-commissionRate">
            Comisión <span className="font-normal text-muted-foreground">(opcional)</span>
          </FieldLabel>
          <div className="relative">
            <Input
              {...errorProps('commissionRate', fieldErrors)}
              inputMode="decimal"
              autoComplete="off"
              placeholder="0"
              className="pr-8"
              value={values.commissionRate}
              onChange={(event) => change('commissionRate', event.target.value)}
            />
            <span className="pointer-events-none absolute top-1/2 right-3 -translate-y-1/2 text-muted-foreground">%</span>
          </div>
          {error('commissionRate')}
        </Field>
      </div>

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
