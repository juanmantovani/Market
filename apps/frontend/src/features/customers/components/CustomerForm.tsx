import { useState, type FormEvent, type ReactNode } from 'react'
import { ApiError, messageFromError } from '../../../shared/api/client'
import { useFormDialog, useUnsavedChanges } from '../../../shared/hooks/use-form-dialog'
import { Button } from '../../../shared/ui/button'
import { Field, FieldError, FieldLabel } from '../../../shared/ui/field'
import { Input } from '../../../shared/ui/input'
import { NativeSelect } from '../../../shared/ui/native-select'
import { Textarea } from '../../../shared/ui/textarea'
import {
  CUSTOMER_FIELD_ERRORS,
  customerFormSchema,
  DOCUMENT_TYPES,
  TAX_CONDITIONS,
  type CustomerField,
  type CustomerFormValues,
  type CustomerRequest,
  type DocumentType,
  type TaxCondition,
} from '../types'

type CustomerFormProps = {
  initialValues: CustomerFormValues
  submitLabel: string
  onSubmit: (request: CustomerRequest) => Promise<unknown>
}

type FieldErrors = Partial<Record<CustomerField, string>>

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

function errorProps(field: CustomerField, errors: FieldErrors) {
  return {
    id: `customer-${field}`,
    'aria-invalid': Boolean(errors[field]),
    'aria-describedby': errors[field] ? `customer-${field}-error` : undefined,
  }
}

export function CustomerForm({ initialValues, submitLabel, onSubmit }: CustomerFormProps) {
  const [values, setValues] = useState(initialValues)
  const { requestClose } = useFormDialog()
  useUnsavedChanges(values)
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  function change<F extends CustomerField>(field: F, value: CustomerFormValues[F]) {
    setValues((current) => ({ ...current, [field]: value }))
    setFieldErrors((current) => ({ ...current, [field]: undefined }))
  }

  function changeDocumentType(documentType: DocumentType) {
    setValues((current) => ({
      ...current,
      documentType,
      documentNumber: documentType === 'none' ? '' : current.documentNumber,
    }))
    setFieldErrors((current) => ({ ...current, documentType: undefined, documentNumber: undefined, taxCondition: undefined }))
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const parsed = customerFormSchema.safeParse(values)
    if (!parsed.success) {
      const errors: FieldErrors = {}
      for (const issue of parsed.error.issues) {
        const field = issue.path[0] as CustomerField
        errors[field] ??= issue.message
      }
      setFieldErrors(errors)
      return
    }

    setIsSubmitting(true)
    try {
      await onSubmit(parsed.data)
    } catch (error) {
      const field = error instanceof ApiError ? CUSTOMER_FIELD_ERRORS[error.code] : undefined
      if (field) {
        setFieldErrors({ [field]: messageFromError(error) })
      } else {
        setFormError(messageFromError(error))
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  const error = (field: CustomerField) => <FieldError id={`customer-${field}-error`}>{fieldErrors[field]}</FieldError>
  const withoutDocument = values.documentType === 'none'

  return (
    <form noValidate onSubmit={handleSubmit} className="flex flex-col gap-6">
      <Section title="Datos del cliente">
        <Field>
          <FieldLabel htmlFor="customer-name">Nombre o razón social</FieldLabel>
          <Input
            {...errorProps('name', fieldErrors)}
            autoFocus
            autoComplete="off"
            value={values.name}
            onChange={(event) => change('name', event.target.value)}
          />
          {error('name')}
        </Field>
        <div className="grid gap-4 sm:grid-cols-2">
          <Field>
            <FieldLabel htmlFor="customer-documentType">Tipo de documento</FieldLabel>
            <NativeSelect
              {...errorProps('documentType', fieldErrors)}
              value={values.documentType}
              onChange={(event) => changeDocumentType(event.target.value as DocumentType)}
            >
              {DOCUMENT_TYPES.map((type) => (
                <option key={type.value} value={type.value}>
                  {type.label}
                </option>
              ))}
            </NativeSelect>
            {error('documentType')}
          </Field>
          <Field>
            <FieldLabel htmlFor="customer-documentNumber">Número de documento</FieldLabel>
            <Input
              {...errorProps('documentNumber', fieldErrors)}
              autoComplete="off"
              inputMode="numeric"
              disabled={withoutDocument}
              value={values.documentNumber}
              onChange={(event) => change('documentNumber', event.target.value)}
            />
            {error('documentNumber')}
          </Field>
        </div>
        <Field className="sm:max-w-[calc(50%-0.5rem)]">
          <FieldLabel htmlFor="customer-taxCondition">Condición frente al IVA</FieldLabel>
          <NativeSelect
            {...errorProps('taxCondition', fieldErrors)}
            value={values.taxCondition}
            onChange={(event) => change('taxCondition', event.target.value as TaxCondition)}
          >
            {TAX_CONDITIONS.map((condition) => (
              <option key={condition.value} value={condition.value}>
                {condition.label}
              </option>
            ))}
          </NativeSelect>
          {error('taxCondition')}
        </Field>
      </Section>

      <Section title="Contacto">
        <div className="grid gap-4 sm:grid-cols-2">
          <Field>
            <FieldLabel htmlFor="customer-email">
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
            <FieldLabel htmlFor="customer-phone">
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
          <FieldLabel htmlFor="customer-address">
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
          <FieldLabel htmlFor="customer-notes">
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
        <Button type="button" variant="outline" size="lg" className="h-11 md:h-9" onClick={requestClose}>
          Cancelar
        </Button>
        <Button type="submit" size="lg" className="h-11 md:h-9" disabled={isSubmitting}>
          {isSubmitting ? 'Guardando…' : submitLabel}
        </Button>
      </div>
    </form>
  )
}
