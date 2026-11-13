import { useReturnFocus } from '../shared/hooks/use-return-focus'
import { Button } from '../shared/ui/button'
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogTitle } from '../shared/ui/dialog'
import { SHORTCUT_GROUPS } from './shortcuts'

type ShortcutsDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
}

export default function ShortcutsDialog({ open, onOpenChange }: ShortcutsDialogProps) {
  const finalFocus = useReturnFocus(open)

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent finalFocus={finalFocus} className="flex flex-col gap-5 p-6">
        <div className="flex flex-col gap-1">
          <DialogTitle>Atajos de teclado</DialogTitle>
          <DialogDescription>La venta se puede operar completa sin el mouse.</DialogDescription>
        </div>

        {SHORTCUT_GROUPS.map((group) => (
          <section key={group.title} aria-label={group.title} className="flex flex-col gap-2">
            <h3 className="text-sm font-medium text-muted-foreground">{group.title}</h3>
            <dl className="flex flex-col divide-y rounded-lg border">
              {group.shortcuts.map(({ keys, description }) => (
                <div key={description} className="flex items-center justify-between gap-4 px-3 py-2.5">
                  <dt className="text-sm">{description}</dt>
                  <dd className="flex shrink-0 gap-1">
                    {keys.map((key) => (
                      <kbd key={key} className="min-w-7 rounded border bg-muted px-1.5 py-0.5 text-center text-xs font-medium">
                        {key}
                      </kbd>
                    ))}
                  </dd>
                </div>
              ))}
            </dl>
          </section>
        ))}

        <div className="flex justify-end">
          <DialogClose render={<Button type="button" variant="outline" size="lg" className="h-11 md:h-9" />}>Cerrar</DialogClose>
        </div>
      </DialogContent>
    </Dialog>
  )
}
