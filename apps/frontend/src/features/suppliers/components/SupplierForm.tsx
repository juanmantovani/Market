import { useState, type FormEvent, type ReactNode } from 'react'
import { ApiError, messageFromError } from '../../../shared/api/client'
import { Button } from '../../../shared/ui/button'
import { Field, FieldError, FieldLabel } from '../../../shared/ui/field'
import { Input } from '../../../shared/ui/input'
import { Textarea } from '../../../shared/ui/textarea'
import {
  SUPPLIER_FIELD_ERRORS,
  supplierFormSchema,
  type SupplierField,
  type SupplierFormValues,
  type SupplierRequest,
} from '../types'

type SupplierFormProps = {
  initialValues: SupplierFormValues
  submitLabel: string
  onSubmit: (request: SupplierRequest) => Promise<unknown>
  onCancel: () => void
}

type FieldErrors = Partial<Record<SupplierField, string>>

function Section({ title, children }: { title: string; children: ReactNode }) {
  return (
    <fieldset className="flex flex-col gap-4">
      <legend className="mb-1 text-sm font-semibold text-muted-foreground">{title}</legend>
      {children}
    </fieldset>
  )
}

function Optional() {
  return <span className="font-normal text-muted-foreground">(opcional)</span>
}

function errorProps(field: SupplierField, errors: FieldErrors) {
  return {
    id: `supplier-${field}`,
    'aria-invalid': Boolean(errors[field]),
    'aria-describedby': errors[field] ? `supplier-${field}-error` : undefined,
  }
}

export function SupplierForm({ initialValues, submitLabel, onSubmit, onCancel }: SupplierFormProps) {
  const [values, setValues] = useState(initialValues)
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  function change(field: SupplierField, value: string) {
    setValues((current) => ({ ...current, [field]: value }))
    setFieldErrors((current) => ({ ...current, [field]: undefined }))
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const parsed = supplierFormSchema.safeParse(values)
    if (!parsed.success) {
      const errors: FieldErrors = {}
      for (const issue of parsed.error.issues) {
        const field = issue.path[0] as SupplierField
        errors[field] ??= issue.message
      }
      setFieldErrors(errors)
      return
    }

    setIsSubmitting(true)
    try {
      await onSubmit(parsed.data)
    } catch (error) {
      const field = error instanceof ApiError ? SUPPLIER_FIELD_ERRORS[error.code] : undefined
      if (field) {
        setFieldErrors({ [field]: messageFromError(error) })
      } else {
        setFormError(messageFromError(error))
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  const error = (field: SupplierField) => <FieldError id={`supplier-${field}-error`}>{fieldErrors[field]}</FieldError>

  return (
    <form noValidate onSubmit={handleSubmit} className="flex flex-col gap-6">
      <Section title="Datos del proveedor">
        <Field>
          <FieldLabel htmlFor="supplier-businessName">Razón social</FieldLabel>
          <Input
            {...errorProps('businessName', fieldErrors)}
            autoFocus
            autoComplete="off"
            value={values.businessName}
            onChange={(event) => change('businessName', event.target.value)}
          />
          {error('businessName')}
        </Field>
        <Field className="sm:max-w-[calc(50%-0.5rem)]">
          <FieldLabel htmlFor="supplier-cuit">
            CUIT <Optional />
          </FieldLabel>
          <Input
            {...errorProps('cuit', fieldErrors)}
            autoComplete="off"
            inputMode="numeric"
            value={values.cuit}
            onChange={(event) => change('cuit', event.target.value)}
          />
          {error('cuit')}
        </Field>
      </Section>

      <Section title="Contacto">
        <div className="grid gap-4 sm:grid-cols-2">
          <Field>
            <FieldLabel htmlFor="supplier-contactName">
              Persona de contacto <Optional />
            </FieldLabel>
            <Input
              {...errorProps('contactName', fieldErrors)}
              autoComplete="off"
              value={values.contactName}
              onChange={(event) => change('contactName', event.target.value)}
            />
            {error('contactName')}
          </Field>
          <Field>
            <FieldLabel htmlFor="supplier-phone">
              Teléfono <Optional />
            </FieldLabel>
            <Input
              {...errorProps('phone', fieldErrors)}
              type="tel"
              autoComplete="off"
              value={values.phone}
              onChange={(event) => change('phone', event.target.value)}
            />
            {error('phone')}
          </Field>
        </div>
        <Field>
          <FieldLabel htmlFor="supplier-email">
            Email <Optional />
          </FieldLabel>
          <Input
            {...errorProps('email', fieldErrors)}
            type="email"
            autoComplete="off"
            value={values.email}
            onChange={(event) => change('email', event.target.value)}
          />
          {error('email')}
        </Field>
        <Field>
          <FieldLabel htmlFor="supplier-address">
            Dirección <Optional />
          </FieldLabel>
          <Input
            {...errorProps('address', fieldErrors)}
            autoComplete="off"
            value={values.address}
            onChange={(event) => change('address', event.target.value)}
          />
          {error('address')}
        </Field>
        <Field>
          <FieldLabel htmlFor="supplier-notes">
            Notas <Optional />
          </FieldLabel>
          <Textarea
            {...errorProps('notes', fieldErrors)}
            rows={2}
            value={values.notes}
            onChange={(event) => change('notes', event.target.value)}
          />
          {error('notes')}
        </Field>
      </Section>

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
