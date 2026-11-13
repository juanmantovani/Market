import { useState, type FormEvent } from 'react'
import { formatMoney } from '../../../shared/lib/money'
import { parseQuantity, unitShort } from '../../../shared/lib/quantity'
import { Button } from '../../../shared/ui/button'
import { Field, FieldDescription, FieldError, FieldLabel } from '../../../shared/ui/field'
import { FormDialog } from '../../../shared/ui/form-dialog'
import { Input } from '../../../shared/ui/input'
import { isValidQuantity, lineAmount } from '../lib/cart'
import type { CatalogItem } from '../types'

type QuantityDialogProps = {
  open: boolean
  // The item sold by weight or length being added. It stays set while the
  // dialog closes, so the closing animation keeps its content.
  item: CatalogItem | null
  onAdd: (item: CatalogItem, quantity: number) => void
  onClose: () => void
}

// Items sold by weight ask for the quantity when they are added.
export function QuantityDialog({ open, item, onAdd, onClose }: QuantityDialogProps) {
  return (
    <FormDialog
      open={open && item !== null}
      onOpenChange={(next) => !next && onClose()}
      title={item ? `Cantidad de ${item.name}` : 'Cantidad'}
    >
      {item && <QuantityForm item={item} onAdd={onAdd} onCancel={onClose} />}
    </FormDialog>
  )
}

function QuantityForm({ item, onAdd, onCancel }: { item: CatalogItem; onAdd: QuantityDialogProps['onAdd']; onCancel: () => void }) {
  const [text, setText] = useState('')
  const [error, setError] = useState<string | null>(null)
  const unit = unitShort(item.unit)
  const quantity = parseQuantity(text, true)
  const preview = quantity !== null && isValidQuantity(quantity) ? lineAmount({ unitPrice: item.salePrice, quantity }) : null

  function handleSubmit(event: FormEvent<HTMLFormElement>) {
    event.preventDefault()
    if (quantity === null || !isValidQuantity(quantity)) {
      setError(`Ingresar una cantidad mayor a cero, por ejemplo 0,750 ${unit}`)
      return
    }
    onAdd(item, quantity)
  }

  return (
    <form noValidate onSubmit={handleSubmit} className="flex flex-col gap-5">
      <Field>
        <FieldLabel htmlFor="pos-quantity">Cantidad ({unit})</FieldLabel>
        <Input
          id="pos-quantity"
          autoFocus
          inputMode="decimal"
          autoComplete="off"
          value={text}
          aria-invalid={Boolean(error)}
          aria-describedby={error ? 'pos-quantity-error' : 'pos-quantity-hint'}
          onChange={(event) => {
            setText(event.target.value)
            setError(null)
          }}
        />
        {error ? (
          <FieldError id="pos-quantity-error">{error}</FieldError>
        ) : (
          <FieldDescription id="pos-quantity-hint">
            {formatMoney(item.salePrice)} por {unit}
            {preview !== null && <> · Importe: {formatMoney(preview)}</>}
          </FieldDescription>
        )}
      </Field>
      <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
        <Button type="button" variant="outline" size="lg" className="h-11 md:h-9" onClick={onCancel}>
          Cancelar
        </Button>
        <Button type="submit" size="lg" className="h-11 md:h-9">
          Agregar
        </Button>
      </div>
    </form>
  )
}
