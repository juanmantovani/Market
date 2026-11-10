import { Toast } from '@base-ui/react/toast'

export function useToast() {
  const manager = Toast.useToastManager()

  return {
    success: (title: string, description?: string) =>
      manager.add({ title, description, type: 'success' }),
    error: (title: string, description?: string) =>
      manager.add({ title, description, type: 'error', priority: 'high' }),
  }
}
