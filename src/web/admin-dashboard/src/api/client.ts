const BASE = import.meta.env.VITE_API_URL ?? '/api'

async function req<T>(path: string, init?: RequestInit): Promise<T> {
  const token = localStorage.getItem('admin_token')
  const res = await fetch(`${BASE}${path}`, {
    ...init,
    headers: { 'Content-Type': 'application/json', ...(token ? { Authorization: `Bearer ${token}` } : {}), ...init?.headers },
  })
  if (res.status === 401) { localStorage.removeItem('admin_token'); window.location.href = '/login'; throw new Error('Unauthorized') }
  if (!res.ok) throw new Error(`HTTP ${res.status}`)
  return res.json() as Promise<T>
}

export const api = {
  get:    <T>(p: string) => req<T>(p),
  post:   <T>(p: string, b: unknown) => req<T>(p, { method: 'POST', body: JSON.stringify(b) }),
  put:    <T>(p: string, b: unknown) => req<T>(p, { method: 'PUT',  body: JSON.stringify(b) }),
  delete: <T>(p: string) => req<T>(p, { method: 'DELETE' }),
}

export const ordersApi   = { list: (p = 1) => api.get<import('../types').Order[]>(`/admin/orders?page=${p}`), update: (id: string, status: string) => api.put(`/admin/orders/${id}`, { status }) }
export const productsApi = { list: (p = 1) => api.get<import('../types').Product[]>(`/admin/products?page=${p}`), create: (b: unknown) => api.post('/admin/products', b), delete: (id: string) => api.delete(`/admin/products/${id}`) }
export const usersApi    = { list: (p = 1) => api.get<import('../types').User[]>(`/admin/users?page=${p}`) }
export const statsApi    = { get: () => api.get<import('../types').DashboardStats>('/admin/stats') }
