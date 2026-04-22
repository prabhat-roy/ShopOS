import { create } from 'zustand'
import * as SecureStore from 'expo-secure-store'

interface AuthStore {
  token: string | null
  user: { id: string; email: string; firstName: string } | null
  setAuth: (token: string, user: AuthStore['user']) => void
  logout: () => void
}

export const useAuth = create<AuthStore>((set) => ({
  token: null,
  user: null,
  setAuth: async (token, user) => {
    await SecureStore.setItemAsync('auth_token', token)
    set({ token, user })
  },
  logout: async () => {
    await SecureStore.deleteItemAsync('auth_token')
    set({ token: null, user: null })
  },
}))
