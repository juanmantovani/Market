import { X } from 'lucide-react'
import type { ReactNode } from 'react'

import { Button } from '@/shared/ui/button'
import { Dialog, DialogClose, DialogContent, DialogDescription, DialogTitle } from '@/shared/ui/dialog'

type FormDialogProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description?: string
  children: ReactNode
}

// The modal every entity uses to create and edit. Focus is trapped while
// open, Escape or the backdrop close it, and long forms scroll inside it.
export function FormDialog({ open, onOpenChange, title, description, children }: FormDialogProps) {
  return (
    <Dialog open={open} onOpenChange={(next) => onOpenChange(next)}>
      <DialogContent className="flex max-w-lg flex-col rounded-2xl">
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
        <div className="px-5 py-5">{children}</div>
      </DialogContent>
    </Dialog>
  )
}
