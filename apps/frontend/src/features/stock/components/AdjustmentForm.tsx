import { useState, type FormEvent } from 'react'
import { ApiError, messageFromError } from '../../../shared/api/client'
import { cn } from '../../../shared/lib/utils'
import { formatUnitQuantity, parseQuantity, unitShort } from '../../../shared/lib/quantity'
import { Button } from '../../../shared/ui/button'
import { Field, FieldError, FieldLabel } from '../../../shared/ui/field'
import { Input } from '../../../shared/ui/input'
import {
  ADJUSTMENT_FIELD_ERRORS,
  ADJUSTMENT_REASONS,
  adjustmentFormSchema,
  type AdjustmentField,
  type AdjustmentFields,
  type AdjustmentFormValues,
  type StockLevel,
} from '../types'

type AdjustmentFormProps = {
  level: StockLevel
  onSubmit: (fields: AdjustmentFields) => Promise<unknown>
  onCancel: () => void
}

type FieldErrors = Partial<Record<AdjustmentField, string>>

const DIRECTIONS = [
  { value: 'in', label: 'Sumar' },
  { value: 'out', label: 'Restar' },
] as const

function errorProps(field: AdjustmentField, errors: FieldErrors) {
  return {
    id: `adjustment-${field}`,
    'aria-invalid': Boolean(errors[field]),
    'aria-describedby': errors[field] ? `adjustment-${field}-error` : undefined,
  }
}

export function AdjustmentForm({ level, onSubmit, onCancel }: AdjustmentFormProps) {
  const [values, setValues] = useState<AdjustmentFormValues>({ direction: 'out', quantity: '', reason: '' })
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  function change<F extends AdjustmentField>(field: F, value: AdjustmentFormValues[F]) {
    setValues((current) => ({ ...current, [field]: value }))
    setFieldErrors((current) => ({ ...current, [field]: undefined, ...(field === 'direction' ? { quantity: undefined } : {}) }))
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const parsed = adjustmentFormSchema(level).safeParse(values)
    if (!parsed.success) {
      const errors: FieldErrors = {}
      for (const issue of parsed.error.issues) {
        const field = issue.path[0] as AdjustmentField
        errors[field] ??= issue.message
      }
      setFieldErrors(errors)
      return
    }

    setIsSubmitting(true)
    try {
      await onSubmit(parsed.data)
    } catch (error) {
      const field = error instanceof ApiError ? ADJUSTMENT_FIELD_ERRORS[error.code] : undefined
      if (field) {
        setFieldErrors({ [field]: messageFromError(error) })
      } else {
        setFormError(messageFromError(error))
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  const typed = parseQuantity(values.quantity, level.unit !== 'unit')
  const after = typed ? level.quantity + (values.direction === 'out' ? -typed : typed) : null
  const error = (field: AdjustmentField) => <FieldError id={`adjustment-${field}-error`}>{fieldErrors[field]}</FieldError>

  return (
    <form noValidate onSubmit={handleSubmit} className="flex flex-col gap-5">
      <p className="rounded-lg bg-muted px-3 py-2 text-sm">
        Stock actual: <strong className="tabular-nums">{formatUnitQuantity(level.quantity, level.unit)}</strong>
      </p>

      <fieldset className="flex flex-col gap-2">
        <legend className="mb-1 text-sm font-medium">Tipo de ajuste</legend>
        <div className="grid grid-cols-2 gap-2">
          {DIRECTIONS.map((direction) => (
            <label
              key={direction.value}
              className={cn(
                'flex min-h-11 cursor-pointer items-center justify-center gap-2 rounded-lg border px-3 text-sm font-medium transition-colors has-focus-visible:ring-3 has-focus-visible:ring-ring/50',
                values.direction === direction.value ? 'border-primary bg-primary/10 text-foreground' : 'hover:bg-muted',
              )}
            >
              <input
                type="radio"
                name="adjustment-direction"
                value={direction.value}
                className="size-4 accent-primary"
                checked={values.direction === direction.value}
                onChange={() => change('direction', direction.value)}
              />
              {direction.label}
            </label>
          ))}
        </div>
      </fieldset>

      <Field>
        <FieldLabel htmlFor="adjustment-quantity">
          Cantidad <span className="font-normal text-muted-foreground">({unitShort(level.unit)})</span>
        </FieldLabel>
        <Input
          {...errorProps('quantity', fieldErrors)}
          autoFocus
          inputMode="decimal"
          autoComplete="off"
          value={values.quantity}
          onChange={(event) => change('quantity', event.target.value)}
        />
        {error('quantity')}
        {after !== null && !fieldErrors.quantity && (
          <p aria-live="polite" className={cn('text-sm', after < 0 ? 'text-destructive' : 'text-muted-foreground')}>
            Queda en <span className="tabular-nums">{formatUnitQuantity(after, level.unit)}</span>
          </p>
        )}
      </Field>

      <Field>
        <FieldLabel htmlFor="adjustment-reason">Motivo</FieldLabel>
        <Input
          {...errorProps('reason', fieldErrors)}
          list="adjustment-reasons"
          autoComplete="off"
          placeholder="Por ejemplo: conteo o rotura"
          value={values.reason}
          onChange={(event) => change('reason', event.target.value)}
        />
        <datalist id="adjustment-reasons">
          {ADJUSTMENT_REASONS.map((reason) => (
            <option key={reason} value={reason} />
          ))}
        </datalist>
        {error('reason')}
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
          {isSubmitting ? 'Guardando…' : 'Guardar ajuste'}
        </Button>
      </div>
    </form>
  )
}
