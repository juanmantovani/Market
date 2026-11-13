export type ShortcutGroup = {
  title: string
  shortcuts: readonly { keys: readonly string[]; description: string }[]
}

export const SHORTCUT_GROUPS: readonly ShortcutGroup[] = [
  {
    title: 'Punto de venta',
    shortcuts: [
      { keys: ['F2'], description: 'Buscar o escanear un artículo' },
      { keys: ['Enter'], description: 'Agregar el artículo del código escrito o escaneado' },
      { keys: ['↓'], description: 'Pasar de la búsqueda al catálogo' },
      { keys: ['F4'], description: 'Elegir el vendedor' },
      { keys: ['F6'], description: 'Elegir el cliente' },
      { keys: ['F9'], description: 'Cobrar' },
      { keys: ['Esc'], description: 'Cerrar el cuadro abierto o limpiar la búsqueda' },
    ],
  },
  {
    title: 'En cualquier pantalla',
    shortcuts: [{ keys: ['?'], description: 'Ver esta ayuda' }],
  },
]

const TEXT_ENTRY = 'input, textarea, select, [contenteditable=""], [contenteditable="true"]'

// `?` is a character people type, so it only opens the help when nothing is
// being typed and no other dialog is in the way.
export function isHelpKey(event: KeyboardEvent) {
  if (event.key !== '?' || event.ctrlKey || event.metaKey || event.altKey) return false
  if (event.target instanceof Element && event.target.closest(TEXT_ENTRY)) return false
  return document.querySelector('[role="dialog"], [role="alertdialog"]') === null
}
