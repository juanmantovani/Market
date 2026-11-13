import { describe, expect, it } from 'vitest'
import { buildContentSecurityPolicy, contentSecurityPolicyPlugin } from './content-security-policy'

describe('buildContentSecurityPolicy', () => {
  it('limits connections to the same origin by default', () => {
    const csp = buildContentSecurityPolicy({ isDev: false })

    expect(csp).toContain("connect-src 'self'")
    expect(csp).not.toContain('http')
  })

  it('adds only the exact origin of a separate API', () => {
    const csp = buildContentSecurityPolicy({ apiUrl: 'https://api.example.com/v1/', isDev: false })

    expect(csp).toContain("connect-src 'self' https://api.example.com;")
    expect(csp).not.toContain('*')
  })

  it('allows inline styles only in dev, for Vite and Tailwind CSS injection', () => {
    expect(buildContentSecurityPolicy({ isDev: true })).toContain("style-src 'self' 'unsafe-inline'")

    const prod = buildContentSecurityPolicy({ isDev: false })
    expect(prod).toContain("style-src 'self'")
    expect(prod).not.toContain('unsafe-inline')
  })

  it('never allows unsafe-eval or inline scripts', () => {
    for (const isDev of [true, false]) {
      const csp = buildContentSecurityPolicy({ isDev })

      expect(csp).toContain("script-src 'self'")
      expect(csp).not.toContain('unsafe-eval')
    }
  })
})

describe('contentSecurityPolicyPlugin', () => {
  it('injects the policy as a meta tag in the head', () => {
    const plugin = contentSecurityPolicyPlugin({ isDev: false })
    const transform = plugin.transformIndexHtml as () => Array<{
      tag: string
      attrs: Record<string, string>
      injectTo: string
    }>

    const [tag] = transform()

    expect(tag.tag).toBe('meta')
    expect(tag.injectTo).toBe('head')
    expect(tag.attrs['http-equiv']).toBe('Content-Security-Policy')
    expect(tag.attrs.content).toBe(buildContentSecurityPolicy({ isDev: false }))
  })
})
