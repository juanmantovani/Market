import { useEffect, useRef, type RefObject } from 'react'
import { useLocation } from 'react-router'

const WAIT_FOR_PAGE_MS = 3000

// A screen change in a single-page app leaves focus on the link that was
// clicked, so a screen reader never announces the new screen. After each
// navigation the focus goes to the page's h1, as soon as the lazy page
// renders it, unless the page already took focus itself (the point of sale
// focuses its search).
export function useFocusHeadingOnNavigation(main: RefObject<HTMLElement | null>) {
  const { pathname } = useLocation()
  const isFirstRender = useRef(true)

  useEffect(() => {
    if (isFirstRender.current) {
      isFirstRender.current = false
      return
    }
    const container = main.current
    if (!container) return

    function focusHeading() {
      const heading = container?.querySelector('h1')
      if (!heading) return false
      if (!container?.contains(document.activeElement)) {
        heading.tabIndex = -1
        heading.focus()
      }
      return true
    }

    if (focusHeading()) return
    const observer = new MutationObserver(() => {
      if (focusHeading()) observer.disconnect()
    })
    observer.observe(container, { childList: true, subtree: true })
    const timer = window.setTimeout(() => observer.disconnect(), WAIT_FOR_PAGE_MS)
    return () => {
      observer.disconnect()
      window.clearTimeout(timer)
    }
  }, [pathname, main])
}
