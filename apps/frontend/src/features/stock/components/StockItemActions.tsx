import { useState } from 'react'
import { History, PackageOpen, SlidersHorizontal } from 'lucide-react'
import { formatUnitQuantity } from '../../../shared/lib/quantity'
import { Button } from '../../../shared/ui/button'
import { FormDialog } from '../../../shared/ui/form-dialog'
import { useToast } from '../../../shared/ui/use-toast'
import { useRegisterStockAdjustment, useRegisterStockInitial } from '../hooks/use-stock'
import type { AdjustmentFields, InitialStockFields, StockItem, StockLevel } from '../types'
import { AdjustmentForm } from './AdjustmentForm'
import { InitialStockForm } from './InitialStockForm'
import { MovementHistory } from './MovementHistory'

type StockItemActionsProps = {
  item: StockItem
  // In thousandths of the item's unit; undefined while it loads.
  quantity: number | undefined
  // An item without movements opens its stock instead of adjusting it.
  hasMovements: boolean
  supplierName: (supplierId: string) => string
  // Icon buttons for a table row, or buttons with text for a card.
  layout: 'icons' | 'buttons'
}

// `open` is kept apart from what each overlay shows, so its content stays in
// place while the closing animation runs instead of flashing empty.
// `opensStock` is fixed when the adjustment dialog opens: once the first count
// is saved the item has movements, and the dialog must not turn into an
// adjustment while it closes.
type Dialog = { open: boolean; formKey: number; opensStock: boolean }
const CLOSED: Dialog = { open: false, formKey: 0, opensStock: false }

// What Artículos offers for the stock of one item: adjust it (or load the
// first count) and see its movements.
export function StockItemActions({ item, quantity, hasMovements, supplierName, layout }: StockItemActionsProps) {
  const [adjustment, setAdjustment] = useState(CLOSED)
  const [history, setHistory] = useState(CLOSED)
  const registerAdjustment = useRegisterStockAdjustment()
  const registerInitial = useRegisterStockInitial()
  const toast = useToast()

  const level: StockLevel | null =
    quantity === undefined
      ? null
      : {
          itemId: item.id,
          code: item.code,
          name: item.name,
          categoryId: item.categoryId,
          unit: item.unit,
          minStock: item.minStock,
          quantity,
          lowStock: false,
          updatedAt: null,
        }
  const opensStock = !hasMovements
  const adjustLabel = opensStock ? 'Cargar stock inicial' : 'Ajustar stock'

  function close(setter: typeof setAdjustment) {
    setter((current) => ({ ...current, open: false }))
  }

  function open(setter: typeof setAdjustment) {
    setter((current) => ({ open: true, formKey: current.formKey + 1, opensStock }))
  }

  async function adjust(fields: AdjustmentFields) {
    const movement = await registerAdjustment.mutateAsync({ itemId: item.id, ...fields })
    close(setAdjustment)
    toast.success('Stock ajustado', `Ahora hay ${formatUnitQuantity(movement.balanceAfter, item.unit)} de «${item.name}».`)
  }

  async function openStock(fields: InitialStockFields) {
    const movement = await registerInitial.mutateAsync({ itemId: item.id, ...fields })
    close(setAdjustment)
    toast.success('Stock inicial cargado', `Ahora hay ${formatUnitQuantity(movement.balanceAfter, item.unit)} de «${item.name}».`)
  }

  const buttons =
    layout === 'icons' ? (
      <>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          aria-label={`${adjustLabel} de ${item.name}`}
          title={adjustLabel}
          disabled={level === null}
          onClick={() => open(setAdjustment)}
        >
          {opensStock ? <PackageOpen aria-hidden /> : <SlidersHorizontal aria-hidden />}
        </Button>
        <Button
          type="button"
          variant="ghost"
          size="icon"
          aria-label={`Movimientos de ${item.name}`}
          title="Movimientos"
          onClick={() => open(setHistory)}
        >
          <History aria-hidden />
        </Button>
      </>
    ) : (
      <>
        <Button
          type="button"
          variant="outline"
          className="h-11 flex-1"
          aria-label={`${adjustLabel} de ${item.name}`}
          disabled={level === null}
          onClick={() => open(setAdjustment)}
        >
          {opensStock ? <PackageOpen aria-hidden /> : <SlidersHorizontal aria-hidden />}
          {opensStock ? 'Stock inicial' : 'Ajustar stock'}
        </Button>
        <Button
          type="button"
          variant="outline"
          className="h-11 flex-1"
          aria-label={`Movimientos de ${item.name}`}
          onClick={() => open(setHistory)}
        >
          <History aria-hidden />
          Movimientos
        </Button>
      </>
    )

  return (
    <>
      {buttons}

      <FormDialog
        open={adjustment.open}
        onOpenChange={(next) => {
          if (!next) close(setAdjustment)
        }}
        title={adjustment.opensStock ? 'Cargar stock inicial' : 'Ajustar stock'}
        description={item.name}
      >
        {level &&
          (adjustment.opensStock ? (
            <InitialStockForm key={adjustment.formKey} item={item} onSubmit={openStock} />
          ) : (
            <AdjustmentForm key={adjustment.formKey} level={level} onSubmit={adjust} />
          ))}
      </FormDialog>

      <FormDialog
        open={history.open}
        onOpenChange={(next) => {
          if (!next) close(setHistory)
        }}
        size="wide"
        title="Movimientos"
        description={item.name}
      >
        {level && <MovementHistory key={history.formKey} level={level} supplierName={supplierName} />}
      </FormDialog>
    </>
  )
}
