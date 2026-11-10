import { Toast } from '@base-ui/react/toast'
import { CircleAlert, CircleCheck, X } from 'lucide-react'
import type { ReactNode } from 'react'

import { cn } from '@/shared/lib/utils'

export function ToastProvider({ children }: { children: ReactNode }) {
  return (
    <Toast.Provider timeout={4000}>
      {children}
      <Toast.Portal>
        <Toast.Viewport className="fixed right-4 bottom-20 left-4 z-[60] flex flex-col gap-2 outline-none md:bottom-6 md:left-auto md:w-96">
          <ToastList />
        </Toast.Viewport>
      </Toast.Portal>
    </Toast.Provider>
  )
}

function ToastList() {
  const { toasts } = Toast.useToastManager()

  return toasts.map((toast) => {
    const isError = toast.type === 'error'
    const Icon = isError ? CircleAlert : CircleCheck

    return (
      <Toast.Root
        key={toast.id}
        toast={toast}
        className="flex items-start gap-3 rounded-xl border bg-popover p-4 text-popover-foreground shadow-lg transition-all duration-200 data-ending-style:translate-y-2 data-ending-style:opacity-0 data-starting-style:translate-y-2 data-starting-style:opacity-0"
      >
        <Icon
          className={cn('mt-0.5 size-5 shrink-0', isError ? 'text-destructive' : 'text-success')}
          aria-hidden
        />
        <div className="flex min-w-0 flex-1 flex-col gap-0.5">
          <Toast.Title className="text-sm font-medium">{toast.title}</Toast.Title>
          {toast.description && (
            <Toast.Description className="text-sm text-muted-foreground">
              {toast.description}
            </Toast.Description>
          )}
        </div>
        <Toast.Close
          aria-label="Cerrar aviso"
          className="-m-1 flex size-8 shrink-0 items-center justify-center rounded-md text-muted-foreground hover:text-foreground"
        >
          <X className="size-4" aria-hidden />
        </Toast.Close>
      </Toast.Root>
    )
  })
}
