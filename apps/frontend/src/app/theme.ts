export type Theme = 'light' | 'dark' | 'system'

const THEMES: readonly Theme[] = ['light', 'dark', 'system']
const STORAGE_KEY = 'market.theme'
const DARK_QUERY = '(prefers-color-scheme: dark)'

export const THEME_LABELS: Record<Theme, string> = {
  light: 'Claro',
  dark: 'Oscuro',
  system: 'Automático',
}

function isTheme(value: unknown): value is Theme {
  return THEMES.includes(value as Theme)
}

// Storage can be unavailable (private mode, blocked site data): the theme is
// a convenience, so any failure falls back to following the system.
export function readStoredTheme(): Theme {
  try {
    const stored = window.localStorage.getItem(STORAGE_KEY)
    return isTheme(stored) ? stored : 'system'
  } catch {
    return 'system'
  }
}

export function storeTheme(theme: Theme) {
  try {
    window.localStorage.setItem(STORAGE_KEY, theme)
  } catch {
    // The choice still applies for this visit.
  }
}

export function systemPrefersDark(): boolean {
  return typeof window.matchMedia === 'function' && window.matchMedia(DARK_QUERY).matches
}

export function watchSystemTheme(onChange: () => void): () => void {
  if (typeof window.matchMedia !== 'function') {
    return () => {}
  }

  const media = window.matchMedia(DARK_QUERY)
  media.addEventListener('change', onChange)
  return () => media.removeEventListener('change', onChange)
}

export function resolveTheme(theme: Theme, prefersDark: boolean): 'light' | 'dark' {
  if (theme === 'system') {
    return prefersDark ? 'dark' : 'light'
  }
  return theme
}

export function applyTheme(theme: Theme) {
  const dark = resolveTheme(theme, systemPrefersDark()) === 'dark'
  document.documentElement.classList.toggle('dark', dark)
}

export function nextTheme(theme: Theme): Theme {
  return THEMES[(THEMES.indexOf(theme) + 1) % THEMES.length]
}
