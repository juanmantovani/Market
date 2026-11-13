import { createContext, useContext, useEffect, useState } from 'react'

type FormDialogControls = {
  // Closes the dialog, asking first when the form has unsaved changes.
  requestClose: () => void
  setHasChanges: (hasChanges: boolean) => void
}

export const FormDialogContext = createContext<FormDialogControls | null>(null)

export function useFormDialog() {
  const controls = useContext(FormDialogContext)
  if (!controls) throw new Error('useFormDialog must be used inside a FormDialog')
  return controls
}

// Tells the surrounding dialog whether `values` differ from what the form
// started with, so closing it can ask before discarding them.
export function useUnsavedChanges(values: unknown) {
  const { setHasChanges } = useFormDialog()
  const [initial] = useState(() => JSON.stringify(values))
  const hasChanges = JSON.stringify(values) !== initial

  useEffect(() => {
    setHasChanges(hasChanges)
    return () => setHasChanges(false)
  }, [hasChanges, setHasChanges])
}
