const STORAGE_KEY = 'market.sidebar-collapsed'

export function readSidebarCollapsed(): boolean {
  try {
    return window.localStorage.getItem(STORAGE_KEY) === 'true'
  } catch {
    return false
  }
}

export function storeSidebarCollapsed(collapsed: boolean) {
  try {
    window.localStorage.setItem(STORAGE_KEY, String(collapsed))
  } catch {
    // The choice still applies for this visit.
  }
}
