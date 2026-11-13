import { afterEach, describe, expect, it, vi } from 'vitest'
import {
  applyTheme,
  nextTheme,
  readStoredTheme,
  resolveTheme,
  storeTheme,
  watchSystemTheme,
} from './theme'

function stubMatchMedia(matches: boolean) {
  const listeners = new Set<() => void>()
  const media = {
    matches,
    addEventListener: vi.fn((_: string, listener: () => void) => listeners.add(listener)),
    removeEventListener: vi.fn((_: string, listener: () => void) => listeners.delete(listener)),
  }
  vi.stubGlobal('matchMedia', vi.fn(() => media))
  return { media, listeners }
}

afterEach(() => {
  window.localStorage.clear()
  document.documentElement.classList.remove('dark')
  vi.unstubAllGlobals()
})

describe('stored theme', () => {
  it('follows the system when nothing valid is stored', () => {
    expect(readStoredTheme()).toBe('system')

    window.localStorage.setItem('market.theme', 'neon')
    expect(readStoredTheme()).toBe('system')
  })

  it('round-trips a stored choice', () => {
    storeTheme('dark')
    expect(readStoredTheme()).toBe('dark')
  })

  it('survives unavailable storage', () => {
    vi.spyOn(Storage.prototype, 'getItem').mockImplementation(() => {
      throw new Error('blocked')
    })
    vi.spyOn(Storage.prototype, 'setItem').mockImplementation(() => {
      throw new Error('blocked')
    })

    expect(() => storeTheme('light')).not.toThrow()
    expect(readStoredTheme()).toBe('system')
  })
})

describe('resolveTheme', () => {
  it('uses the system preference only for the automatic theme', () => {
    expect(resolveTheme('system', true)).toBe('dark')
    expect(resolveTheme('system', false)).toBe('light')
    expect(resolveTheme('light', true)).toBe('light')
    expect(resolveTheme('dark', false)).toBe('dark')
  })
})

describe('applyTheme', () => {
  it('marks the document dark when the system prefers it', () => {
    stubMatchMedia(true)

    applyTheme('system')
    expect(document.documentElement).toHaveClass('dark')

    applyTheme('light')
    expect(document.documentElement).not.toHaveClass('dark')
  })
})

describe('watchSystemTheme', () => {
  it('subscribes to system changes and unsubscribes on cleanup', () => {
    const { listeners } = stubMatchMedia(false)
    const onChange = vi.fn()

    const stop = watchSystemTheme(onChange)
    listeners.forEach((listener) => listener())
    expect(onChange).toHaveBeenCalledTimes(1)

    stop()
    expect(listeners.size).toBe(0)
  })

  it('does nothing without matchMedia support', () => {
    expect(() => watchSystemTheme(vi.fn())()).not.toThrow()
  })
})

describe('nextTheme', () => {
  it('cycles light, dark and automatic', () => {
    expect(nextTheme('light')).toBe('dark')
    expect(nextTheme('dark')).toBe('system')
    expect(nextTheme('system')).toBe('light')
  })
})
