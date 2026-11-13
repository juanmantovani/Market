import { X } from 'lucide-react'
import { useCallback, useMemo, useRef, useState, type ReactNode } from 'react'

import { FormDialogContext } from '@/shared/hooks/use-form-dialog'
import { useReturnFocus } from '@/shared/hooks/use-return-focus'
import { cn } from '@/shared/lib/utils'
import { Button } from '@/shared/ui/button'
import { ConfirmDialog } from '@/shared/ui/confirm-dialog'
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogTitle } from '@/shared/ui/dialog'

type FormDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description?: string
  // `wide` fits forms laid out in two columns.
  size?: 'default' | 'wide'
  children: ReactNode
}

// The modal every entity uses to create and edit. Focus is trapped while
// open, Escape, the X, the backdrop and the form's Cancel button close it, and
// long forms scroll inside it. A form that reports unsaved changes through
// `useUnsavedChanges` asks before discarding them; closing it from the page
// (after saving) never asks.
export function FormDialog({ open, onOpenChange, title, description, size = 'default', children }: FormDialogProps) {
  const finalFocus = useReturnFocus(open)
  const hasChanges = useRef(false)
  const [asking, setAsking] = useState(false)

  if (!open && asking) setAsking(false)

  const setHasChanges = useCallback((value: boolean) => {
    hasChanges.current = value
  }, [])

  const requestClose = useCallback(() => {
    if (hasChanges.current) {
      setAsking(true)
    } else {
      onOpenChange(false)
    }
  }, [onOpenChange])

  const controls = useMemo(() => ({ requestClose, setHasChanges }), [requestClose, setHasChanges])

  return (
    <Dialog open={open} onOpenChange={(next) => (next ? onOpenChange(true) : requestClose())}>
      <DialogContent
        finalFocus={finalFocus}
        className={cn('flex flex-col rounded-2xl', size === 'wide' ? 'max-w-2xl' : 'max-w-lg')}
      >
        <header className="flex items-start justify-between gap-4 border-b px-5 py-4">
          <div className="flex flex-col gap-1">
            <DialogTitle>{title}</DialogTitle>
            {description && <DialogDescription>{description}</DialogDescription>}
          </div>
          <DialogClose
            render={
              <Button
                type="button"
                variant="ghost"
                size="icon-lg"
                className="-mr-2 size-11 text-muted-foreground md:size-9"
                aria-label="Cerrar"
              />
            }
          >
            <X aria-hidden />
          </DialogClose>
        </header>
        <FormDialogContext.Provider value={controls}>
          <div className="px-5 py-5">{children}</div>
        </FormDialogContext.Provider>

        <ConfirmDialog
          open={open && asking}
          onOpenChange={setAsking}
          title="¿Descartar los cambios?"
          description="Lo ingresado en el formulario se perderá."
          confirmLabel="Descartar"
          cancelLabel="Seguir editando"
          onConfirm={() => {
            setAsking(false)
            onOpenChange(false)
          }}
        />
      </DialogContent>
    </Dialog>
  )
}
