import { useRef, useState, type FormEvent } from 'react'
import { Plus, Trash2 } from 'lucide-react'
import { ApiError, messageFromError } from '../../../shared/api/client'
import { unitShort } from '../../../shared/lib/quantity'
import { useFormDialog, useUnsavedChanges } from '../../../shared/hooks/use-form-dialog'
import { Button } from '../../../shared/ui/button'
import { Field, FieldError, FieldLabel } from '../../../shared/ui/field'
import { Input } from '../../../shared/ui/input'
import { NativeSelect } from '../../../shared/ui/native-select'
import {
  emptyEntryLine,
  ENTRY_FIELD_ERRORS,
  entryFormSchema,
  type EntryFormValues,
  type EntryLineValues,
  type StockEntryRequest,
  type SupplierOption,
} from '../types'
import { ItemPicker } from './ItemPicker'

type EntryFormProps = {
  suppliers: readonly SupplierOption[]
  onSubmit: (request: StockEntryRequest) => Promise<unknown>
}

// Errors are keyed by their path in the form: "reference", "lines",
// "lines.0.quantity"…
type FieldErrors = Record<string, string | undefined>

export function EntryForm({ suppliers, onSubmit }: EntryFormProps) {
  const [values, setValues] = useState<EntryFormValues>({ supplierId: '', reference: '', lines: [emptyEntryLine(1)] })
  const { requestClose } = useFormDialog()
  useUnsavedChanges(values)
  const [nextKey, setNextKey] = useState(2)
  // The newest line's search takes the focus, so a scanner can go on reading
  // codes after "Agregar artículo".
  const [focusKey, setFocusKey] = useState(1)
  const quantityInputs = useRef(new Map<number, HTMLInputElement>())
  const [fieldErrors, setFieldErrors] = useState<FieldErrors>({})
  const [formError, setFormError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  function clearError(path: string) {
    setFieldErrors((current) => ({ ...current, [path]: undefined }))
  }

  function changeLine(index: number, change: Partial<EntryLineValues>, path: string) {
    setValues((current) => ({
      ...current,
      lines: current.lines.map((line, position) => (position === index ? { ...line, ...change } : line)),
    }))
    clearError(`lines.${index}.${path}`)
  }

  function pickItem(index: number, line: EntryLineValues, item: EntryLineValues['item']) {
    changeLine(index, { item }, 'item')
    if (item) {
      requestAnimationFrame(() => quantityInputs.current.get(line.key)?.focus())
    }
  }

  function addLine() {
    setValues((current) => ({ ...current, lines: [...current.lines, emptyEntryLine(nextKey)] }))
    setFocusKey(nextKey)
    setNextKey((key) => key + 1)
    clearError('lines')
  }

  function removeLine(index: number) {
    setValues((current) => ({ ...current, lines: current.lines.filter((_, position) => position !== index) }))
    // Line errors are keyed by position, so they would point at the wrong
    // line after a removal.
    setFieldErrors((current) =>
      Object.fromEntries(Object.entries(current).filter(([path]) => !path.startsWith('lines.'))),
    )
  }

  async function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    setFormError(null)

    const parsed = entryFormSchema.safeParse(values)
    if (!parsed.success) {
      const errors: FieldErrors = {}
      for (const issue of parsed.error.issues) {
        errors[issue.path.join('.')] ??= issue.message
      }
      setFieldErrors(errors)
      return
    }

    setIsSubmitting(true)
    try {
      await onSubmit(parsed.data)
    } catch (error) {
      const field = error instanceof ApiError ? ENTRY_FIELD_ERRORS[error.code] : undefined
      if (field) {
        setFieldErrors({ [field]: messageFromError(error) })
      } else {
        setFormError(messageFromError(error))
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  function errorId(path: string) {
    return `entry-${path.replaceAll('.', '-')}-error`
  }

  function errorProps(path: string) {
    return {
      'aria-invalid': Boolean(fieldErrors[path]),
      'aria-describedby': fieldErrors[path] ? errorId(path) : undefined,
    }
  }

  const error = (path: string) => <FieldError id={errorId(path)}>{fieldErrors[path]}</FieldError>

  return (
    <form noValidate onSubmit={handleSubmit} className="flex flex-col gap-6">
      <div className="grid gap-4 sm:grid-cols-2">
        <Field>
          <FieldLabel htmlFor="entry-supplierId">
            Proveedor <span className="font-normal text-muted-foreground">(opcional)</span>
          </FieldLabel>
          <NativeSelect
            id="entry-supplierId"
            {...errorProps('supplierId')}
            value={values.supplierId}
            onChange={(event) => {
              setValues((current) => ({ ...current, supplierId: event.target.value }))
              clearError('supplierId')
            }}
          >
            <option value="">Sin proveedor</option>
            {suppliers.map((supplier) => (
              <option key={supplier.id} value={supplier.id}>
                {supplier.name}
              </option>
            ))}
          </NativeSelect>
          {error('supplierId')}
        </Field>
        <Field>
          <FieldLabel htmlFor="entry-reference">
            Comprobante <span className="font-normal text-muted-foreground">(opcional)</span>
          </FieldLabel>
          <Input
            id="entry-reference"
            {...errorProps('reference')}
            autoComplete="off"
            placeholder="Remito o factura"
            value={values.reference}
            onChange={(event) => {
              setValues((current) => ({ ...current, reference: event.target.value }))
              clearError('reference')
            }}
          />
          {error('reference')}
        </Field>
      </div>

      <fieldset className="flex flex-col gap-3" aria-describedby={fieldErrors.lines ? errorId('lines') : undefined}>
        <legend className="mb-1 text-sm font-semibold text-muted-foreground">Artículos</legend>
        <ol className="flex flex-col gap-3">
          {values.lines.map((line, index) => {
            const number = index + 1
            const unit = line.item ? `(${unitShort(line.item.unit)})` : null
            return (
              <li key={line.key} aria-label={`Línea ${number}`} className="flex flex-col gap-3 rounded-xl border p-3">
                <Field>
                  <div className="flex items-center justify-between gap-2">
                    <FieldLabel htmlFor={`entry-line-${line.key}-item`}>Artículo {number}</FieldLabel>
                    {values.lines.length > 1 && (
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="size-11 text-destructive hover:bg-destructive/10 hover:text-destructive md:size-9"
                        aria-label={`Quitar la línea ${number}`}
                        title="Quitar"
                        onClick={() => removeLine(index)}
                      >
                        <Trash2 aria-hidden />
                      </Button>
                    )}
                  </div>
                  <ItemPicker
                    id={`entry-line-${line.key}-item`}
                    value={line.item}
                    autoFocus={line.key === focusKey}
                    onChange={(item) => pickItem(index, line, item)}
                    invalid={Boolean(fieldErrors[`lines.${index}.item`])}
                    describedBy={fieldErrors[`lines.${index}.item`] ? errorId(`lines.${index}.item`) : undefined}
                  />
                  {error(`lines.${index}.item`)}
                </Field>
                <div className="grid gap-3 sm:grid-cols-2">
                  <Field>
                    <FieldLabel htmlFor={`entry-line-${line.key}-quantity`}>
                      Cantidad {unit && <span className="font-normal text-muted-foreground">{unit}</span>}
                    </FieldLabel>
                    <Input
                      id={`entry-line-${line.key}-quantity`}
                      ref={(element) => {
                        if (element) quantityInputs.current.set(line.key, element)
                        else quantityInputs.current.delete(line.key)
                      }}
                      {...errorProps(`lines.${index}.quantity`)}
                      inputMode="decimal"
                      autoComplete="off"
                      value={line.quantity}
                      onChange={(event) => changeLine(index, { quantity: event.target.value }, 'quantity')}
                    />
                    {error(`lines.${index}.quantity`)}
                  </Field>
                  <Field>
                    <FieldLabel htmlFor={`entry-line-${line.key}-unitCost`}>
                      Costo unitario <span className="font-normal text-muted-foreground">(opcional)</span>
                    </FieldLabel>
                    <div className="relative">
                      <span className="pointer-events-none absolute top-1/2 left-3 -translate-y-1/2 text-muted-foreground">$</span>
                      <Input
                        id={`entry-line-${line.key}-unitCost`}
                        {...errorProps(`lines.${index}.unitCost`)}
                        inputMode="decimal"
                        autoComplete="off"
                        placeholder="0,00"
                        className="pl-7"
                        value={line.unitCost}
                        onChange={(event) => changeLine(index, { unitCost: event.target.value }, 'unitCost')}
                      />
                    </div>
                    {error(`lines.${index}.unitCost`)}
                  </Field>
                </div>
              </li>
            )
          })}
        </ol>
        {error('lines')}
        <Button type="button" variant="outline" size="lg" className="h-11 self-start md:h-9" onClick={addLine}>
          <Plus aria-hidden />
          Agregar artículo
        </Button>
        <p className="text-sm text-muted-foreground">El costo unitario que se ingresa pasa a ser el costo del artículo.</p>
      </fieldset>

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
          {isSubmitting ? 'Guardando…' : 'Registrar ingreso'}
        </Button>
      </div>
    </form>
  )
}
