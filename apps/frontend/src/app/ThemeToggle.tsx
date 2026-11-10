import { Monitor, Moon, Sun } from 'lucide-react'
import { Button } from '../shared/ui/button'
import { nextTheme, THEME_LABELS } from './theme'
import { useTheme } from './use-theme'

const THEME_ICONS = { light: Sun, dark: Moon, system: Monitor }

export default function ThemeToggle() {
  const { theme, setTheme } = useTheme()
  const Icon = THEME_ICONS[theme]
  const next = nextTheme(theme)
  const label = `Tema ${THEME_LABELS[theme].toLowerCase()}. Cambiar a ${THEME_LABELS[next].toLowerCase()}`

  return (
    <Button
      type="button"
      variant="ghost"
      size="icon-lg"
      className="size-11 text-muted-foreground md:size-9"
      aria-label={label}
      title={label}
      onClick={() => setTheme(next)}
    >
      <Icon className="size-5" aria-hidden />
    </Button>
  )
}
