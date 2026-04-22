import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthStore {
  token: string | null
  email: string | null
  setAuth: (token: string, email: string) => void
  logout: () => void
}

export const useAuth = create<AuthStore>()(
  persist(
    (set) => ({
      token: null, email: null,
      setAuth: (token, email) => { localStorage.setItem('admin_token', token); set({ token, email }) },
      logout: () => { localStorage.removeItem('admin_token'); set({ token: null, email: null }) },
    }),
    { name: 'admin-auth' }
  )
)
