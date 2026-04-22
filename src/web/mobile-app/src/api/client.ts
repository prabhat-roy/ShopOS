import * as SecureStore from 'expo-secure-store'

const BASE = process.env.EXPO_PUBLIC_API_URL ?? 'http://mobile-bff:8082'

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const token = await SecureStore.getItemAsync('auth_token')
  const res = await fetch(`${BASE}/api${path}`, {
    ...init,
    headers: {
      'Content-Type': 'application/json',
      ...(token ? { Authorization: `Bearer ${token}` } : {}),
      ...init?.headers,
    },
  })
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json() as Promise<T>
}

export const api = {
  get:  <T>(p: string)              => apiFetch<T>(p),
  post: <T>(p: string, b: unknown) => apiFetch<T>(p, { method: 'POST', body: JSON.stringify(b) }),
}
