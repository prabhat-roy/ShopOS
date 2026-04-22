import { create } from 'zustand'
import { persist } from 'zustand/middleware'

interface AuthStore { token: string | null; email: string | null; setAuth: (t: string, e: string) => void; logout: () => void }
export const useAuth = create<AuthStore>()(persist(
  (set) => ({
    token: null, email: null,
    setAuth: (token, email) => { localStorage.setItem('dev_token', token); set({ token, email }) },
    logout: () => { localStorage.removeItem('dev_token'); set({ token: null, email: null }) },
  }),
  { name: 'dev-auth' }
))
