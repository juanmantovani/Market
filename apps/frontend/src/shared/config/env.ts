// Empty means the API shares the frontend's origin (Vite proxy in
// development, reverse proxy or the Go binary in production).
export const apiBaseUrl = (import.meta.env.VITE_API_URL ?? '').replace(/\/+$/, '')
