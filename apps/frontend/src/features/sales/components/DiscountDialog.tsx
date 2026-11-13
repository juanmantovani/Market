import { useState, type FormEvent } from 'react'
import { centsToInput, formatMoney, parseMoney } from '../../../shared/lib/money'
import { parsePercent, percentToInput } from '../../../shared/lib/percent'
import { Button } from '../../../shared/ui/button'
import { Field, FieldError, FieldLabel } from '../../../shared/ui/field'
import { FormDialog } from '../../../shared/ui/form-dialog'
import { Input } from '../../../shared/ui/input'
import { NativeSelect } from '../../../shared/ui/native-select'
import { discountOff, PERCENT_SCALE } from '../lib/discounts'
import type { Discount } from '../types'

type DiscountDialogProps = {
  open: boolean
  title: string
  // What the discount is taken from: the line's amount, or what the lines
  // add up to after their own discounts. A discount can't be more.
  base: number
  current: Discount | null
  // null takes the discount off.
  onApply: (discount: Discount | null) => void
  onClose: () => void
}

export function DiscountDialog({ open, title, base, current, onApply, onClose }: DiscountDialogProps) {
  return (
    <FormDialog open={open} onOpenChange={(next) => !next && onClose()} title={title}>
      {open && <DiscountForm base={base} current={current} onApply={onApply} onCancel={onClose} />}
    </FormDialog>
  )
}

type DiscountFormProps = {
  base: number
  current: Discount | null
  onApply: (discount: Discount | null) => void
  onCancel: () => void
}

function DiscountForm({ base, current, onApply, onCancel }: DiscountFormProps) {
  const [type, setType] = useState<Discount['type']>(current?.type ?? 'percent')
  const [text, setText] = useState(() => {
    if (!current) return ''
    return current.type === 'percent' ? percentToInput(current.value) : centsToInput(current.value)
  })
  const [error, setError] = useState<string | null>(null)
  const value = text.trim() === '' ? null : type === 'percent' ? parsePercent(text) : parseMoney(text)
  const valid = value !== null && value > 0 && (type === 'percent' ? value <= PERCENT_SCALE : value <= base)
  const off = valid ? discountOff({ type, value }, base) : null

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (!valid) {
      setError(
        type === 'percent'
          ? 'Ingresar un porcentaje entre 0,01 y 100, por ejemplo 12,5'
          : `Ingresar un monto mayor a cero y de hasta ${formatMoney(base)}`,
      )
      return
    }
    onApply({ type, value })
  }

  return (
    <form noValidate onSubmit={handleSubmit} className="flex flex-col gap-5">
      <div className="grid grid-cols-[minmax(0,1fr)_minmax(0,1fr)] gap-3">
        <Field>
          <FieldLabel htmlFor="discount-type">Tipo de descuento</FieldLabel>
          <NativeSelect
            id="discount-type"
            value={type}
            onChange={(event) => {
              setType(event.target.value as Discount['type'])
              setText('')
              setError(null)
            }}
          >
            <option value="percent">Porcentaje</option>
            <option value="amount">Monto</option>
          </NativeSelect>
        </Field>
        <Field>
          <FieldLabel htmlFor="discount-value">{type === 'percent' ? 'Porcentaje (%)' : 'Monto ($)'}</FieldLabel>
          <Input
            id="discount-value"
            autoFocus
            inputMode="decimal"
            autoComplete="off"
            value={text}
            aria-invalid={Boolean(error)}
            aria-describedby={error ? 'discount-error' : undefined}
            onFocus={(event) => event.currentTarget.select()}
            onChange={(event) => {
              setText(event.target.value)
              setError(null)
            }}
          />
        </Field>
      </div>
      <FieldError id="discount-error">{error}</FieldError>

      <p aria-live="polite" className="rounded-xl bg-muted px-4 py-3 text-sm">
        {off !== null ? `Descuento: ${formatMoney(off)} de ${formatMoney(base)}` : `Importe sin descuento: ${formatMoney(base)}`}
      </p>

      <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
        {current && (
          <Button type="button" variant="ghost" size="lg" className="h-11 text-muted-foreground md:h-9 sm:mr-auto" onClick={() => onApply(null)}>
            Quitar descuento
          </Button>
        )}
        <Button type="button" variant="outline" size="lg" className="h-11 md:h-9" onClick={onCancel}>
          Cancelar
        </Button>
        <Button type="submit" size="lg" className="h-11 md:h-9">
          Aplicar
        </Button>
      </div>
    </form>
  )
}
