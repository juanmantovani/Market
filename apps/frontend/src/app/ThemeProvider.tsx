import { useCallback, useEffect, useMemo, useState, type ReactNode } from 'react'
import { applyTheme, readStoredTheme, storeTheme, watchSystemTheme, type Theme } from './theme'
import { ThemeContext } from './theme-context'

export default function ThemeProvider({ children }: { children: ReactNode }) {
  const [theme, setThemeState] = useState<Theme>(readStoredTheme)

  useEffect(() => {
    applyTheme(theme)
    if (theme !== 'system') {
      return
    }
    return watchSystemTheme(() => applyTheme('system'))
  }, [theme])

  const setTheme = useCallback((next: Theme) => {
    storeTheme(next)
    setThemeState(next)
  }, [])

  const value = useMemo(() => ({ theme, setTheme }), [theme, setTheme])

  return <ThemeContext value={value}>{children}</ThemeContext>
}
