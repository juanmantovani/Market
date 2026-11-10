import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from 'react-router'
import '@fontsource-variable/inter'
import Providers from './app/Providers'
import { router } from './app/router'
import { applyTheme, readStoredTheme } from './app/theme'
import './index.css'

// Applied before the first render so a dark theme doesn't flash light.
applyTheme(readStoredTheme())

createRoot(document.getElementById('root')!).render(
  <StrictMode>
    <Providers>
      <RouterProvider router={router} />
    </Providers>
  </StrictMode>,
)
