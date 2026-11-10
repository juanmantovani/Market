import { useState, type ReactNode } from 'react'
import { QueryClientProvider, type QueryClient } from '@tanstack/react-query'
import { createQueryClient } from './query-client'
import ThemeProvider from './ThemeProvider'

type ProvidersProps = {
  children: ReactNode
  queryClient?: QueryClient
}

export default function Providers({ children, queryClient }: ProvidersProps) {
  const [client] = useState(() => queryClient ?? createQueryClient())

  return (
    <QueryClientProvider client={client}>
      <ThemeProvider>{children}</ThemeProvider>
    </QueryClientProvider>
  )
}
