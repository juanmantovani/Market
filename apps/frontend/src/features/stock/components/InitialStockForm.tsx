import { useState, type FormEvent } from 'react'
import { ApiError, messageFromError } from '../../../shared/api/client'
import { useFormDialog, useUnsavedChanges } from '../../../shared/hooks/use-form-dialog'
import { unitShort } from '../../../shared/lib/quantity'
import { Button } from '../../../shared/ui/button'
import { Field, FieldDescription, FieldError, FieldLabel } from '../../../shared/ui/field'
import { Input } from '../../../shared/ui/input'
import {
  INITIAL_STOCK_FIELD_ERRORS,
  initialStockFormSchema,
  type InitialStockFields,
  type InitialStockFormValues,
  type StockItem,
} from '../types'

type InitialStockFormProps = {
  item: StockItem
  onSubmit: (fields: InitialStockFields) => Promise<unknown>
}

// Opens the stock of an item that has no movements yet, so the first count
// is not mixed up with later breakages or shortfalls.
export function InitialStockForm({ item, onSubmit }: InitialStockFormProps) {
  const [values, setValues] = useState<InitialStockFormValues>({ quantity: '' })
  const { requestClose } = useFormDialog()
  useUnsavedChanges(values)
  const [fieldError, setFieldError] = useState<string | null>(null)
  const [formError, setFormError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const parsed = initialStockFormSchema(item.unit).safeParse(values)
    if (!parsed.success) {
      setFieldError(parsed.error.issues[0].message)
      return
    }

    setIsSubmitting(true)
    try {
      await onSubmit(parsed.data)
    } catch (error) {
      if (error instanceof ApiError && INITIAL_STOCK_FIELD_ERRORS[error.code]) {
        setFieldError(messageFromError(error))
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
        <FieldLabel htmlFor="initial-quantity">
          Cantidad que hay hoy <span className="font-normal text-muted-foreground">({unitShort(item.unit)})</span>
        </FieldLabel>
        <Input
          id="initial-quantity"
          autoFocus
          inputMode="decimal"
          autoComplete="off"
          value={values.quantity}
          aria-invalid={Boolean(fieldError)}
          aria-describedby={fieldError ? 'initial-quantity-error' : 'initial-quantity-hint'}
          onChange={(event) => {
            setValues({ quantity: event.target.value })
            setFieldError(null)
          }}
        />
        <FieldError id="initial-quantity-error">{fieldError}</FieldError>
        {!fieldError && (
          <FieldDescription id="initial-quantity-hint">
            Es el primer movimiento del artículo: después, los cambios se registran como ingresos o ajustes.
          </FieldDescription>
        )}
      </Field>

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
          {isSubmitting ? 'Guardando…' : 'Cargar stock inicial'}
        </Button>
      </div>
    </form>
  )
}
