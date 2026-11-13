const dateTime = new Intl.DateTimeFormat('es-AR', { dateStyle: 'short', timeStyle: 'short' })

// The backend sends instants in UTC (RFC 3339); people read them in their
// own time zone.
export function formatDateTime(iso: string): string {
  return dateTime.format(new Date(iso))
}
