import { Toast } from '@base-ui/react/toast'

export function useToast() {
  const manager = Toast.useToastManager()

  return {
    success: (title: string, description?: string) =>
      manager.add({ title, description, type: 'success' }),
    error: (title: string, description?: string) =>
      manager.add({ title, description, type: 'error', priority: 'high' }),
    // A notice with "Deshacer", for actions that can be taken back instead
    // of asking for confirmation first.
    undoable: (title: string, onUndo: () => void) => {
      const id = manager.add({
        title,
        type: 'success',
        timeout: 6000,
        actionProps: {
          children: 'Deshacer',
          onClick: () => {
            onUndo()
            manager.close(id)
          },
        },
      })
      return id
    },
  }
}
