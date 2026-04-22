const BASE = import.meta.env.VITE_API_URL ?? '/api'

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const token = localStorage.getItem('seller_token')
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}), ...init?.headers },
  })
  if (res.status === 401) { localStorage.removeItem('seller_token'); window.location.href = '/login'; throw new Error('Unauthorized') }
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json() as Promise<T>
}

export const api = {
  get:    <T>(p: string)              => req<T>(p),
  post:   <T>(p: string, b: unknown) => req<T>(p, { method: 'POST',   body: JSON.stringify(b) }),
  put:    <T>(p: string, b: unknown) => req<T>(p, { method: 'PUT',    body: JSON.stringify(b) }),
  delete: <T>(p: string)             => req<T>(p, { method: 'DELETE' }),
}
