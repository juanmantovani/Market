import { useState } from 'react'
import { Minus, Plus, Trash2 } from 'lucide-react'
import { formatMoney } from '../../../shared/lib/money'
import { formatQuantity, formatUnitQuantity, parseQuantity, unitShort } from '../../../shared/lib/quantity'
import { Button } from '../../../shared/ui/button'
import { FieldError } from '../../../shared/ui/field'
import { Input } from '../../../shared/ui/input'
import { isValidQuantity, lineAmount, quantityStep } from '../lib/cart'
import type { CartLine } from '../types'

type CartLineRowProps = {
  line: CartLine
  onQuantityChange: (itemId: string, quantity: number) => void
  onRemove: (itemId: string) => void
}

export function CartLineRow({ line, onQuantityChange, onRemove }: CartLineRowProps) {
  const [quantityError, setQuantityError] = useState<string | null>(null)
  const step = quantityStep(line.unit)
  const stepText = formatUnitQuantity(step, line.unit)
  const errorId = `cart-quantity-error-${line.itemId}`
  const total = lineAmount(line)

  function change(quantity: number) {
    setQuantityError(null)
    onQuantityChange(line.itemId, quantity)
  }

  return (
    <li className="flex flex-col gap-2 py-3">
      <div className="flex items-start justify-between gap-3">
        <div className="min-w-0">
          <p className="truncate text-sm font-medium">{line.name}</p>
          <p className="text-xs text-muted-foreground">
            {formatMoney(line.unitPrice)} {line.unit === 'unit' ? 'c/u' : `/ ${unitShort(line.unit)}`}
          </p>
        </div>
        <span className="shrink-0 text-sm font-semibold tabular-nums">{formatMoney(total)}</span>
      </div>
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center gap-1">
          <Button
            type="button"
            variant="outline"
            size="icon"
            className="size-11 md:size-9"
            aria-label={`Restar ${stepText} de ${line.name}`}
            disabled={line.quantity <= step}
            onClick={() => change(line.quantity - step)}
          >
            <Minus aria-hidden />
          </Button>
          <QuantityInput
            key={line.quantity}
            line={line}
            errorId={errorId}
            invalid={quantityError !== null}
            onCommit={change}
            onError={setQuantityError}
          />
          <Button
            type="button"
            variant="outline"
            size="icon"
            className="size-11 md:size-9"
            aria-label={`Sumar ${stepText} a ${line.name}`}
            onClick={() => change(line.quantity + step)}
          >
            <Plus aria-hidden />
          </Button>
          {line.unit !== 'unit' && <span className="ml-1 text-sm text-muted-foreground">{unitShort(line.unit)}</span>}
        </div>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          className="size-11 text-muted-foreground hover:text-destructive md:size-9"
          aria-label={`Quitar ${line.name}`}
          onClick={() => onRemove(line.itemId)}
        >
          <Trash2 aria-hidden />
        </Button>
      </div>
      <FieldError id={errorId}>{quantityError}</FieldError>
    </li>
  )
}

type QuantityInputProps = {
  line: CartLine
  errorId: string
  invalid: boolean
  onCommit: (quantity: number) => void
  onError: (message: string | null) => void
}

// Keeps what is being typed until Enter or leaving the field; keyed by the
// line's quantity, so + and − show the new value right away.
function QuantityInput({ line, errorId, invalid, onCommit, onError }: QuantityInputProps) {
  const [draft, setDraft] = useState(formatQuantity(line.quantity))

  function commit() {
    const quantity = parseQuantity(draft, line.unit !== 'unit')
    if (quantity === null || !isValidQuantity(quantity)) {
      onError(
        line.unit === 'unit'
          ? 'Ingresar una cantidad entera mayor a cero'
          : `Ingresar una cantidad mayor a cero, por ejemplo 0,750 ${unitShort(line.unit)}`,
      )
      return
    }
    onError(null)
    if (quantity !== line.quantity) onCommit(quantity)
  }

  return (
    <Input
      aria-label={`Cantidad de ${line.name}`}
      inputMode={line.unit === 'unit' ? 'numeric' : 'decimal'}
      autoComplete="off"
      value={draft}
      aria-invalid={invalid}
      aria-describedby={invalid ? errorId : undefined}
      onChange={(event) => setDraft(event.target.value)}
      onBlur={commit}
      onKeyDown={(event) => {
        if (event.key === 'Enter') {
          event.preventDefault()
          commit()
        } else if (event.key === 'Escape' && invalid) {
          event.preventDefault()
          setDraft(formatQuantity(line.quantity))
          onError(null)
        }
      }}
      className="h-11 w-20 text-center tabular-nums md:h-9"
    />
  )
}
