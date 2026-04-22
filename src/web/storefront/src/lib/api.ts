const BASE = process.env.NEXT_PUBLIC_API_URL ?? ''

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const res = await fetch(`${BASE}/api${path}`, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...init?.headers },
  })
  if (!res.ok) {
    const err = await res.text()
    throw new Error(err || `HTTP ${res.status}`)
  }
  return res.json() as Promise<T>
}

export const api = {
  get:    <T>(path: string)              => req<T>(path),
  post:   <T>(path: string, body: unknown) => req<T>(path, { method: 'POST',   body: JSON.stringify(body) }),
  put:    <T>(path: string, body: unknown) => req<T>(path, { method: 'PUT',    body: JSON.stringify(body) }),
  delete: <T>(path: string)              => req<T>(path, { method: 'DELETE' }),
}
