const BASE = import.meta.env.VITE_API_URL ?? '/api'

let refreshPromise: Promise<string> | null = null

async function refreshToken(): Promise<string> {
  if (refreshPromise) return refreshPromise
  refreshPromise = fetch(`${BASE}/auth/refresh`, {
    method: 'POST',
    credentials: 'include',
  })
    .then(r => { if (!r.ok) throw new Error('Refresh failed'); return r.json() })
    .then(d => { localStorage.setItem('token', d.token); return d.token as string })
    .catch(e => { localStorage.removeItem('token'); window.location.href = '/login'; throw e })
    .finally(() => { refreshPromise = null })
  return refreshPromise
}

async function req<T>(path: string, init: RequestInit = {}, retry = true): Promise<T> {
  const token = localStorage.getItem('token')
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init.headers,
    },
  })

  if (res.status === 401 && retry) {
    try {
      const newToken = await refreshToken()
      return req<T>(path, {
        ...init,
        headers: { ...init.headers, Authorization: `Bearer ${newToken}` },
      }, false)
    } catch {
      throw new Error('Session expired')
    }
  }

  if (!res.ok) {
    const text = await res.text()
    throw new Error(text || `HTTP ${res.status}: ${path}`)
  }
  return res.json() as Promise<T>
}

export const api = {
  get:    <T>(p: string)              => req<T>(p),
  post:   <T>(p: string, b: unknown) => req<T>(p, { method: 'POST',   body: JSON.stringify(b) }),
  put:    <T>(p: string, b: unknown) => req<T>(p, { method: 'PUT',    body: JSON.stringify(b) }),
  patch:  <T>(p: string, b: unknown) => req<T>(p, { method: 'PATCH',  body: JSON.stringify(b) }),
  delete: <T>(p: string)             => req<T>(p, { method: 'DELETE' }),
}
