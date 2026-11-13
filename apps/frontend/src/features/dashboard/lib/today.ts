// The summary covers the person's own day, so "today" runs from local
// midnight to the next one, sent to the backend as instants.
export function todayRange(now: Date = new Date()): { from: string; to: string } {
  const start = new Date(now.getFullYear(), now.getMonth(), now.getDate())
  const end = new Date(now.getFullYear(), now.getMonth(), now.getDate() + 1)
  return { from: start.toISOString(), to: end.toISOString() }
}
