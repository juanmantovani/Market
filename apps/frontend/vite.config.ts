import path from 'node:path'
import react from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'
import { loadEnv } from 'vite'
import { defineConfig } from 'vitest/config'
import { contentSecurityPolicyPlugin } from './vite-plugins/content-security-policy.ts'

const DEFAULT_PROXY_TARGET = 'http://localhost:8080'

export default defineConfig(({ mode, command }) => {
  const env = loadEnv(mode, process.cwd(), '')
  const proxyTarget = env.API_PROXY_TARGET || DEFAULT_PROXY_TARGET

  return {
    plugins: [
      react(),
      tailwindcss(),
      contentSecurityPolicyPlugin({
        apiUrl: env.VITE_API_URL || undefined,
        isDev: command === 'serve',
      }),
    ],
    resolve: {
      alias: {
        '@': path.resolve(import.meta.dirname, './src'),
      },
    },
    build: {
      rolldownOptions: {
        output: {
          codeSplitting: {
            // React and the router change far less often than the app, so
            // their own chunk stays cached across deploys.
            groups: [{ name: 'react', test: /[\\/]node_modules[\\/](react|react-dom|react-router|scheduler)[\\/]/ }],
          },
        },
      },
    },
    server: {
      proxy: {
        '/api': proxyTarget,
      },
      watch: env.WATCH_POLLING === 'true' ? { usePolling: true, interval: 200 } : undefined,
    },
    test: {
      environment: 'jsdom',
      setupFiles: './src/test/setup.ts',
      // Page tests type into real forms; with coverage on and every worker
      // busy they take several seconds each, beyond Vitest's 5 s default.
      testTimeout: 15_000,
      coverage: {
        provider: 'v8',
        reporter: ['text', 'html', 'lcov'],
        reportsDirectory: './coverage',
        include: ['src/features/**/*.{ts,tsx}', 'src/shared/api/**/*.ts', 'src/shared/lib/**/*.ts'],
        exclude: ['src/**/*.test.{ts,tsx}'],
        thresholds: {
          perFile: true,
          branches: 75,
          functions: 80,
          lines: 80,
          statements: 80,
        },
      },
    },
  }
})
