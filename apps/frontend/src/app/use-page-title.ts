import { useEffect } from 'react'
import { useMatches } from 'react-router'

const APP_NAME = 'Market'

export type RouteHandle = { title: string }

// Keeps the browser tab title in step with the screen (WCAG 2.4.2), from the
// `handle.title` of the deepest matching route.
export function usePageTitle() {
  const matches = useMatches()
  const title = (matches.at(-1)?.handle as RouteHandle | undefined)?.title

  useEffect(() => {
    document.title = title ? `${title} · ${APP_NAME}` : APP_NAME
  }, [title])
}
