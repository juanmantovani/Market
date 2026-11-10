import { Dialog, DialogClose, DialogContent, DialogDescription, DialogTitle } from '@/shared/ui/dialog'
import { Button } from '@/shared/ui/button'

type ConfirmDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description: string
  confirmLabel: string
  onConfirm: () => void
  isPending?: boolean
}

// Asks before a destructive action. Cancel is the safe default and closes
// without side effects.
export function ConfirmDialog({
  open,
  onOpenChange,
  title,
  description,
  confirmLabel,
  onConfirm,
  isPending = false,
}: ConfirmDialogProps) {
  return (
    <Dialog open={open} onOpenChange={(next) => onOpenChange(next)}>
      <DialogContent role="alertdialog" className="flex max-w-md flex-col gap-4 p-6">
        <div className="flex flex-col gap-2">
          <DialogTitle>{title}</DialogTitle>
          <DialogDescription>{description}</DialogDescription>
        </div>
        <div className="flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <DialogClose
            render={<Button type="button" variant="outline" size="lg" className="h-11 md:h-9" />}
          >
            Cancelar
          </DialogClose>
          <Button
            type="button"
            variant="destructive"
            size="lg"
            className="h-11 md:h-9"
            disabled={isPending}
            onClick={onConfirm}
          >
            {confirmLabel}
          </Button>
        </div>
      </DialogContent>
    </Dialog>
  )
}
