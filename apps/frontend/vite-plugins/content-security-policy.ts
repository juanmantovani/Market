import type { Plugin } from 'vite'

export interface CspOptions {
  // Only set when the API lives on another origin; same-origin requests are
  // already covered by 'self'.
  apiUrl?: string
  isDev: boolean
}

export function buildContentSecurityPolicy({ apiUrl, isDev }: CspOptions): string {
  const styleSrc = isDev ? `'self' 'unsafe-inline'` : `'self'`
  const connectSrc = apiUrl ? `'self' ${new URL(apiUrl).origin}` : `'self'`

  return [
    `default-src 'self'`,
    `script-src 'self'`,
    `style-src ${styleSrc}`,
    `img-src 'self' data: blob:`,
    `font-src 'self'`,
    `connect-src ${connectSrc}`,
    `base-uri 'self'`,
    `form-action 'self'`,
    `object-src 'none'`,
  ].join('; ')
}

export function contentSecurityPolicyPlugin(options: CspOptions): Plugin {
  const content = buildContentSecurityPolicy(options)

  return {
    name: 'content-security-policy',
    transformIndexHtml() {
      return [
        {
          tag: 'meta',
          attrs: { 'http-equiv': 'Content-Security-Policy', content },
          injectTo: 'head',
        },
      ]
    },
  }
}
