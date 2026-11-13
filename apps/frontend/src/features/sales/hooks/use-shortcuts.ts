import { useEffect, useRef } from 'react'

// Function keys that work anywhere on the counter screen, even while
// typing in a field: F2 to search, and so on. Each handler replaces what the
// browser would do with the key.
export function useShortcuts(handlers: Partial<Record<string, () => void>>) {
  const latest = useRef(handlers)

  useEffect(() => {
    latest.current = handlers
  })

  useEffect(() => {
    function onKeyDown(event: KeyboardEvent) {
      if (event.altKey || event.ctrlKey || event.metaKey) return
      const handler = latest.current[event.key]
      if (!handler) return
      event.preventDefault()
      handler()
    }

    window.addEventListener('keydown', onKeyDown)
    return () => window.removeEventListener('keydown', onKeyDown)
  }, [])
}
