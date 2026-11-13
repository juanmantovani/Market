import { useState } from 'react'

// Dialogs here open from plain buttons through controlled state, so Base UI has
// no trigger to hand focus back to, and an `autoFocus` field inside the dialog
// takes focus before Base UI records where it came from. The element is read
// while rendering, before the dialog mounts and moves focus.
export function useReturnFocus(open: boolean) {
  const [opener, setOpener] = useState<HTMLElement | null>(null)
  const [wasOpen, setWasOpen] = useState(open)

  if (open !== wasOpen) {
    setWasOpen(open)
    if (open) {
      const active = document.activeElement
      setOpener(active instanceof HTMLElement && active !== document.body ? active : null)
    }
  }

  // `true` keeps Base UI's own choice when the opener is gone (for example a
  // row that was just deleted).
  return () => (opener?.isConnected ? opener : true)
}
