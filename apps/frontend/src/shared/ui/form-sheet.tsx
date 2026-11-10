import { Dialog as DialogPrimitive } from '@base-ui/react/dialog'
import { X } from 'lucide-react'
import type { ReactNode } from 'react'

import { Button } from '@/shared/ui/button'

type FormSheetProps = {
  open: boolean
  onOpenChange: (open: boolean) => void
  title: string
  description?: string
  children: ReactNode
}

// A side panel on desktop and a full-screen page on a phone, for create and
// edit forms. Focus is trapped while open and Escape closes it.
export function FormSheet({ open, onOpenChange, title, description, children }: FormSheetProps) {
  return (
    <DialogPrimitive.Root open={open} onOpenChange={(next) => onOpenChange(next)}>
      <DialogPrimitive.Portal>
        <DialogPrimitive.Backdrop className="fixed inset-0 z-50 bg-foreground/40 transition-opacity duration-200 data-closed:opacity-0 data-open:opacity-100" />
        <DialogPrimitive.Popup className="fixed inset-0 z-50 flex flex-col bg-card text-card-foreground shadow-xl outline-none transition-transform duration-200 data-closed:translate-x-full md:inset-y-0 md:right-0 md:left-auto md:w-full md:max-w-md md:border-l">
          <header className="flex items-start justify-between gap-4 border-b px-5 py-4">
            <div className="flex flex-col gap-1">
              <DialogPrimitive.Title className="text-lg font-semibold tracking-tight">
                {title}
              </DialogPrimitive.Title>
              {description && (
                <DialogPrimitive.Description className="text-sm text-muted-foreground">
                  {description}
                </DialogPrimitive.Description>
              )}
            </div>
            <DialogPrimitive.Close
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
            </DialogPrimitive.Close>
          </header>
          <div className="flex-1 overflow-y-auto px-5 py-5">{children}</div>
        </DialogPrimitive.Popup>
      </DialogPrimitive.Portal>
    </DialogPrimitive.Root>
  )
}
