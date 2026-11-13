import { useState } from 'react'
import { PackagePlus } from 'lucide-react'
import { Button } from '../../../shared/ui/button'
import { FormDialog } from '../../../shared/ui/form-dialog'
import { useToast } from '../../../shared/ui/use-toast'
import { useRegisterStockEntry } from '../hooks/use-stock'
import type { StockEntryRequest, SupplierOption } from '../types'
import { EntryForm } from './EntryForm'

type ReceiveStockProps = {
  suppliers: readonly SupplierOption[]
}

// "Ingresar mercadería": the button and its dialog. `open` is kept apart from
// the form key, so the content stays in place while the closing animation
// runs instead of flashing empty.
export function ReceiveStock({ suppliers }: ReceiveStockProps) {
  const [dialog, setDialog] = useState({ open: false, formKey: 0 })
  const registerEntry = useRegisterStockEntry()
  const toast = useToast()

  async function receive(request: StockEntryRequest) {
    const { movements } = await registerEntry.mutateAsync(request)
    setDialog((current) => ({ ...current, open: false }))
    const count = movements.length === 1 ? '1 artículo' : `${movements.length} artículos`
    toast.success('Ingreso registrado', `Se actualizó el stock de ${count}.`)
  }

  return (
    <>
      <Button
        type="button"
        variant="outline"
        size="lg"
        className="h-11 md:h-9"
        onClick={() => setDialog((current) => ({ open: true, formKey: current.formKey + 1 }))}
      >
        <PackagePlus aria-hidden />
        Ingresar mercadería
      </Button>
      <FormDialog
        open={dialog.open}
        onOpenChange={(open) => {
          if (!open) setDialog((current) => ({ ...current, open: false }))
        }}
        size="wide"
        title="Ingresar mercadería"
      >
        <EntryForm key={dialog.formKey} suppliers={suppliers} onSubmit={receive} />
      </FormDialog>
    </>
  )
}
